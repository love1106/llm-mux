package provider

import (
	"context"
	"strings"
	"time"

	log "github.com/nghyane/llm-mux/internal/logging"
	"golang.org/x/sync/semaphore"
)

const (
	// maxConcurrentRefreshes limits the number of concurrent refresh goroutines
	// to prevent goroutine explosion under high load with many auths.
	maxConcurrentRefreshes = 10

	refreshCheckInterval  = 5 * time.Second
	refreshPendingBackoff = time.Minute
	refreshFailureBackoff = 5 * time.Minute
)

// newRefreshSemaphore creates a weighted semaphore for bounding concurrent refresh operations.
func newRefreshSemaphore() *semaphore.Weighted {
	return semaphore.NewWeighted(maxConcurrentRefreshes)
}

// StartAutoRefresh launches a background loop that evaluates auth freshness
// every few seconds and triggers refresh operations when required.
// Only one loop is kept alive; starting a new one cancels the previous run.
func (m *Manager) StartAutoRefresh(parent context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = refreshCheckInterval
	}
	if m.refreshCancel != nil {
		m.refreshCancel()
		m.refreshCancel = nil
	}
	ctx, cancel := context.WithCancel(parent)
	m.refreshCancel = cancel
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Cleanup provider stats every hour to prevent memory leak
		cleanupTicker := time.NewTicker(1 * time.Hour)
		defer cleanupTicker.Stop()

		m.checkRefreshes(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.checkRefreshes(ctx)
			case <-cleanupTicker.C:
				// Remove provider stats older than 24 hours
				removed := m.CleanupProviderStats(24 * time.Hour)
				if removed > 0 {
					log.Debugf("Cleaned up %d stale provider stats entries", removed)
				}
			}
		}
	}()
}

// StopAutoRefresh cancels the background refresh loop, if running.
func (m *Manager) StopAutoRefresh() {
	if m.refreshCancel != nil {
		m.refreshCancel()
		m.refreshCancel = nil
	}
}

// checkRefreshes evaluates all registered auths and triggers refreshes as needed.
// Uses a semaphore to bound the number of concurrent refresh goroutines.
func (m *Manager) checkRefreshes(ctx context.Context) {
	now := time.Now()
	var snapshot []*Auth
	if m.registry != nil {
		snapshot = m.registry.List()
	} else {
		snapshot = m.snapshotAuths()
	}
	log.Debugf("[refresh] checking %d auths", len(snapshot))
	for _, a := range snapshot {
		typ, _ := a.AccountInfo()
		if typ != "api_key" {
			shouldRefresh := m.shouldRefresh(a, now)
			expiry, hasExpiry := a.ExpirationTime()
			log.Debugf("[refresh] auth=%s provider=%s type=%s shouldRefresh=%v hasExpiry=%v expiry=%s",
				a.ID, a.Provider, typ, shouldRefresh, hasExpiry, expiry.Format(time.RFC3339))
			if !shouldRefresh {
				continue
			}
			log.Infof("[refresh] triggering refresh for provider=%s auth=%s type=%s", a.Provider, a.ID, typ)

		if exec := m.executorFor(a.Provider); exec == nil {
			log.Warnf("[refresh] no executor found for provider=%s auth=%s", a.Provider, a.ID)
			continue
		}
		if m.refreshSem != nil {
			if !m.refreshSem.TryAcquire(1) {
				log.Debugf("[refresh] skipped auth=%s: semaphore full", a.ID)
				continue
			}
			if !m.markRefreshPending(a.ID, now) {
				m.refreshSem.Release(1)
				continue
			}
			go func(authID string) {
				defer m.refreshSem.Release(1)
				m.refreshAuth(ctx, authID)
			}(a.ID)
		} else {
			if !m.markRefreshPending(a.ID, now) {
				continue
			}
			go m.refreshAuth(ctx, a.ID)
		}
		}
	}
}

// snapshotAuths creates a copy of all currently registered auths.
func (m *Manager) snapshotAuths() []*Auth {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Auth, 0, len(m.auths))
	for _, a := range m.auths {
		out = append(out, a.Clone())
	}
	return out
}

// shouldRefresh determines if an auth needs refresh based on expiration and refresh rules.
func (m *Manager) shouldRefresh(a *Auth, now time.Time) bool {
	if a == nil {
		return false
	}
	if a.Disabled {
		log.Debugf("[shouldRefresh] auth=%s skip: disabled", a.ID)
		return false
	}
	if !a.NextRefreshAfter.IsZero() && now.Before(a.NextRefreshAfter) {
		log.Debugf("[shouldRefresh] auth=%s skip: waiting until %s", a.ID, a.NextRefreshAfter.Format(time.RFC3339))
		return false
	}
	if evaluator, ok := a.Runtime.(RefreshEvaluator); ok && evaluator != nil {
		result := evaluator.ShouldRefresh(now, a)
		log.Debugf("[shouldRefresh] auth=%s evaluator returned %v", a.ID, result)
		return result
	}

	lastRefresh := a.LastRefreshedAt
	if lastRefresh.IsZero() {
		if ts, ok := authLastRefreshTimestamp(a); ok {
			lastRefresh = ts
		}
	}

	expiry, hasExpiry := a.ExpirationTime()
	log.Debugf("[shouldRefresh] auth=%s lastRefresh=%s expiry=%s hasExpiry=%v",
		a.ID, lastRefresh.Format(time.RFC3339), expiry.Format(time.RFC3339), hasExpiry)

	if interval := authPreferredInterval(a); interval > 0 {
		if hasExpiry && !expiry.IsZero() {
			if !expiry.After(now) {
				log.Debugf("[shouldRefresh] auth=%s refresh: expired (interval mode)", a.ID)
				return true
			}
			if expiry.Sub(now) <= interval {
				log.Debugf("[shouldRefresh] auth=%s refresh: within interval %v of expiry", a.ID, interval)
				return true
			}
		}
		if lastRefresh.IsZero() {
			log.Debugf("[shouldRefresh] auth=%s refresh: no lastRefresh (interval mode)", a.ID)
			return true
		}
		shouldRefresh := now.Sub(lastRefresh) >= interval
		log.Debugf("[shouldRefresh] auth=%s interval check: sinceLastRefresh=%v interval=%v result=%v",
			a.ID, now.Sub(lastRefresh), interval, shouldRefresh)
		return shouldRefresh
	}

	provider := strings.ToLower(a.Provider)
	lead := ProviderRefreshLead(provider, a.Runtime)
	if lead == nil {
		log.Debugf("[shouldRefresh] auth=%s skip: no refresh lead for provider %s", a.ID, provider)
		return false
	}
	log.Infof("[shouldRefresh] auth=%s provider=%s lead=%v", a.ID, provider, *lead)
	if *lead <= 0 {
		if hasExpiry && !expiry.IsZero() {
			result := now.After(expiry)
			log.Debugf("[shouldRefresh] auth=%s lead<=0, expired check: %v", a.ID, result)
			return result
		}
		return false
	}
	if hasExpiry && !expiry.IsZero() {
		timeUntilExpiry := time.Until(expiry)
		shouldRefresh := timeUntilExpiry <= *lead
		log.Debugf("[shouldRefresh] auth=%s timeUntilExpiry=%v lead=%v shouldRefresh=%v",
			a.ID, timeUntilExpiry, *lead, shouldRefresh)
		return shouldRefresh
	}
	if !lastRefresh.IsZero() {
		sinceLastRefresh := now.Sub(lastRefresh)
		shouldRefresh := sinceLastRefresh >= *lead
		log.Debugf("[shouldRefresh] auth=%s sinceLastRefresh=%v lead=%v shouldRefresh=%v",
			a.ID, sinceLastRefresh, *lead, shouldRefresh)
		return shouldRefresh
	}
	log.Debugf("[shouldRefresh] auth=%s refresh: fallback true", a.ID)
	return true
}

// authPreferredInterval extracts the refresh interval from auth metadata and attributes.
func authPreferredInterval(a *Auth) time.Duration {
	if a == nil {
		return 0
	}
	if d := durationFromMetadata(a.Metadata, "refresh_interval_seconds", "refreshIntervalSeconds", "refresh_interval", "refreshInterval"); d > 0 {
		return d
	}
	if d := durationFromAttributes(a.Attributes, "refresh_interval_seconds", "refreshIntervalSeconds", "refresh_interval", "refreshInterval"); d > 0 {
		return d
	}
	return 0
}

// durationFromMetadata extracts a duration from metadata.
func durationFromMetadata(meta map[string]any, keys ...string) time.Duration {
	if len(meta) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, ok := meta[key]; ok {
			if dur := parseDurationValue(val); dur > 0 {
				return dur
			}
		}
	}
	return 0
}

// durationFromAttributes extracts a duration from string attributes.
func durationFromAttributes(attrs map[string]string, keys ...string) time.Duration {
	if len(attrs) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, ok := attrs[key]; ok {
			if dur := parseDurationString(val); dur > 0 {
				return dur
			}
		}
	}
	return 0
}

// authLastRefreshTimestamp looks up the last refresh time from auth metadata.
func authLastRefreshTimestamp(a *Auth) (time.Time, bool) {
	if a == nil {
		return time.Time{}, false
	}
	if a.Metadata != nil {
		if ts, ok := lookupMetadataTime(a.Metadata, "last_refresh", "lastRefresh", "last_refreshed_at", "lastRefreshedAt"); ok {
			return ts, true
		}
	}
	if a.Attributes != nil {
		for _, key := range []string{"last_refresh", "lastRefresh", "last_refreshed_at", "lastRefreshedAt"} {
			if val := strings.TrimSpace(a.Attributes[key]); val != "" {
				if ts, ok := parseTimeValue(val); ok {
					return ts, true
				}
			}
		}
	}
	return time.Time{}, false
}

// lookupMetadataTime looks up a time value from metadata.
func lookupMetadataTime(meta map[string]any, keys ...string) (time.Time, bool) {
	for _, key := range keys {
		if val, ok := meta[key]; ok {
			if ts, ok1 := parseTimeValue(val); ok1 {
				return ts, true
			}
		}
	}
	return time.Time{}, false
}

package provider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/nghyane/llm-mux/internal/json"
)

// RoundRobinSelector provides a simple provider scoped round-robin selection strategy.
// It uses a sharded StickyStore for 60-second sticky sessions to maintain conversation continuity.
type RoundRobinSelector struct {
	cursorMu sync.Mutex
	cursors  map[string]int
	sticky   *StickyStore
}

// Start launches the background cleanup goroutine for sticky sessions.
func (s *RoundRobinSelector) Start() {
	if s.sticky == nil {
		s.sticky = NewStickyStore()
	}
	s.sticky.Start()
}

// Stop gracefully shuts down the background cleanup goroutine.
func (s *RoundRobinSelector) Stop() {
	if s.sticky != nil {
		s.sticky.Stop()
	}
}

type blockReason int

// SelectorLifecycle is optionally implemented by Selectors needing lifecycle management.
type SelectorLifecycle interface {
	Start()
	Stop()
}

const (
	blockReasonNone blockReason = iota
	blockReasonCooldown
	blockReasonDisabled
	blockReasonOther
)

type modelCooldownError struct {
	model    string
	resetIn  time.Duration
	provider string
	reason   string // upstream error that triggered the cooldown
}

func newModelCooldownError(model, provider string, resetIn time.Duration) *modelCooldownError {
	if resetIn < 0 {
		resetIn = 0
	}
	return &modelCooldownError{
		model:    model,
		provider: provider,
		resetIn:  resetIn,
	}
}

func newModelCooldownErrorWithReason(model, provider string, resetIn time.Duration, reason string) *modelCooldownError {
	e := newModelCooldownError(model, provider, resetIn)
	e.reason = reason
	return e
}

func (e *modelCooldownError) Error() string {
	modelName := e.model
	if modelName == "" {
		modelName = "requested model"
	}
	message := fmt.Sprintf("All credentials for model %s are cooling down", modelName)
	if e.provider != "" {
		message = fmt.Sprintf("%s via provider %s", message, e.provider)
	}
	resetSeconds := int(math.Ceil(e.resetIn.Seconds()))
	if resetSeconds < 0 {
		resetSeconds = 0
	}
	displayDuration := e.resetIn
	if displayDuration > 0 && displayDuration < time.Second {
		displayDuration = time.Second
	} else {
		displayDuration = displayDuration.Round(time.Second)
	}
	errorBody := map[string]any{
		"code":          "model_cooldown",
		"message":       message,
		"model":         e.model,
		"reset_time":    displayDuration.String(),
		"reset_seconds": resetSeconds,
	}
	if e.provider != "" {
		errorBody["provider"] = e.provider
	}
	if e.reason != "" {
		errorBody["reason"] = e.reason
	}
	payload := map[string]any{"error": errorBody}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"error":{"code":"model_cooldown","message":"%s"}}`, message)
	}
	return string(data)
}

func (e *modelCooldownError) StatusCode() int {
	return http.StatusTooManyRequests
}

func (e *modelCooldownError) Headers() http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	resetSeconds := int(math.Ceil(e.resetIn.Seconds()))
	if resetSeconds < 0 {
		resetSeconds = 0
	}
	headers.Set("Retry-After", strconv.Itoa(resetSeconds))
	return headers
}

// Pick selects the next available auth for the provider in a round-robin manner.
func (s *RoundRobinSelector) Pick(ctx context.Context, provider, model string, opts Options, auths []*Auth) (*Auth, error) {
	_ = ctx
	if len(auths) == 0 {
		return nil, &Error{Code: "auth_not_found", Message: "no auth candidates"}
	}

	s.cursorMu.Lock()
	if s.cursors == nil {
		s.cursors = make(map[string]int)
	}
	if s.sticky == nil {
		s.sticky = NewStickyStore()
		s.sticky.Start()
	}
	s.cursorMu.Unlock()

	available := make([]*Auth, 0, len(auths))
	now := time.Now()
	cooldownCount := 0
	var earliest time.Time
	var lastReason string
	for i := 0; i < len(auths); i++ {
		candidate := auths[i]
		blocked, reason, next := isAuthBlockedForModel(candidate, model, now)
		if !blocked {
			available = append(available, candidate)
			continue
		}
		if reason == blockReasonCooldown || reason == blockReasonOther {
			if !next.IsZero() && (earliest.IsZero() || next.Before(earliest)) {
				earliest = next
				lastReason = extractAuthCooldownReason(candidate, model)
			}
		}
		if reason == blockReasonCooldown {
			cooldownCount++
		}
	}
	if len(available) == 0 {
		if !earliest.IsZero() {
			resetIn := earliest.Sub(now)
			if resetIn < 0 {
				resetIn = 0
			}
			return nil, newModelCooldownErrorWithReason(model, provider, resetIn, lastReason)
		}
		return nil, &Error{Code: "auth_unavailable", Message: "no auth available"}
	}
	if len(available) > 1 {
		sort.Slice(available, func(i, j int) bool { return available[i].ID < available[j].ID })
	}
	key := provider + ":" + model

	if !opts.ForceRotate {
		if authID, ok := s.sticky.Get(key); ok {
			for _, auth := range available {
				if auth.ID == authID {
					return auth, nil
				}
			}
		}
	}

	s.cursorMu.Lock()
	index := s.cursors[key]
	if index >= 1_000_000_000 || index < 0 {
		index = 0
	}
	s.cursors[key] = index + 1
	s.cursorMu.Unlock()

	selected := available[index%len(available)]
	s.sticky.Set(key, selected.ID)
	return selected, nil
}

func isAuthBlockedForModel(auth *Auth, model string, now time.Time) (bool, blockReason, time.Time) {
	if auth == nil {
		return true, blockReasonOther, time.Time{}
	}
	if auth.Disabled || auth.Status == StatusDisabled {
		return true, blockReasonDisabled, time.Time{}
	}
	if model != "" {
		if len(auth.ModelStates) > 0 {
			// First check the specific model state
			if state, ok := auth.ModelStates[model]; ok && state != nil {
				if state.Status == StatusDisabled {
					return true, blockReasonDisabled, time.Time{}
				}
				if state.Unavailable {
					if state.NextRetryAfter.IsZero() {
						// Block auth with unknown retry time - prevents routing to failed auth
						// without cooldown info. Manager should set default cooldown to enable recovery.
						return true, blockReasonOther, time.Time{}
					}
					if state.NextRetryAfter.After(now) {
						next := state.NextRetryAfter
						if !state.Quota.NextRecoverAt.IsZero() && state.Quota.NextRecoverAt.After(now) {
							next = state.Quota.NextRecoverAt
						}
						if next.Before(now) {
							next = now
						}
						if state.Quota.Exceeded {
							return true, blockReasonCooldown, next
						}
						return true, blockReasonOther, next
					}
				}
				return false, blockReasonNone, time.Time{}
			}
		}

		// Fast path: check quota group index (O(1) lookup)
		// This is only checked when model state doesn't exist yet
		if HasQuotaGrouping(auth.Provider) {
			quotaGroup := ResolveQuotaGroup(auth.Provider, model)
			if quotaGroup != "" {
				if idx := getQuotaGroupIndex(auth); idx != nil {
					if blocked, next := idx.isGroupBlocked(quotaGroup, now); blocked {
						return true, blockReasonCooldown, next
					}
				}
			}
		}

		return false, blockReasonNone, time.Time{}
	}
	if auth.Unavailable && auth.NextRetryAfter.After(now) {
		next := auth.NextRetryAfter
		if !auth.Quota.NextRecoverAt.IsZero() && auth.Quota.NextRecoverAt.After(now) {
			next = auth.Quota.NextRecoverAt
		}
		if next.Before(now) {
			next = now
		}
		if auth.Quota.Exceeded {
			return true, blockReasonCooldown, next
		}
		return true, blockReasonOther, next
	}
	return false, blockReasonNone, time.Time{}
}

// extractAuthCooldownReason returns a human-readable reason for why an auth is in
// cooldown, derived from model-level state or auth-level error info.
func extractAuthCooldownReason(auth *Auth, model string) string {
	if auth == nil {
		return ""
	}
	if model != "" && len(auth.ModelStates) > 0 {
		if state, ok := auth.ModelStates[model]; ok && state != nil {
			if state.LastError != nil && state.LastError.Message != "" {
				return state.LastError.Message
			}
			if state.StatusMessage != "" {
				return state.StatusMessage
			}
		}
	}
	if auth.LastError != nil && auth.LastError.Message != "" {
		return auth.LastError.Message
	}
	if auth.StatusMessage != "" {
		return auth.StatusMessage
	}
	return ""
}

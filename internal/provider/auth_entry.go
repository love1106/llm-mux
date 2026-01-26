package provider

import (
	"runtime"
	"sync/atomic"
	"time"

	baseauth "github.com/nghyane/llm-mux/internal/auth"
)

// AuthMetadata contains mutable metadata that changes infrequently.
// This struct is immutable after creation - use Copy methods to create modified versions.
// Used with atomic.Pointer for copy-on-write semantics.
type AuthMetadata struct {
	// Label is a user-friendly name for the auth.
	Label string
	// Status indicates the current auth status.
	Status Status
	// StatusMessage provides context for the status.
	StatusMessage string
	// ProxyURL is an optional proxy for this auth.
	ProxyURL string
	// Attributes contains provider-specific string attributes (e.g., api_key).
	Attributes map[string]string
	// Metadata contains provider-specific metadata (e.g., access_token, email).
	Metadata map[string]any
	// LastError records the most recent error.
	LastError *Error
	// CreatedAt is the auth creation timestamp.
	CreatedAt time.Time
	// UpdatedAt is the last modification timestamp.
	UpdatedAt time.Time
	// MaterialVersion tracks credential material changes.
	MaterialVersion int64
	// LastRefreshedAt is the last successful token refresh time.
	LastRefreshedAt time.Time
	// NextRefreshAfter is when the next refresh should occur.
	NextRefreshAfter time.Time
	// NextRetryAfter is when to retry after failure.
	NextRetryAfter time.Time
	// Storage reference for token persistence.
	Storage baseauth.TokenStorage
	// FileName for file-based storage.
	FileName string
	// Runtime holds provider-specific runtime state.
	Runtime any
}

// Clone creates a deep copy of AuthMetadata.
func (m *AuthMetadata) Clone() *AuthMetadata {
	if m == nil {
		return nil
	}
	clone := &AuthMetadata{
		Label:            m.Label,
		Status:           m.Status,
		StatusMessage:    m.StatusMessage,
		ProxyURL:         m.ProxyURL,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		MaterialVersion:  m.MaterialVersion,
		LastRefreshedAt:  m.LastRefreshedAt,
		NextRefreshAfter: m.NextRefreshAfter,
		NextRetryAfter:   m.NextRetryAfter,
		Storage:          m.Storage,
		FileName:         m.FileName,
		Runtime:          m.Runtime,
	}
	if m.LastError != nil {
		clone.LastError = &Error{
			Code:       m.LastError.Code,
			Message:    m.LastError.Message,
			Retryable:  m.LastError.Retryable,
			HTTPStatus: m.LastError.HTTPStatus,
		}
	}
	if len(m.Attributes) > 0 {
		clone.Attributes = make(map[string]string, len(m.Attributes))
		for k, v := range m.Attributes {
			clone.Attributes[k] = v
		}
	}
	if len(m.Metadata) > 0 {
		clone.Metadata = make(map[string]any, len(m.Metadata))
		for k, v := range m.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

// WithStatus returns a copy with updated status.
func (m *AuthMetadata) WithStatus(status Status, message string) *AuthMetadata {
	clone := m.Clone()
	clone.Status = status
	clone.StatusMessage = message
	clone.UpdatedAt = time.Now()
	return clone
}

// WithError returns a copy with updated error.
func (m *AuthMetadata) WithError(err *Error) *AuthMetadata {
	clone := m.Clone()
	if err != nil {
		clone.LastError = &Error{
			Code:       err.Code,
			Message:    err.Message,
			Retryable:  err.Retryable,
			HTTPStatus: err.HTTPStatus,
		}
		clone.Status = StatusError
		if err.Message != "" {
			clone.StatusMessage = err.Message
		}
	}
	clone.UpdatedAt = time.Now()
	return clone
}

// WithAttribute returns a copy with an added/updated attribute.
func (m *AuthMetadata) WithAttribute(key, value string) *AuthMetadata {
	clone := m.Clone()
	if clone.Attributes == nil {
		clone.Attributes = make(map[string]string)
	}
	clone.Attributes[key] = value
	clone.UpdatedAt = time.Now()
	return clone
}

// WithMetadataValue returns a copy with an added/updated metadata value.
func (m *AuthMetadata) WithMetadataValue(key string, value any) *AuthMetadata {
	clone := m.Clone()
	if clone.Metadata == nil {
		clone.Metadata = make(map[string]any)
	}
	clone.Metadata[key] = value
	clone.UpdatedAt = time.Now()
	return clone
}

// WithRefresh returns a copy with updated refresh timestamps.
func (m *AuthMetadata) WithRefresh(lastRefreshed, nextRefresh time.Time) *AuthMetadata {
	clone := m.Clone()
	clone.LastRefreshedAt = lastRefreshed
	clone.NextRefreshAfter = nextRefresh
	clone.LastError = nil
	clone.UpdatedAt = time.Now()
	return clone
}

// WithRetryAfter returns a copy with updated retry time.
func (m *AuthMetadata) WithRetryAfter(retryAfter time.Time) *AuthMetadata {
	clone := m.Clone()
	clone.NextRetryAfter = retryAfter
	clone.UpdatedAt = time.Now()
	return clone
}

// ClearError returns a copy with error cleared and status set to active.
func (m *AuthMetadata) ClearError() *AuthMetadata {
	clone := m.Clone()
	clone.LastError = nil
	clone.Status = StatusActive
	clone.StatusMessage = ""
	clone.NextRetryAfter = time.Time{}
	clone.UpdatedAt = time.Now()
	return clone
}

// ModelStateSnapshot is an immutable snapshot of a model's state.
// Used in ModelStatesSnapshot for COW pattern.
type ModelStateSnapshot struct {
	Status         Status
	StatusMessage  string
	Unavailable    bool
	NextRetryAfter int64 // UnixNano for atomic-friendly storage
	QuotaExceeded  bool
	QuotaReason    string
	QuotaRecover   int64 // UnixNano
	BackoffLevel   int
	LastError      *Error
	UpdatedAt      int64 // UnixNano
}

// Clone creates a copy of the snapshot.
func (s ModelStateSnapshot) Clone() ModelStateSnapshot {
	clone := s
	if s.LastError != nil {
		clone.LastError = &Error{
			Code:       s.LastError.Code,
			Message:    s.LastError.Message,
			Retryable:  s.LastError.Retryable,
			HTTPStatus: s.LastError.HTTPStatus,
		}
	}
	return clone
}

// ToModelState converts to the legacy ModelState format.
func (s ModelStateSnapshot) ToModelState() *ModelState {
	ms := &ModelState{
		Status:        s.Status,
		StatusMessage: s.StatusMessage,
		Unavailable:   s.Unavailable,
		UpdatedAt:     time.Unix(0, s.UpdatedAt),
	}
	if s.NextRetryAfter != 0 {
		ms.NextRetryAfter = time.Unix(0, s.NextRetryAfter)
	}
	if s.LastError != nil {
		ms.LastError = &Error{
			Code:       s.LastError.Code,
			Message:    s.LastError.Message,
			Retryable:  s.LastError.Retryable,
			HTTPStatus: s.LastError.HTTPStatus,
		}
	}
	ms.Quota = QuotaState{
		Exceeded:     s.QuotaExceeded,
		Reason:       s.QuotaReason,
		BackoffLevel: s.BackoffLevel,
	}
	if s.QuotaRecover != 0 {
		ms.Quota.NextRecoverAt = time.Unix(0, s.QuotaRecover)
	}
	return ms
}

// ModelStatesSnapshot is an immutable snapshot of all model states.
// Used with atomic.Pointer for copy-on-write semantics.
type ModelStatesSnapshot struct {
	States map[string]ModelStateSnapshot
}

// Clone creates a deep copy of the snapshot.
func (s *ModelStatesSnapshot) Clone() *ModelStatesSnapshot {
	if s == nil {
		return &ModelStatesSnapshot{
			States: make(map[string]ModelStateSnapshot),
		}
	}
	clone := &ModelStatesSnapshot{
		States: make(map[string]ModelStateSnapshot, len(s.States)),
	}
	for k, v := range s.States {
		clone.States[k] = v.Clone()
	}
	return clone
}

// Get returns the state for a model, or zero value if not found.
func (s *ModelStatesSnapshot) Get(model string) (ModelStateSnapshot, bool) {
	if s == nil || s.States == nil {
		return ModelStateSnapshot{}, false
	}
	state, ok := s.States[model]
	return state, ok
}

// With returns a new snapshot with the given model state updated.
func (s *ModelStatesSnapshot) With(model string, state ModelStateSnapshot) *ModelStatesSnapshot {
	clone := s.Clone()
	clone.States[model] = state
	return clone
}

// Without returns a new snapshot with the given model removed.
func (s *ModelStatesSnapshot) Without(model string) *ModelStatesSnapshot {
	clone := s.Clone()
	delete(clone.States, model)
	return clone
}

// ToModelStates converts to the legacy map format.
func (s *ModelStatesSnapshot) ToModelStates() map[string]*ModelState {
	if s == nil || len(s.States) == 0 {
		return nil
	}
	result := make(map[string]*ModelState, len(s.States))
	for model, snapshot := range s.States {
		result[model] = snapshot.ToModelState()
	}
	return result
}

// QuotaFields holds lock-free quota state using atomics.
type QuotaFields struct {
	ActiveRequests  atomic.Int64
	CooldownUntil   atomic.Int64 // UnixNano
	TotalTokensUsed atomic.Int64
	LastExhaustedAt atomic.Int64 // UnixNano
	LearnedLimit    atomic.Int64
	LearnedCooldown atomic.Int64 // Duration in nanoseconds
	RealQuota       atomic.Pointer[RealQuotaSnapshot]
}

// GetCooldownUntil returns the cooldown deadline as time.Time.
func (q *QuotaFields) GetCooldownUntil() time.Time {
	ns := q.CooldownUntil.Load()
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// SetCooldownUntil sets the cooldown deadline.
func (q *QuotaFields) SetCooldownUntil(t time.Time) {
	if t.IsZero() {
		q.CooldownUntil.Store(0)
	} else {
		q.CooldownUntil.Store(t.UnixNano())
	}
}

// GetLastExhaustedAt returns the last exhausted time.
func (q *QuotaFields) GetLastExhaustedAt() time.Time {
	ns := q.LastExhaustedAt.Load()
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// SetLastExhaustedAt sets the last exhausted time.
func (q *QuotaFields) SetLastExhaustedAt(t time.Time) {
	if t.IsZero() {
		q.LastExhaustedAt.Store(0)
	} else {
		q.LastExhaustedAt.Store(t.UnixNano())
	}
}

// GetLearnedCooldown returns the learned cooldown duration.
func (q *QuotaFields) GetLearnedCooldown() time.Duration {
	return time.Duration(q.LearnedCooldown.Load())
}

// SetLearnedCooldown sets the learned cooldown duration.
func (q *QuotaFields) SetLearnedCooldown(d time.Duration) {
	q.LearnedCooldown.Store(int64(d))
}

// Snapshot returns a point-in-time copy of quota state.
func (q *QuotaFields) Snapshot() *AuthQuotaStateSnapshot {
	return &AuthQuotaStateSnapshot{
		CooldownUntil:   q.GetCooldownUntil(),
		ActiveRequests:  q.ActiveRequests.Load(),
		TotalTokensUsed: q.TotalTokensUsed.Load(),
		LastExhaustedAt: q.GetLastExhaustedAt(),
		LearnedLimit:    q.LearnedLimit.Load(),
		LearnedCooldown: q.GetLearnedCooldown(),
	}
}

// TokenFields holds lock-free token state using atomics.
type TokenFields struct {
	ExpiresAt  atomic.Int64 // UnixNano
	RefreshAt  atomic.Int64 // UnixNano
	Refreshing atomic.Bool
}

// GetExpiresAt returns the token expiration time.
func (t *TokenFields) GetExpiresAt() time.Time {
	ns := t.ExpiresAt.Load()
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// SetExpiresAt sets the token expiration time.
func (t *TokenFields) SetExpiresAt(tm time.Time) {
	if tm.IsZero() {
		t.ExpiresAt.Store(0)
	} else {
		t.ExpiresAt.Store(tm.UnixNano())
	}
}

// GetRefreshAt returns when token refresh should occur.
func (t *TokenFields) GetRefreshAt() time.Time {
	ns := t.RefreshAt.Load()
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// SetRefreshAt sets when token refresh should occur.
func (t *TokenFields) SetRefreshAt(tm time.Time) {
	if tm.IsZero() {
		t.RefreshAt.Store(0)
	} else {
		t.RefreshAt.Store(tm.UnixNano())
	}
}

// NeedsRefresh returns true if token needs refreshing.
func (t *TokenFields) NeedsRefresh(now time.Time) bool {
	refreshAt := t.GetRefreshAt()
	if refreshAt.IsZero() {
		return false
	}
	return now.After(refreshAt) || now.Equal(refreshAt)
}

// TryStartRefresh attempts to mark refresh as in-progress.
// Returns true if this goroutine won the race.
func (t *TokenFields) TryStartRefresh() bool {
	return t.Refreshing.CompareAndSwap(false, true)
}

// FinishRefresh marks refresh as complete.
func (t *TokenFields) FinishRefresh() {
	t.Refreshing.Store(false)
}

// AuthEntry is the unified auth state container.
// Uses atomics and COW pointers for lock-free hot path reads.
type AuthEntry struct {
	// Immutable after creation
	id       string
	provider string
	index    uint64

	// Copy-on-write metadata (atomic pointer)
	metadata atomic.Pointer[AuthMetadata]

	// Lock-free quota state (atomics)
	Quota QuotaFields

	// Lock-free token state (atomics)
	Token TokenFields

	// Lock-free status flags
	disabled    atomic.Bool
	unavailable atomic.Bool

	// COW model states
	modelStates atomic.Pointer[ModelStatesSnapshot]
}

// NewAuthEntry creates a new AuthEntry from an Auth.
func NewAuthEntry(auth *Auth) *AuthEntry {
	if auth == nil {
		return nil
	}

	entry := &AuthEntry{
		id:       auth.ID,
		provider: auth.Provider,
		index:    auth.Index,
	}

	// Initialize metadata
	meta := &AuthMetadata{
		Label:            auth.Label,
		Status:           auth.Status,
		StatusMessage:    auth.StatusMessage,
		ProxyURL:         auth.ProxyURL,
		CreatedAt:        auth.CreatedAt,
		UpdatedAt:        auth.UpdatedAt,
		MaterialVersion:  auth.MaterialVersion,
		LastRefreshedAt:  auth.LastRefreshedAt,
		NextRefreshAfter: auth.NextRefreshAfter,
		NextRetryAfter:   auth.NextRetryAfter,
		Storage:          auth.Storage,
		FileName:         auth.FileName,
		Runtime:          auth.Runtime,
	}
	if auth.LastError != nil {
		meta.LastError = &Error{
			Code:       auth.LastError.Code,
			Message:    auth.LastError.Message,
			Retryable:  auth.LastError.Retryable,
			HTTPStatus: auth.LastError.HTTPStatus,
		}
	}
	if len(auth.Attributes) > 0 {
		meta.Attributes = make(map[string]string, len(auth.Attributes))
		for k, v := range auth.Attributes {
			meta.Attributes[k] = v
		}
	}
	if len(auth.Metadata) > 0 {
		meta.Metadata = make(map[string]any, len(auth.Metadata))
		for k, v := range auth.Metadata {
			meta.Metadata[k] = v
		}
	}
	entry.metadata.Store(meta)

	// Initialize flags
	entry.disabled.Store(auth.Disabled)
	entry.unavailable.Store(auth.Unavailable)

	// Initialize model states
	if len(auth.ModelStates) > 0 {
		snapshot := &ModelStatesSnapshot{
			States: make(map[string]ModelStateSnapshot, len(auth.ModelStates)),
		}
		for model, state := range auth.ModelStates {
			if state == nil {
				continue
			}
			ms := ModelStateSnapshot{
				Status:        state.Status,
				StatusMessage: state.StatusMessage,
				Unavailable:   state.Unavailable,
				BackoffLevel:  state.Quota.BackoffLevel,
				QuotaExceeded: state.Quota.Exceeded,
				QuotaReason:   state.Quota.Reason,
			}
			if !state.NextRetryAfter.IsZero() {
				ms.NextRetryAfter = state.NextRetryAfter.UnixNano()
			}
			if !state.Quota.NextRecoverAt.IsZero() {
				ms.QuotaRecover = state.Quota.NextRecoverAt.UnixNano()
			}
			if !state.UpdatedAt.IsZero() {
				ms.UpdatedAt = state.UpdatedAt.UnixNano()
			}
			if state.LastError != nil {
				ms.LastError = &Error{
					Code:       state.LastError.Code,
					Message:    state.LastError.Message,
					Retryable:  state.LastError.Retryable,
					HTTPStatus: state.LastError.HTTPStatus,
				}
			}
			snapshot.States[model] = ms
		}
		entry.modelStates.Store(snapshot)
	} else {
		entry.modelStates.Store(&ModelStatesSnapshot{
			States: make(map[string]ModelStateSnapshot),
		})
	}

	// Initialize quota from auth
	if !auth.Quota.NextRecoverAt.IsZero() {
		entry.Quota.SetCooldownUntil(auth.Quota.NextRecoverAt)
	}
	if !auth.NextRetryAfter.IsZero() {
		entry.Quota.SetCooldownUntil(auth.NextRetryAfter)
	}

	if meta.Metadata != nil {
		if ts, ok := expirationFromMap(meta.Metadata); ok {
			entry.Token.SetExpiresAt(ts)
		}
	}

	return entry
}

// ID returns the auth ID.
func (e *AuthEntry) ID() string {
	return e.id
}

// Provider returns the provider name.
func (e *AuthEntry) Provider() string {
	return e.provider
}

// Index returns the global index.
func (e *AuthEntry) Index() uint64 {
	return e.index
}

// Metadata returns the current metadata snapshot (immutable).
func (e *AuthEntry) Metadata() *AuthMetadata {
	return e.metadata.Load()
}

// ModelStates returns the current model states snapshot (immutable).
func (e *AuthEntry) ModelStates() *ModelStatesSnapshot {
	return e.modelStates.Load()
}

// IsDisabled returns true if the auth is disabled.
func (e *AuthEntry) IsDisabled() bool {
	return e.disabled.Load()
}

// SetDisabled sets the disabled flag.
func (e *AuthEntry) SetDisabled(disabled bool) {
	e.disabled.Store(disabled)
}

// IsUnavailable returns true if the auth is unavailable.
func (e *AuthEntry) IsUnavailable() bool {
	return e.unavailable.Load()
}

// SetUnavailable sets the unavailable flag.
func (e *AuthEntry) SetUnavailable(unavailable bool) {
	e.unavailable.Store(unavailable)
}

// UpdateMetadata atomically updates metadata using CAS loop.
// The function fn receives the current metadata and returns the new metadata.
// fn MUST NOT modify the input - it must return a new AuthMetadata.
func (e *AuthEntry) UpdateMetadata(fn func(*AuthMetadata) *AuthMetadata) {
	for attempt := 0; attempt < 100; attempt++ {
		old := e.metadata.Load()
		newMeta := fn(old)
		if newMeta == nil {
			return
		}
		if e.metadata.CompareAndSwap(old, newMeta) {
			return
		}
		runtime.Gosched() // Yield on contention
	}
	// Fallback: force store after too many retries
	e.metadata.Store(fn(e.metadata.Load()))
}

// UpdateModelState atomically updates a model's state using CAS loop.
func (e *AuthEntry) UpdateModelState(model string, fn func(ModelStateSnapshot) ModelStateSnapshot) {
	for attempt := 0; attempt < 100; attempt++ {
		old := e.modelStates.Load()
		current, _ := old.Get(model)
		updated := fn(current)
		newSnapshot := old.With(model, updated)
		if e.modelStates.CompareAndSwap(old, newSnapshot) {
			return
		}
		runtime.Gosched()
	}
	// Fallback
	old := e.modelStates.Load()
	current, _ := old.Get(model)
	e.modelStates.Store(old.With(model, fn(current)))
}

// UpdateAllModelStates atomically updates all model states using CAS loop.
// Useful for quota group propagation.
func (e *AuthEntry) UpdateAllModelStates(fn func(*ModelStatesSnapshot) *ModelStatesSnapshot) {
	for attempt := 0; attempt < 100; attempt++ {
		old := e.modelStates.Load()
		newSnapshot := fn(old)
		if newSnapshot == nil {
			return
		}
		if e.modelStates.CompareAndSwap(old, newSnapshot) {
			return
		}
		runtime.Gosched()
	}
	// Fallback
	e.modelStates.Store(fn(e.modelStates.Load()))
}

// ClearModelState removes a model's state.
func (e *AuthEntry) ClearModelState(model string) {
	for attempt := 0; attempt < 100; attempt++ {
		old := e.modelStates.Load()
		newSnapshot := old.Without(model)
		if e.modelStates.CompareAndSwap(old, newSnapshot) {
			return
		}
		runtime.Gosched()
	}
}

// ToAuth converts AuthEntry to the legacy Auth format.
// This creates a snapshot suitable for external consumers.
func (e *AuthEntry) ToAuth() *Auth {
	if e == nil {
		return nil
	}

	meta := e.metadata.Load()
	modelStates := e.modelStates.Load()

	auth := &Auth{
		ID:               e.id,
		Index:            e.index,
		Provider:         e.provider,
		Label:            meta.Label,
		Status:           meta.Status,
		StatusMessage:    meta.StatusMessage,
		Disabled:         e.disabled.Load(),
		Unavailable:      e.unavailable.Load(),
		ProxyURL:         meta.ProxyURL,
		CreatedAt:        meta.CreatedAt,
		UpdatedAt:        meta.UpdatedAt,
		MaterialVersion:  meta.MaterialVersion,
		LastRefreshedAt:  meta.LastRefreshedAt,
		NextRefreshAfter: meta.NextRefreshAfter,
		NextRetryAfter:   meta.NextRetryAfter,
		Storage:          meta.Storage,
		FileName:         meta.FileName,
		Runtime:          meta.Runtime,
		indexAssigned:    true,
	}

	if meta.LastError != nil {
		auth.LastError = &Error{
			Code:       meta.LastError.Code,
			Message:    meta.LastError.Message,
			Retryable:  meta.LastError.Retryable,
			HTTPStatus: meta.LastError.HTTPStatus,
		}
	}

	if len(meta.Attributes) > 0 {
		auth.Attributes = make(map[string]string, len(meta.Attributes))
		for k, v := range meta.Attributes {
			auth.Attributes[k] = v
		}
	}

	if len(meta.Metadata) > 0 {
		auth.Metadata = make(map[string]any, len(meta.Metadata))
		for k, v := range meta.Metadata {
			auth.Metadata[k] = v
		}
	}

	// Convert quota state
	cooldownUntil := e.Quota.GetCooldownUntil()
	if !cooldownUntil.IsZero() {
		auth.Quota.Exceeded = true
		auth.Quota.NextRecoverAt = cooldownUntil
	}

	// Convert model states
	auth.ModelStates = modelStates.ToModelStates()

	return auth
}

// IsBlockedForModel checks if this auth is blocked for a specific model.
func (e *AuthEntry) IsBlockedForModel(model string, now time.Time) (blocked bool, reason blockReason, retryAt time.Time) {
	if e.IsDisabled() {
		return true, blockReasonDisabled, time.Time{}
	}

	meta := e.Metadata()
	if meta.Status == StatusDisabled {
		return true, blockReasonDisabled, time.Time{}
	}

	if model != "" {
		states := e.ModelStates()
		if state, ok := states.Get(model); ok {
			if state.Status == StatusDisabled {
				return true, blockReasonDisabled, time.Time{}
			}
			if state.Unavailable {
				if state.NextRetryAfter == 0 {
					return true, blockReasonOther, time.Time{}
				}
				retryTime := time.Unix(0, state.NextRetryAfter)
				if retryTime.After(now) {
					next := retryTime
					if state.QuotaRecover != 0 {
						recoverTime := time.Unix(0, state.QuotaRecover)
						if recoverTime.After(now) {
							next = recoverTime
						}
					}
					if state.QuotaExceeded {
						return true, blockReasonCooldown, next
					}
					return true, blockReasonOther, next
				}
			}
			return false, blockReasonNone, time.Time{}
		}
	}

	// Check auth-level unavailability
	if e.IsUnavailable() {
		cooldown := e.Quota.GetCooldownUntil()
		if cooldown.After(now) {
			return true, blockReasonCooldown, cooldown
		}
	}

	return false, blockReasonNone, time.Time{}
}

// IncrementActiveRequests atomically increments active request count.
func (e *AuthEntry) IncrementActiveRequests() {
	e.Quota.ActiveRequests.Add(1)
}

// DecrementActiveRequests atomically decrements active request count.
func (e *AuthEntry) DecrementActiveRequests() {
	for {
		current := e.Quota.ActiveRequests.Load()
		if current <= 0 {
			return
		}
		if e.Quota.ActiveRequests.CompareAndSwap(current, current-1) {
			return
		}
	}
}

// RecordTokenUsage adds to the total token usage count.
func (e *AuthEntry) RecordTokenUsage(tokens int64) {
	if tokens > 0 {
		e.Quota.TotalTokensUsed.Add(tokens)
	}
}

// SetCooldown sets the cooldown period.
func (e *AuthEntry) SetCooldown(until time.Time) {
	e.Quota.SetCooldownUntil(until)
	e.Quota.SetLastExhaustedAt(time.Now())
	e.SetUnavailable(true)
}

// ClearCooldown removes the cooldown state.
func (e *AuthEntry) ClearCooldown() {
	e.Quota.SetCooldownUntil(time.Time{})
	e.SetUnavailable(false)
}

// IsInCooldown returns true if the auth is currently in cooldown.
func (e *AuthEntry) IsInCooldown(now time.Time) bool {
	cooldown := e.Quota.GetCooldownUntil()
	return !cooldown.IsZero() && cooldown.After(now)
}

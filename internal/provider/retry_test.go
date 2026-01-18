package provider

import (
	"context"
	"testing"
	"time"
)

func TestNextQuotaCooldown_ExponentialBackoff(t *testing.T) {
	SetQuotaCooldownDisabled(false)
	defer SetQuotaCooldownDisabled(false)

	tests := []struct {
		name         string
		prevLevel    int
		expectMin    time.Duration
		expectMax    time.Duration
		expectLevel  int
	}{
		{
			name:        "level 0 returns 1s",
			prevLevel:   0,
			expectMin:   1 * time.Second,
			expectMax:   1 * time.Second,
			expectLevel: 1,
		},
		{
			name:        "level 1 returns 2s",
			prevLevel:   1,
			expectMin:   2 * time.Second,
			expectMax:   2 * time.Second,
			expectLevel: 2,
		},
		{
			name:        "level 2 returns 4s",
			prevLevel:   2,
			expectMin:   4 * time.Second,
			expectMax:   4 * time.Second,
			expectLevel: 3,
		},
		{
			name:        "level 3 returns 8s",
			prevLevel:   3,
			expectMin:   8 * time.Second,
			expectMax:   8 * time.Second,
			expectLevel: 4,
		},
		{
			name:        "level 10 returns 1024s (~17min)",
			prevLevel:   10,
			expectMin:   1024 * time.Second,
			expectMax:   1024 * time.Second,
			expectLevel: 11,
		},
		{
			name:        "very high level caps at 30min",
			prevLevel:   20,
			expectMin:   30 * time.Minute,
			expectMax:   30 * time.Minute,
			expectLevel: 20,
		},
		{
			name:        "negative level treated as 0",
			prevLevel:   -1,
			expectMin:   1 * time.Second,
			expectMax:   1 * time.Second,
			expectLevel: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cooldown, level := nextQuotaCooldown(tt.prevLevel)
			if cooldown < tt.expectMin || cooldown > tt.expectMax {
				t.Errorf("cooldown = %v, want between %v and %v", cooldown, tt.expectMin, tt.expectMax)
			}
			if level != tt.expectLevel {
				t.Errorf("level = %d, want %d", level, tt.expectLevel)
			}
		})
	}
}

func TestNextQuotaCooldown_DisabledReturnsZero(t *testing.T) {
	SetQuotaCooldownDisabled(true)
	defer SetQuotaCooldownDisabled(false)

	cooldown, level := nextQuotaCooldown(5)
	if cooldown != 0 {
		t.Errorf("expected 0 cooldown when disabled, got %v", cooldown)
	}
	if level != 5 {
		t.Errorf("expected level to remain 5, got %d", level)
	}
}

func TestManager_hasAvailableAuth(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})
	m.Register(ctx, &Auth{ID: "auth2", Provider: "test", Disabled: true})
	m.Register(ctx, &Auth{ID: "auth3", Provider: "other", Status: StatusActive})

	tests := []struct {
		name      string
		providers []string
		model     string
		want      bool
	}{
		{
			name:      "auth available for test provider",
			providers: []string{"test"},
			model:     "",
			want:      true,
		},
		{
			name:      "auth available for other provider",
			providers: []string{"other"},
			model:     "",
			want:      true,
		},
		{
			name:      "no auth for unknown provider",
			providers: []string{"unknown"},
			model:     "",
			want:      false,
		},
		{
			name:      "empty providers returns false",
			providers: []string{},
			model:     "",
			want:      false,
		},
		{
			name:      "nil manager returns false",
			providers: []string{"test"},
			model:     "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var manager *Manager
			if tt.name != "nil manager returns false" {
				manager = m
			}
			got := manager.hasAvailableAuth(tt.providers, tt.model)
			if got != tt.want {
				t.Errorf("hasAvailableAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_hasAvailableAuth_WithCooldown(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	nextRetry := time.Now().Add(1 * time.Hour)
	m.Register(ctx, &Auth{
		ID:             "auth1",
		Provider:       "test",
		Unavailable:    true,
		NextRetryAfter: nextRetry,
	})
	m.Register(ctx, &Auth{
		ID:       "auth2",
		Provider: "test",
		Status:   StatusActive,
	})

	if !m.hasAvailableAuth([]string{"test"}, "") {
		t.Error("expected auth2 to be available")
	}
}

func TestManager_hasAvailableAuth_AllInCooldown(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	nextRetry := time.Now().Add(1 * time.Hour)
	m.Register(ctx, &Auth{
		ID:             "auth1",
		Provider:       "test",
		Unavailable:    true,
		NextRetryAfter: nextRetry,
	})

	if m.hasAvailableAuth([]string{"test"}, "") {
		t.Error("expected no auth to be available when all in cooldown")
	}
}

func TestManager_waitForAvailableAuth_ImmediateAvailable(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})

	err := m.waitForAvailableAuth(ctx, []string{"test"}, "", 5*time.Second)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestManager_waitForAvailableAuth_ZeroMaxWait(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()

	err := m.waitForAvailableAuth(ctx, []string{"test"}, "model", 0)
	if err != nil {
		t.Errorf("expected no error with zero maxWait, got %v", err)
	}
}

func TestManager_waitForAvailableAuth_ContextCanceled(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := m.waitForAvailableAuth(ctx, []string{"nonexistent"}, "model", 5*time.Second)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestManager_waitForAvailableAuth_Timeout(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()

	start := time.Now()
	err := m.waitForAvailableAuth(ctx, []string{"nonexistent"}, "", 600*time.Millisecond)
	elapsed := time.Since(start)

	if err != errCooldownTimeout {
		t.Errorf("expected errCooldownTimeout, got %v", err)
	}
	if elapsed < 500*time.Millisecond {
		t.Errorf("expected wait of at least 500ms, got %v", elapsed)
	}
}

func TestManager_shouldRetryAfterError_NilError(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	if m.shouldRetryAfterError(nil, 0, 3, []string{"test"}, "model") {
		t.Error("expected false for nil error")
	}
}

func TestManager_shouldRetryAfterError_MaxAttemptsReached(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	err := &Error{Code: "quota_error", HTTPStatus: 429, ErrCategory: CategoryQuotaError}
	if m.shouldRetryAfterError(err, 2, 3, []string{"test"}, "model") {
		t.Error("expected false when max attempts reached")
	}
}

func TestManager_shouldRetryAfterError_UserError(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	err := &Error{Code: "bad_request", HTTPStatus: 400, ErrCategory: CategoryUserError}
	if m.shouldRetryAfterError(err, 0, 3, []string{"test"}, "model") {
		t.Error("expected false for user error")
	}
}

func TestManager_shouldRetryAfterError_QuotaErrorWithAvailableAuth(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})

	err := &Error{Code: "quota_error", HTTPStatus: 429, ErrCategory: CategoryQuotaError}
	if !m.shouldRetryAfterError(err, 0, 3, []string{"test"}, "") {
		t.Error("expected true for quota error with available auth")
	}
}

func TestManager_shouldRetryAfterError_TransientError(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})

	err := &Error{Code: "server_error", HTTPStatus: 500, ErrCategory: CategoryTransient}
	if !m.shouldRetryAfterError(err, 0, 3, []string{"test"}, "") {
		t.Error("expected true for transient error with available auth")
	}
}

func TestManager_shouldRetryAfterError_AuthError(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})

	err := &Error{Code: "auth_error", HTTPStatus: 401, ErrCategory: CategoryAuthError}
	if !m.shouldRetryAfterError(err, 0, 3, []string{"test"}, "") {
		t.Error("expected true for auth error with available auth (fallback to another)")
	}
}

func TestManager_shouldRetryAfterError_NotFoundError(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	m.Register(ctx, &Auth{ID: "auth1", Provider: "test", Status: StatusActive})

	err := &Error{Code: "not_found", HTTPStatus: 404, ErrCategory: CategoryNotFound}
	if m.shouldRetryAfterError(err, 0, 3, []string{"test"}, "") {
		t.Error("expected false for not found error")
	}
}

func TestManager_closestCooldownWait(t *testing.T) {
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	defer m.Stop()

	ctx := context.Background()
	nextRetry := time.Now().Add(30 * time.Second)
	m.Register(ctx, &Auth{
		ID:             "auth1",
		Provider:       "test",
		Unavailable:    true,
		NextRetryAfter: nextRetry,
	})
	m.Register(ctx, &Auth{
		ID:             "auth2",
		Provider:       "test",
		Unavailable:    true,
		NextRetryAfter: time.Now().Add(60 * time.Second),
	})

	wait, found := m.closestCooldownWait([]string{"test"}, "")
	if !found {
		t.Error("expected to find cooldown wait")
	}
	if wait < 25*time.Second || wait > 35*time.Second {
		t.Errorf("expected wait around 30s, got %v", wait)
	}
}

func TestManager_closestCooldownWait_NilManager(t *testing.T) {
	var m *Manager
	wait, found := m.closestCooldownWait([]string{"test"}, "model")
	if found {
		t.Error("expected not found for nil manager")
	}
	if wait != 0 {
		t.Error("expected 0 wait for nil manager")
	}
}

func TestCategoryFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: CategoryUnknown,
		},
		{
			name:     "Error with category set",
			err:      &Error{ErrCategory: CategoryQuotaError},
			expected: CategoryQuotaError,
		},
		{
			name:     "Error with transient category",
			err:      &Error{HTTPStatus: 500, ErrCategory: CategoryTransient},
			expected: CategoryTransient,
		},
		{
			name:     "Error with user error category",
			err:      &Error{HTTPStatus: 400, ErrCategory: CategoryUserError},
			expected: CategoryUserError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoryFromError(tt.err)
			if got != tt.expected {
				t.Errorf("categoryFromError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStatusCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: 0,
		},
		{
			name:     "Error with status",
			err:      &Error{HTTPStatus: 429},
			expected: 429,
		},
		{
			name:     "Error without status",
			err:      &Error{Message: "test"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusCodeFromError(tt.err)
			if got != tt.expected {
				t.Errorf("statusCodeFromError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRetryAfterFromError(t *testing.T) {
	type retryAfterError struct {
		Error
		retryAfter *time.Duration
	}

	ra := func(d time.Duration) *time.Duration { return &d }

	tests := []struct {
		name     string
		err      error
		expected *time.Duration
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "error without retry-after",
			err:      &Error{HTTPStatus: 429},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryAfterFromError(tt.err)
			if (got == nil) != (tt.expected == nil) {
				t.Errorf("retryAfterFromError() = %v, want %v", got, tt.expected)
			}
		})
	}

	_ = ra
}

func TestErrorCategory_ShouldFallback(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		want     bool
	}{
		{CategoryUnknown, false},
		{CategoryUserError, false},
		{CategoryAuthError, true},
		{CategoryAuthRevoked, false},
		{CategoryQuotaError, true},
		{CategoryTransient, true},
		{CategoryNotFound, false},
		{CategoryClientCanceled, false},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			if got := tt.category.ShouldFallback(); got != tt.want {
				t.Errorf("ShouldFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCategory_IsUserFault(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		want     bool
	}{
		{CategoryUnknown, false},
		{CategoryUserError, true},
		{CategoryAuthError, false},
		{CategoryQuotaError, false},
		{CategoryTransient, false},
		{CategoryNotFound, true},
		{CategoryClientCanceled, true},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			if got := tt.category.IsUserFault(); got != tt.want {
				t.Errorf("IsUserFault() = %v, want %v", got, tt.want)
			}
		})
	}
}

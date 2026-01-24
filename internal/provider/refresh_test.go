package provider

import (
	"testing"
	"time"
)

func TestShouldRefresh_NilAuth(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	if m.shouldRefresh(nil, now) {
		t.Error("shouldRefresh should return false for nil auth")
	}
}

func TestShouldRefresh_DisabledAuth(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	auth := &Auth{
		ID:       "test-auth",
		Disabled: true,
	}

	if m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return false for disabled auth")
	}
}

func TestShouldRefresh_WaitingForNextRefresh(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	auth := &Auth{
		ID:               "test-auth",
		Disabled:         false,
		NextRefreshAfter: now.Add(time.Hour),
	}

	if m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return false when NextRefreshAfter is in the future")
	}
}

func TestShouldRefresh_NextRefreshAfterPassed(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	auth := &Auth{
		ID:               "test-auth",
		Provider:         "test",
		Disabled:         false,
		NextRefreshAfter: now.Add(-time.Hour),
	}

	RegisterRefreshLeadProvider("test", func() *time.Duration {
		d := 4 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "test")
		refreshLeadMu.Unlock()
	}()

	if !m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return true when NextRefreshAfter has passed and no expiry")
	}
}

func TestShouldRefresh_WithinRefreshLead(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	RegisterRefreshLeadProvider("testprovider", func() *time.Duration {
		d := 4 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "testprovider")
		refreshLeadMu.Unlock()
	}()

	auth := &Auth{
		ID:       "test-auth",
		Provider: "testprovider",
		Disabled: false,
		Metadata: map[string]any{
			"expires_at": now.Add(2 * time.Hour).Format(time.RFC3339),
		},
	}

	if !m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return true when within refresh lead time")
	}
}

func TestShouldRefresh_OutsideRefreshLead(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	RegisterRefreshLeadProvider("testprovider2", func() *time.Duration {
		d := 4 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "testprovider2")
		refreshLeadMu.Unlock()
	}()

	auth := &Auth{
		ID:       "test-auth",
		Provider: "testprovider2",
		Disabled: false,
		Metadata: map[string]any{
			"expires_at": now.Add(8 * time.Hour).Format(time.RFC3339),
		},
	}

	if m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return false when outside refresh lead time")
	}
}

func TestShouldRefresh_NoRefreshLead(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	auth := &Auth{
		ID:       "test-auth",
		Provider: "unknownprovider",
		Disabled: false,
	}

	if m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return false when no refresh lead is registered")
	}
}

func TestShouldRefresh_ExpiredToken(t *testing.T) {
	m := &Manager{
		auths: make(map[string]*Auth),
	}
	now := time.Now()

	RegisterRefreshLeadProvider("testprovider3", func() *time.Duration {
		d := 4 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "testprovider3")
		refreshLeadMu.Unlock()
	}()

	auth := &Auth{
		ID:       "test-auth",
		Provider: "testprovider3",
		Disabled: false,
		Metadata: map[string]any{
			"expires_at": now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	if !m.shouldRefresh(auth, now) {
		t.Error("shouldRefresh should return true for expired token")
	}
}

func TestAuthPreferredInterval_NilAuth(t *testing.T) {
	if interval := authPreferredInterval(nil); interval != 0 {
		t.Errorf("authPreferredInterval should return 0 for nil auth, got %v", interval)
	}
}

func TestAuthPreferredInterval_FromMetadata(t *testing.T) {
	auth := &Auth{
		ID: "test-auth",
		Metadata: map[string]any{
			"refresh_interval_seconds": 3600,
		},
	}

	interval := authPreferredInterval(auth)
	if interval != time.Hour {
		t.Errorf("Expected 1 hour interval, got %v", interval)
	}
}

func TestAuthPreferredInterval_FromAttributes(t *testing.T) {
	auth := &Auth{
		ID: "test-auth",
		Attributes: map[string]string{
			"refresh_interval_seconds": "7200",
		},
	}

	interval := authPreferredInterval(auth)
	if interval != 2*time.Hour {
		t.Errorf("Expected 2 hour interval, got %v", interval)
	}
}

func TestAuthPreferredInterval_NoInterval(t *testing.T) {
	auth := &Auth{
		ID: "test-auth",
	}

	interval := authPreferredInterval(auth)
	if interval != 0 {
		t.Errorf("Expected 0 interval, got %v", interval)
	}
}

func TestProviderRefreshLead_RegisteredProvider(t *testing.T) {
	RegisterRefreshLeadProvider("testlead", func() *time.Duration {
		d := 6 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "testlead")
		refreshLeadMu.Unlock()
	}()

	lead := ProviderRefreshLead("testlead", nil)
	if lead == nil {
		t.Fatal("Expected non-nil lead")
	}
	if *lead != 6*time.Hour {
		t.Errorf("Expected 6 hour lead, got %v", *lead)
	}
}

func TestProviderRefreshLead_UnregisteredProvider(t *testing.T) {
	lead := ProviderRefreshLead("nonexistent", nil)
	if lead != nil {
		t.Errorf("Expected nil lead for unregistered provider, got %v", *lead)
	}
}

func TestProviderRefreshLead_CaseInsensitive(t *testing.T) {
	RegisterRefreshLeadProvider("caselead", func() *time.Duration {
		d := 3 * time.Hour
		return &d
	})
	defer func() {
		refreshLeadMu.Lock()
		delete(refreshLeadFactories, "caselead")
		refreshLeadMu.Unlock()
	}()

	lead := ProviderRefreshLead("CASELEAD", nil)
	if lead == nil {
		t.Fatal("Expected non-nil lead for case-insensitive lookup")
	}
	if *lead != 3*time.Hour {
		t.Errorf("Expected 3 hour lead, got %v", *lead)
	}
}

func TestDurationFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		meta     map[string]any
		keys     []string
		expected time.Duration
	}{
		{
			name:     "nil metadata",
			meta:     nil,
			keys:     []string{"key"},
			expected: 0,
		},
		{
			name:     "empty metadata",
			meta:     map[string]any{},
			keys:     []string{"key"},
			expected: 0,
		},
		{
			name:     "int seconds",
			meta:     map[string]any{"refresh_interval_seconds": 3600},
			keys:     []string{"refresh_interval_seconds"},
			expected: time.Hour,
		},
		{
			name:     "float seconds",
			meta:     map[string]any{"refresh_interval_seconds": 3600.0},
			keys:     []string{"refresh_interval_seconds"},
			expected: time.Hour,
		},
		{
			name:     "string seconds",
			meta:     map[string]any{"refresh_interval_seconds": "3600"},
			keys:     []string{"refresh_interval_seconds"},
			expected: time.Hour,
		},
		{
			name:     "duration string",
			meta:     map[string]any{"refresh_interval": "2h"},
			keys:     []string{"refresh_interval"},
			expected: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := durationFromMetadata(tt.meta, tt.keys...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDurationFromAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]string
		keys     []string
		expected time.Duration
	}{
		{
			name:     "nil attributes",
			attrs:    nil,
			keys:     []string{"key"},
			expected: 0,
		},
		{
			name:     "empty attributes",
			attrs:    map[string]string{},
			keys:     []string{"key"},
			expected: 0,
		},
		{
			name:     "numeric string seconds",
			attrs:    map[string]string{"refresh_interval_seconds": "7200"},
			keys:     []string{"refresh_interval_seconds"},
			expected: 2 * time.Hour,
		},
		{
			name:     "duration string",
			attrs:    map[string]string{"refresh_interval": "30m"},
			keys:     []string{"refresh_interval"},
			expected: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := durationFromAttributes(tt.attrs, tt.keys...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewRefreshSemaphore(t *testing.T) {
	sem := newRefreshSemaphore()
	if sem == nil {
		t.Fatal("Expected non-nil semaphore")
	}
}

func TestAuthLastRefreshTimestamp(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		auth     *Auth
		wantTime time.Time
		wantOK   bool
	}{
		{
			name:   "nil auth",
			auth:   nil,
			wantOK: false,
		},
		{
			name:   "no metadata or attributes",
			auth:   &Auth{ID: "test"},
			wantOK: false,
		},
		{
			name: "from metadata",
			auth: &Auth{
				ID: "test",
				Metadata: map[string]any{
					"last_refresh": now.Format(time.RFC3339),
				},
			},
			wantTime: now,
			wantOK:   true,
		},
		{
			name: "from attributes",
			auth: &Auth{
				ID: "test",
				Attributes: map[string]string{
					"last_refresh": now.Format(time.RFC3339),
				},
			},
			wantTime: now,
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotOK := authLastRefreshTimestamp(tt.auth)
			if gotOK != tt.wantOK {
				t.Errorf("Expected ok=%v, got ok=%v", tt.wantOK, gotOK)
			}
			if tt.wantOK && !gotTime.Equal(tt.wantTime) {
				t.Errorf("Expected time=%v, got time=%v", tt.wantTime, gotTime)
			}
		})
	}
}

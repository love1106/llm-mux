package provider

import (
	"context"
	"testing"
	"time"
)

// TestCircuitBreaker_Registry_401DisablesAuth verifies that a 401 upstream error
// auto-disables the auth entry in the AuthRegistry path.
func TestCircuitBreaker_Registry_401DisablesAuth(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-401-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-401-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "unauthorized",
			Message:    "Unauthorized",
			HTTPStatus: 401,
		},
	})

	entry := registry.GetEntry("cb-401-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if !entry.IsDisabled() {
		t.Error("expected auth to be disabled after 401")
	}
	meta := entry.Metadata()
	if meta.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", meta.Status)
	}
}

// TestCircuitBreaker_Registry_403DisablesAuth verifies that a 403 upstream error
// auto-disables the auth entry in the AuthRegistry path.
func TestCircuitBreaker_Registry_403DisablesAuth(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-403-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-403-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "forbidden",
			Message:    "Forbidden - subscription expired",
			HTTPStatus: 403,
		},
	})

	entry := registry.GetEntry("cb-403-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if !entry.IsDisabled() {
		t.Error("expected auth to be disabled after 403")
	}
	meta := entry.Metadata()
	if meta.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", meta.Status)
	}
}

// TestCircuitBreaker_Registry_429DoesNotDisable verifies that a 429 error
// does NOT disable the auth (only model-level cooldown).
func TestCircuitBreaker_Registry_429DoesNotDisable(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-429-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-429-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "rate_limited",
			Message:    "Rate limit exceeded",
			HTTPStatus: 429,
		},
	})

	entry := registry.GetEntry("cb-429-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if entry.IsDisabled() {
		t.Error("expected auth to NOT be disabled after 429")
	}
	meta := entry.Metadata()
	if meta.Status == StatusDisabled {
		t.Error("expected status to NOT be StatusDisabled after 429")
	}
}

// TestCircuitBreaker_Registry_OAuthRevokedDisablesAuth verifies that an OAuth
// token revoked error (detected by message) auto-disables the auth.
func TestCircuitBreaker_Registry_OAuthRevokedDisablesAuth(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-revoked-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-revoked-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "forbidden",
			Message:    "OAuth token has been revoked. Please obtain a new token.",
			HTTPStatus: 403,
		},
	})

	entry := registry.GetEntry("cb-revoked-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if !entry.IsDisabled() {
		t.Error("expected auth to be disabled after OAuth revoked error")
	}
	meta := entry.Metadata()
	if meta.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", meta.Status)
	}
}

// TestCircuitBreaker_Registry_SuccessDoesNotDisable verifies that a successful
// result does not disable the auth.
func TestCircuitBreaker_Registry_SuccessDoesNotDisable(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-success-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-success-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  true,
	})

	entry := registry.GetEntry("cb-success-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if entry.IsDisabled() {
		t.Error("expected auth to NOT be disabled after success")
	}
	meta := entry.Metadata()
	if meta.Status != StatusActive {
		t.Errorf("expected StatusActive, got %v", meta.Status)
	}
}

// TestCircuitBreaker_Registry_500DoesNotDisable verifies that server errors
// do NOT disable the auth (they are transient).
func TestCircuitBreaker_Registry_500DoesNotDisable(t *testing.T) {
	registry := NewAuthRegistry(nil, nil)
	ctx := context.Background()

	auth := &Auth{
		ID:       "cb-500-registry",
		Provider: "claude",
		Status:   StatusActive,
	}
	_, _ = registry.Register(ctx, auth)

	registry.MarkResult(ctx, Result{
		AuthID:   "cb-500-registry",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "internal_error",
			Message:    "Internal server error",
			HTTPStatus: 500,
		},
	})

	entry := registry.GetEntry("cb-500-registry")
	if entry == nil {
		t.Fatal("entry not found")
	}
	if entry.IsDisabled() {
		t.Error("expected auth to NOT be disabled after 500")
	}
}

// TestCircuitBreaker_Manager_401DisablesAuth tests the Manager.markResultSync path.
func TestCircuitBreaker_Manager_401DisablesAuth(t *testing.T) {
	manager := NewManager(nil, nil, nil)
	defer manager.Stop()
	ctx := context.Background()

	auth := &Auth{ID: "cb-401-manager", Provider: "claude", Status: StatusActive}
	_, _ = manager.Register(ctx, auth)

	manager.MarkResult(ctx, Result{
		AuthID:   "cb-401-manager",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "unauthorized",
			Message:    "Unauthorized",
			HTTPStatus: 401,
		},
	})

	// MarkResult is async; wait for processing
	time.Sleep(100 * time.Millisecond)

	got, ok := manager.GetByID("cb-401-manager")
	if !ok {
		t.Fatal("auth not found")
	}
	if !got.Disabled {
		t.Error("expected auth.Disabled=true after 401")
	}
	if got.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", got.Status)
	}
}

// TestCircuitBreaker_Manager_403DisablesAuth tests the Manager.markResultSync path.
func TestCircuitBreaker_Manager_403DisablesAuth(t *testing.T) {
	manager := NewManager(nil, nil, nil)
	defer manager.Stop()
	ctx := context.Background()

	auth := &Auth{ID: "cb-403-manager", Provider: "claude", Status: StatusActive}
	_, _ = manager.Register(ctx, auth)

	manager.MarkResult(ctx, Result{
		AuthID:   "cb-403-manager",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "forbidden",
			Message:    "Forbidden",
			HTTPStatus: 403,
		},
	})

	time.Sleep(100 * time.Millisecond)

	got, ok := manager.GetByID("cb-403-manager")
	if !ok {
		t.Fatal("auth not found")
	}
	if !got.Disabled {
		t.Error("expected auth.Disabled=true after 403")
	}
	if got.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", got.Status)
	}
}

// TestCircuitBreaker_Manager_429DoesNotDisable tests that 429 does NOT disable in Manager path.
func TestCircuitBreaker_Manager_429DoesNotDisable(t *testing.T) {
	manager := NewManager(nil, nil, nil)
	defer manager.Stop()
	ctx := context.Background()

	auth := &Auth{ID: "cb-429-manager", Provider: "claude", Status: StatusActive}
	_, _ = manager.Register(ctx, auth)

	manager.MarkResult(ctx, Result{
		AuthID:   "cb-429-manager",
		Provider: "claude",
		Model:    "claude-sonnet-4",
		Success:  false,
		Error: &Error{
			Code:       "rate_limited",
			Message:    "Rate limit exceeded",
			HTTPStatus: 429,
		},
	})

	time.Sleep(100 * time.Millisecond)

	got, ok := manager.GetByID("cb-429-manager")
	if !ok {
		t.Fatal("auth not found")
	}
	if got.Disabled {
		t.Error("expected auth.Disabled=false after 429")
	}
	if got.Status == StatusDisabled {
		t.Error("expected status to NOT be StatusDisabled after 429")
	}
}

// TestCircuitBreaker_DisabledAuthNotPicked verifies that the selector skips
// disabled auths, preventing further routing to them.
func TestCircuitBreaker_DisabledAuthNotPicked(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "disabled-auth", Provider: "claude", Disabled: true, Status: StatusDisabled},
	}

	_, err := selector.Pick(context.Background(), "claude", "claude-sonnet-4", Options{}, auths)
	if err == nil {
		t.Fatal("expected error when all auths are disabled")
	}
}

// TestCircuitBreaker_DisabledAuthSkippedInRotation verifies that only the
// non-disabled auth is picked when one is disabled.
func TestCircuitBreaker_DisabledAuthSkippedInRotation(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "disabled-auth", Provider: "claude", Disabled: true, Status: StatusDisabled},
		{ID: "active-auth", Provider: "claude", Status: StatusActive},
	}

	picked, err := selector.Pick(context.Background(), "claude", "claude-sonnet-4", Options{}, auths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if picked.ID != "active-auth" {
		t.Errorf("expected active-auth to be picked, got %s", picked.ID)
	}
}

// TestCircuitBreaker_ApplyAuthFailureState_401 tests the applyAuthFailureState function
// directly for 401 errors.
func TestCircuitBreaker_ApplyAuthFailureState_401(t *testing.T) {
	auth := &Auth{ID: "state-401", Provider: "claude", Status: StatusActive}
	now := time.Now()

	applyAuthFailureState(auth, &Error{
		Message:    "Unauthorized",
		HTTPStatus: 401,
	}, nil, now)

	if !auth.Disabled {
		t.Error("expected auth.Disabled=true after 401")
	}
	if auth.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", auth.Status)
	}
}

// TestCircuitBreaker_ApplyAuthFailureState_403 tests the applyAuthFailureState function
// directly for 403 errors.
func TestCircuitBreaker_ApplyAuthFailureState_403(t *testing.T) {
	auth := &Auth{ID: "state-403", Provider: "claude", Status: StatusActive}
	now := time.Now()

	applyAuthFailureState(auth, &Error{
		Message:    "Forbidden - subscription expired",
		HTTPStatus: 403,
	}, nil, now)

	if !auth.Disabled {
		t.Error("expected auth.Disabled=true after 403")
	}
	if auth.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", auth.Status)
	}
}

// TestCircuitBreaker_ApplyAuthFailureState_OAuthRevoked tests the applyAuthFailureState
// function for OAuth token revoked messages.
func TestCircuitBreaker_ApplyAuthFailureState_OAuthRevoked(t *testing.T) {
	auth := &Auth{ID: "state-revoked", Provider: "claude", Status: StatusActive}
	now := time.Now()

	applyAuthFailureState(auth, &Error{
		Message:    "OAuth token has been revoked. Please obtain a new token.",
		HTTPStatus: 403,
	}, nil, now)

	if !auth.Disabled {
		t.Error("expected auth.Disabled=true after OAuth revoked error")
	}
	if auth.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %v", auth.Status)
	}
}

// TestCircuitBreaker_ApplyAuthFailureState_429NotDisabled tests that 429 errors
// do NOT disable the auth in applyAuthFailureState.
func TestCircuitBreaker_ApplyAuthFailureState_429NotDisabled(t *testing.T) {
	auth := &Auth{ID: "state-429", Provider: "claude", Status: StatusActive}
	now := time.Now()

	applyAuthFailureState(auth, &Error{
		Message:    "Rate limit exceeded",
		HTTPStatus: 429,
	}, nil, now)

	if auth.Disabled {
		t.Error("expected auth.Disabled=false after 429")
	}
	if auth.Status == StatusDisabled {
		t.Error("expected status to NOT be StatusDisabled after 429")
	}
}

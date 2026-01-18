# Phase 2: Authentication & Security Tests

## Context Links
- [Main Plan](./plan.md)
- [Phase 1: Provider Core](./phase-01-provider-selector-retry-tests.md)
- [Research: Go Testing Patterns](./research/researcher-01-go-testing-patterns.md)

## Overview
Implement unit tests for authentication, access control, and OAuth state management.
Focus on security-critical paths and state machine correctness.

## Requirements
- 80% test coverage for access control and OAuth components
- Test all error conditions and edge cases
- Verify state transitions
- Mock external HTTP calls

## Related Code Files

### 1. internal/access/manager.go (Lines 46-150)
```go
func (m *Manager) Authenticate(ctx context.Context, r *http.Request) (*Result, error)
// Critical: Bearer token validation
// Critical: ErrNoCredentials vs ErrInvalidCredential
// Critical: Provider chain authentication
```

### 2. internal/auth/claude/anthropic_auth.go (Lines 100-400)
```go
func GeneratePKCECodes() (verifier, challenge string, err error)
func DetermineSubscriptionType(rawResp json.RawMessage) string
func parseSessionKey(rawResp []byte) (string, error)
// Skip network calls, test pure logic only
```

### 3. internal/oauth/registry.go (Lines 50-250)
```go
func (r *Registry) Create(provider string) (*State, error)
func (r *Registry) Get(stateID string) (*State, error)
func (r *Registry) Complete(stateID string, result interface{}) error
func (r *Registry) Fail(stateID string, err error) error
func (r *Registry) Cancel(stateID string) error
// Critical: State machine transitions
// Critical: TTL expiration handling
```

## Implementation Steps

### Step 1: access_manager_test.go (1.5h)
```go
func TestManager_Authenticate(t *testing.T) {
    tests := []struct {
        name      string
        headers   map[string]string
        providers []Provider
        wantRes   *Result
        wantErr   error
    }{
        {
            name:    "no_auth_header",
            headers: map[string]string{},
            wantErr: ErrNoCredentials,
        },
        {
            name:    "invalid_bearer_token",
            headers: map[string]string{"Authorization": "Bearer invalid"},
            wantErr: ErrInvalidCredential,
        },
        {
            name:    "valid_token_first_provider",
            headers: map[string]string{"Authorization": "Bearer valid123"},
            wantRes: &Result{UserID: "user1", Provider: "provider1"},
        },
        {
            name:    "fallback_to_second_provider",
            headers: map[string]string{"Authorization": "Bearer fallback"},
            wantRes: &Result{UserID: "user2", Provider: "provider2"},
        },
        {
            name:    "x_api_key_header",
            headers: map[string]string{"X-API-Key": "key123"},
            wantRes: &Result{UserID: "api-user", Provider: "api"},
        },
    }
}

func TestManager_ConcurrentAuth(t *testing.T) {
    // Test concurrent authentication requests
}

// Mock provider for testing
type mockProvider struct {
    authFunc func(context.Context, *http.Request) (*Result, error)
}
```

### Step 2: anthropic_auth_test.go (1h)
```go
func TestGeneratePKCECodes(t *testing.T) {
    // Test PKCE generation
    verifier, challenge, err := GeneratePKCECodes()
    require.NoError(t, err)

    // Verify lengths
    assert.Len(t, verifier, 128)
    assert.NotEmpty(t, challenge)

    // Verify base64url encoding
    _, err = base64.RawURLEncoding.DecodeString(challenge)
    assert.NoError(t, err)

    // Verify SHA256 relationship
    h := sha256.Sum256([]byte(verifier))
    expected := base64.RawURLEncoding.EncodeToString(h[:])
    assert.Equal(t, expected, challenge)
}

func TestDetermineSubscriptionType(t *testing.T) {
    tests := []struct {
        name     string
        response string
        want     string
    }{
        {
            name:     "pro_subscription",
            response: `{"account": {"billing_status": "active_pro"}}`,
            want:     "pro",
        },
        {
            name:     "free_subscription",
            response: `{"account": {"billing_status": "free"}}`,
            want:     "free",
        },
        {
            name:     "unknown_format",
            response: `{"different": "structure"}`,
            want:     "unknown",
        },
    }
}

func TestParseSessionKey(t *testing.T) {
    // Test session key extraction from various response formats
}
```

### Step 3: oauth_registry_test.go (0.5h)
```go
func TestRegistry_StateLifecycle(t *testing.T) {
    reg := NewRegistry(5 * time.Minute)

    // Create state
    state, err := reg.Create("github")
    require.NoError(t, err)
    assert.NotEmpty(t, state.ID)
    assert.Equal(t, "github", state.Provider)
    assert.Equal(t, StatusPending, state.Status)

    // Get state
    retrieved, err := reg.Get(state.ID)
    require.NoError(t, err)
    assert.Equal(t, state.ID, retrieved.ID)

    // Complete state
    err = reg.Complete(state.ID, map[string]string{"token": "abc123"})
    require.NoError(t, err)

    retrieved, err = reg.Get(state.ID)
    require.NoError(t, err)
    assert.Equal(t, StatusCompleted, retrieved.Status)

    // Cannot transition completed state
    err = reg.Fail(state.ID, errors.New("test"))
    assert.Error(t, err)
}

func TestRegistry_TTLExpiration(t *testing.T) {
    reg := NewRegistry(100 * time.Millisecond)

    state, _ := reg.Create("test")

    // Should exist initially
    _, err := reg.Get(state.ID)
    assert.NoError(t, err)

    // Wait for expiration
    time.Sleep(150 * time.Millisecond)

    // Should be expired
    _, err = reg.Get(state.ID)
    assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
    // Test concurrent state operations
}
```

## Test Data Setup
```go
// Helper functions
func newTestRequest(headers map[string]string) *http.Request
func newTestManager(providers ...Provider) *Manager
func mockOAuthResponse(status string, data interface{}) []byte
```

## Success Criteria
- [ ] 80% coverage for access/manager.go
- [ ] 80% coverage for auth/claude/anthropic_auth.go (pure functions only)
- [ ] 85% coverage for oauth/registry.go
- [ ] All security edge cases tested
- [ ] State machine transitions verified

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| External API mocking | High | Test only pure functions, skip network code |
| Security test coverage | High | Focus on auth validation paths |
| State race conditions | Medium | Use mutex locks, test with -race |
| TTL timing issues | Low | Use shorter durations in tests |

## Validation Commands
```bash
# Run phase 2 tests
go test -v -race ./internal/access/... ./internal/auth/... ./internal/oauth/...

# Coverage report
go test -coverprofile=phase2.out ./internal/access/... ./internal/auth/... ./internal/oauth/...
go tool cover -html=phase2.out

# Security-focused testing
go test -tags=security ./internal/access/...
```
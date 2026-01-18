# Phase 1: Provider Core Tests

## Context Links
- [Main Plan](./plan.md)
- [Research: Go Testing Patterns](./research/researcher-01-go-testing-patterns.md)
- [Research: Existing Tests](./research/researcher-02-existing-tests-analysis.md)

## Overview
Implement comprehensive unit tests for critical provider selection and retry logic. These components
handle the core business logic of auth selection, round-robin distribution, and failure recovery.

## Requirements
- 85% test coverage for selector.go, retry.go
- Race-safe concurrent testing
- Mock external dependencies
- Benchmark critical paths

## Related Code Files

### 1. internal/provider/selector.go (Lines 124-290)
```go
func (s *RoundRobinSelector) Pick(...) (*Auth, error)
// Critical: Round-robin with sticky sessions
// Critical: isAuthBlockedForModel() filtering
// Critical: Cooldown/disabled account handling
```

### 2. internal/provider/retry.go (Lines 87-350)
```go
func shouldRetryAfterError(err error, attempt, maxAttempts int, ...) bool
func hasAvailableAuth(providers []string, model string) bool
func waitForAvailableAuth(ctx context.Context, ...) error
func nextQuotaCooldown(failureCount int) time.Duration
```

### 3. internal/provider/quota_manager.go (Lines 50-200)
```go
func UpdateQuota(auth *Auth, model string, state QuotaState)
func IsBlocked(auth *Auth, model string) bool
func GetQuotaGroup(auth *Auth) *QuotaGroup
```

### 4. internal/provider/manager.go (Lines 150-400)
```go
func SelectAuth(ctx context.Context, providers []string, model string, ...) (*Auth, error)
func filterProvidersByModel(providers []string, model string) []string
```

## Implementation Steps

### Step 1: selector_test.go (2h)
```go
func TestRoundRobinSelector_Pick(t *testing.T) {
    tests := []struct {
        name     string
        auths    []*Auth
        opts     Options
        expected *Auth
        wantErr  bool
    }{
        // Test cases for:
        // 1. Basic round-robin
        // 2. Sticky session continuity (60s)
        // 3. Cooldown filtering
        // 4. Disabled account skipping
        // 5. Model blocking (isAuthBlockedForModel)
        // 6. Empty auth list
        // 7. All auths blocked
    }
}

func TestRoundRobinSelector_Concurrent(t *testing.T) {
    // Race condition testing with parallel goroutines
}

func BenchmarkSelector_Pick(b *testing.B) {
    // Performance with 100, 1000, 10000 auths
}
```

### Step 2: retry_test.go (1.5h)
```go
func TestManager_shouldRetryAfterError(t *testing.T) {
    tests := []struct {
        name         string
        err          error
        attempt      int
        maxAttempts  int
        providers    []string
        model        string
        wantRetry    bool
    }{
        // Test cases for:
        // 1. Rate limit errors (429)
        // 2. Transient errors (500, 502, 503)
        // 3. Non-retryable errors (401, 403)
        // 4. Max attempts exceeded
        // 5. No available auth
    }
}

func TestManager_waitForAvailableAuth(t *testing.T) {
    // Test timeout scenarios
    // Test context cancellation
    // Test successful wait
}

func TestNextQuotaCooldown(t *testing.T) {
    // Verify exponential backoff: 1s → 2s → 4s → ... → 30min
}
```

### Step 3: quota_manager_test.go enhancements (0.5h)
```go
func TestQuotaManager_ModelLevelBlocking(t *testing.T) {
    // Model-specific quota states
    // Quota group transitions
}

func TestQuotaManager_ConcurrentUpdates(t *testing.T) {
    // Race-safe quota updates
}
```

## Test Data Setup
```go
// Helper functions
func newTestAuth(id string, provider string) *Auth
func newTestSelector(auths []*Auth) *RoundRobinSelector
func newTestManager() *Manager
func mockTimeNow(t time.Time) func()
```

## Success Criteria
- [ ] 85% coverage for selector.go
- [ ] 85% coverage for retry.go
- [ ] All tests pass with `go test -race`
- [ ] Benchmarks show <1ms for Pick() with 1000 auths
- [ ] No flaky tests in 100 consecutive runs

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex mocking | High | Use interfaces, avoid deep mocks |
| Time-dependent tests | Medium | Mock time.Now(), use fixed durations |
| Race conditions | High | Use sync.Mutex where needed, test with -race |
| State pollution | Medium | Reset state in t.Cleanup() |

## Validation Commands
```bash
# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./internal/provider/...

# View coverage report
go tool cover -html=coverage.out

# Benchmark tests
go test -bench=. -benchmem ./internal/provider/...

# Stress test for flakiness
for i in {1..100}; do go test ./internal/provider/... || break; done
```
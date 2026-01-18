# Go Unit Testing Patterns for Proxy/Multiplexer Applications

## 1. Table-Driven Test Patterns

### Approach
- Use subtests with `t.Run()` for multiple test cases
- Leverage anonymous structs for test case definitions
- Separate test inputs, expected outputs, and error conditions

```go
func TestRoundRobinSelector(t *testing.T) {
    testCases := []struct {
        name           string
        initialTargets []string
        selections     int
        expectedOrder  []string
        stickySession  bool
    }{
        {
            name:           "Basic round-robin without sticky session",
            initialTargets: []string{"target1", "target2", "target3"},
            selections:     6,
            expectedOrder:  []string{"target1", "target2", "target3", "target1", "target2", "target3"},
            stickySession:  false,
        },
        // Additional test cases
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            selector := NewRoundRobinSelector(tc.initialTargets, tc.stickySession)
            for i := 0; i < tc.selections; i++ {
                selected := selector.Select()
                assert.Equal(t, tc.expectedOrder[i], selected)
            }
        })
    }
}
```

## 2. Mocking HTTP Clients and Dependencies

### Strategies
- Use interfaces to abstract external dependencies
- Create mock implementations for testing
- Leverage `httptest.NewServer` for controlled HTTP testing

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type MockHTTPClient struct {
    MockDo func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.MockDo(req)
}

func TestRetryLogic(t *testing.T) {
    mockClient := &MockHTTPClient{
        MockDo: func(req *http.Request) (*http.Response, error) {
            // Simulate specific response scenarios
            return nil, errors.New("temporary failure")
        },
    }

    retryClient := NewRetryClient(mockClient, RetryConfig{
        MaxRetries: 3,
        BackoffStrategy: ExponentialBackoff,
    })

    // Test retry behavior
    resp, err := retryClient.Do(someRequest)
    assert.Error(t, err)
    // Verify retry attempts
}
```

## 3. Testing Concurrent Components

### Approach
- Use race detector (`go test -race`)
- Create deterministic test scenarios
- Use channels and synchronization primitives

```go
func TestConcurrentRoundRobinSelector(t *testing.T) {
    targets := []string{"target1", "target2", "target3"}
    selector := NewRoundRobinSelector(targets, false)

    results := make(chan string, 100)
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            results <- selector.Select()
        }()
    }

    wg.Wait()
    close(results)

    // Verify distribution and thread-safety
    distribution := make(map[string]int)
    for result := range results {
        distribution[result]++
    }

    assert.Len(t, distribution, len(targets))
    for _, count := range distribution {
        assert.InDelta(t, 100/len(targets), count, 10)
    }
}
```

## 4. Exponential Backoff and Retry Logic

### Test Pattern
- Test different failure scenarios
- Verify backoff timing and retry attempts
- Mock time for predictable testing

```go
func TestExponentialBackoff(t *testing.T) {
    testCases := []struct {
        name             string
        maxRetries       int
        expectedAttempts int
        failureMode      func() error
    }{
        {
            name:             "Intermittent Failures",
            maxRetries:       3,
            expectedAttempts: 4,
            failureMode: func() error {
                // Simulate occasional failures
                static attempts int
                attempts++
                if attempts < 3 {
                    return errors.New("temporary error")
                }
                return nil
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            attempts := 0
            err := RetryWithBackoff(tc.maxRetries, func() error {
                attempts++
                return tc.failureMode()
            })

            assert.NoError(t, err)
            assert.Equal(t, tc.expectedAttempts, attempts)
        })
    }
}
```

## 5. Interface-Based Mocking Strategies

### Approach
- Define clear interfaces for components
- Create mock implementations for testing
- Use dependency injection for flexibility

```go
type StateTransitioner interface {
    Transition(currentState, event string) (string, error)
}

type MockStateTransitioner struct {
    MockTransition func(currentState, event string) (string, error)
}

func (m *MockStateTransitioner) Transition(currentState, event string) (string, error) {
    return m.MockTransition(currentState, event)
}

func TestOAuthStateMachine(t *testing.T) {
    mockTransitioner := &MockStateTransitioner{
        MockTransition: func(currentState, event string) (string, error) {
            // Simulate state machine logic
            switch {
            case currentState == "INIT" && event == "START_AUTH":
                return "AUTHENTICATING", nil
            case currentState == "AUTHENTICATING" && event == "SUCCESS":
                return "AUTHORIZED", nil
            default:
                return "", errors.New("invalid transition")
            }
        },
    }

    stateMachine := NewOAuthStateMachine(mockTransitioner)

    // Test various state transitions
    newState, err := stateMachine.ProcessEvent("START_AUTH")
    assert.NoError(t, err)
    assert.Equal(t, "AUTHENTICATING", newState)
}
```

## Unresolved Questions
- How to effectively mock time-dependent operations without introducing significant complexity
- Best practices for testing probabilistic components (e.g., load balancers)
- Performance impact of extensive mocking in large test suites

## Recommendations
1. Use table-driven tests for comprehensive coverage
2. Leverage interfaces for dependency injection and mocking
3. Always use race detector for concurrent code
4. Create predictable, deterministic test scenarios
5. Mock external dependencies with controlled behavior
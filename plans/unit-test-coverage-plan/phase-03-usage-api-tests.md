# Phase 3: Usage Tracking & API Handler Tests

## Context Links
- [Main Plan](./plan.md)
- [Phase 1: Provider Core](./phase-01-provider-selector-retry-tests.md)
- [Phase 2: Auth & Security](./phase-02-access-oauth-tests.md)

## Overview
Implement unit tests for usage tracking (SQLite backend) and API handlers.
Focus on data integrity, aggregation correctness, and HTTP handler validation.

## Requirements
- 70% test coverage for usage tracking and API handlers
- Test database operations with in-memory SQLite
- Validate HTTP request/response handling
- Test concurrent database access

## Related Code Files

### 1. internal/usage/sqlite_backend.go (Lines 80-400)
```go
func (b *SQLiteBackend) LogRequest(ctx context.Context, req *Request) error
func (b *SQLiteBackend) QueryUsage(ctx context.Context, filter Filter) (*Report, error)
func (b *SQLiteBackend) aggregateUsage(rows *sql.Rows) (*Report, error)
// Critical: Accurate usage tracking
// Critical: Aggregation calculations
```

### 2. internal/api/handlers/chat.go (Lines 50-300)
```go
func (h *ChatHandler) Handle(c *gin.Context)
// Critical: Request validation
// Critical: Response formatting
// Critical: Error handling
```

### 3. internal/api/handlers/models.go (Lines 30-150)
```go
func (h *ModelsHandler) ListModels(c *gin.Context)
// Model availability filtering
```

## Implementation Steps

### Step 1: sqlite_backend_test.go (1.5h)
```go
func TestSQLiteBackend_LogRequest(t *testing.T) {
    // Setup in-memory database
    db := setupTestDB(t)
    defer db.Close()

    backend := NewSQLiteBackend(db)

    tests := []struct {
        name    string
        request *Request
        wantErr bool
    }{
        {
            name: "successful_log",
            request: &Request{
                UserID:       "user1",
                Provider:     "openai",
                Model:        "gpt-4",
                InputTokens:  100,
                OutputTokens: 200,
                Cost:         0.015,
                Timestamp:    time.Now(),
            },
            wantErr: false,
        },
        {
            name: "missing_required_fields",
            request: &Request{
                UserID: "user1",
                // Missing provider and model
            },
            wantErr: true,
        },
        {
            name: "concurrent_writes",
            request: &Request{
                UserID:   "user2",
                Provider: "claude",
                Model:    "claude-3",
            },
            wantErr: false,
        },
    }
}

func TestSQLiteBackend_QueryUsage(t *testing.T) {
    db := setupTestDB(t)
    backend := NewSQLiteBackend(db)

    // Seed test data
    seedTestUsageData(t, backend)

    tests := []struct {
        name           string
        filter         Filter
        expectedReport *Report
    }{
        {
            name: "filter_by_user",
            filter: Filter{
                UserID:    "user1",
                StartTime: time.Now().Add(-24 * time.Hour),
                EndTime:   time.Now(),
            },
            expectedReport: &Report{
                TotalRequests:     10,
                TotalInputTokens:  1000,
                TotalOutputTokens: 2000,
                TotalCost:        0.15,
            },
        },
        {
            name: "filter_by_provider",
            filter: Filter{
                Provider:  "openai",
                StartTime: time.Now().Add(-7 * 24 * time.Hour),
            },
            expectedReport: &Report{
                TotalRequests: 25,
                ByModel: map[string]*ModelUsage{
                    "gpt-4": {
                        Requests:     15,
                        InputTokens:  5000,
                        OutputTokens: 10000,
                    },
                    "gpt-3.5": {
                        Requests:     10,
                        InputTokens:  2000,
                        OutputTokens: 4000,
                    },
                },
            },
        },
    }
}

func TestSQLiteBackend_Aggregation(t *testing.T) {
    // Test aggregation accuracy
    // Test grouping by model/provider
    // Test time range queries
}

func TestSQLiteBackend_ConcurrentAccess(t *testing.T) {
    db := setupTestDB(t)
    backend := NewSQLiteBackend(db)

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // Concurrent writes
            err := backend.LogRequest(context.Background(), &Request{
                UserID:   fmt.Sprintf("user%d", id),
                Provider: "test",
                Model:    "test-model",
            })
            assert.NoError(t, err)

            // Concurrent reads
            _, err = backend.QueryUsage(context.Background(), Filter{
                UserID: fmt.Sprintf("user%d", id),
            })
            assert.NoError(t, err)
        }(i)
    }
    wg.Wait()
}

// Helper functions
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

    // Create schema
    _, err = db.Exec(createTableSQL)
    require.NoError(t, err)

    t.Cleanup(func() { db.Close() })
    return db
}
```

### Step 2: chat_handler_test.go (1h)
```go
func TestChatHandler_Handle(t *testing.T) {
    tests := []struct {
        name         string
        request      interface{}
        mockResponse interface{}
        wantStatus   int
        wantBody     string
    }{
        {
            name: "successful_chat_completion",
            request: map[string]interface{}{
                "model":    "gpt-4",
                "messages": []map[string]string{{"role": "user", "content": "Hello"}},
            },
            mockResponse: &ChatResponse{
                ID:      "chat-123",
                Choices: []Choice{{Message: Message{Content: "Hi there!"}}},
            },
            wantStatus: 200,
        },
        {
            name: "invalid_request_format",
            request: map[string]interface{}{
                "invalid": "data",
            },
            wantStatus: 400,
            wantBody:   "missing required field: model",
        },
        {
            name: "provider_error",
            request: map[string]interface{}{
                "model":    "gpt-4",
                "messages": []map[string]string{{"role": "user", "content": "Test"}},
            },
            mockResponse: errors.New("provider unavailable"),
            wantStatus:   503,
        },
        {
            name: "streaming_response",
            request: map[string]interface{}{
                "model":    "gpt-4",
                "messages": []map[string]string{{"role": "user", "content": "Stream"}},
                "stream":   true,
            },
            wantStatus: 200,
            // Verify SSE format
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            router := gin.New()
            handler := NewChatHandler(mockProvider)
            router.POST("/chat", handler.Handle)

            // Make request
            body, _ := json.Marshal(tt.request)
            req := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
            rec := httptest.NewRecorder()

            router.ServeHTTP(rec, req)

            // Assertions
            assert.Equal(t, tt.wantStatus, rec.Code)
            if tt.wantBody != "" {
                assert.Contains(t, rec.Body.String(), tt.wantBody)
            }
        })
    }
}

func TestChatHandler_Streaming(t *testing.T) {
    // Test SSE streaming
    // Test chunked responses
    // Test stream interruption
}
```

### Step 3: models_handler_test.go (0.5h)
```go
func TestModelsHandler_ListModels(t *testing.T) {
    tests := []struct {
        name           string
        availableModels []Model
        userAccess     []string
        wantModels     []string
    }{
        {
            name: "all_models_available",
            availableModels: []Model{
                {ID: "gpt-4", Provider: "openai"},
                {ID: "claude-3", Provider: "anthropic"},
            },
            userAccess: []string{"openai", "anthropic"},
            wantModels: []string{"gpt-4", "claude-3"},
        },
        {
            name: "filtered_by_access",
            availableModels: []Model{
                {ID: "gpt-4", Provider: "openai"},
                {ID: "claude-3", Provider: "anthropic"},
            },
            userAccess: []string{"openai"},
            wantModels: []string{"gpt-4"},
        },
    }
}

func BenchmarkModelsHandler_ListModels(b *testing.B) {
    // Benchmark with 1000+ models
}
```

## Test Data Setup
```go
// Helper functions
func seedTestUsageData(t *testing.T, backend *SQLiteBackend)
func mockProvider(response interface{}, err error) Provider
func newTestRouter(handlers ...gin.HandlerFunc) *gin.Engine
```

## Success Criteria
- [ ] 70% coverage for usage/sqlite_backend.go
- [ ] 70% coverage for api/handlers/chat.go
- [ ] 70% coverage for api/handlers/models.go
- [ ] Database operations are atomic and consistent
- [ ] HTTP handlers validate all inputs

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| Database locking | Medium | Use WAL mode, test concurrency |
| HTTP mocking complexity | Medium | Use httptest package |
| Streaming tests | Low | Mock SSE responses |
| Large dataset tests | Low | Use table-driven tests with limits |

## Validation Commands
```bash
# Run phase 3 tests
go test -v -race ./internal/usage/... ./internal/api/...

# Database-specific tests
go test -v -tags=sqlite ./internal/usage/...

# API integration tests
go test -v -tags=integration ./internal/api/...

# Combined coverage
go test -coverprofile=phase3.out ./internal/usage/... ./internal/api/...
go tool cover -func=phase3.out | grep total
```

## Integration Points
- Phase 1 provider selection feeds into API handlers
- Phase 2 auth validates API requests
- Usage tracking records all API interactions

## Next Steps
After Phase 3 completion:
1. Run full test suite with all phases
2. Generate combined coverage report
3. Identify remaining gaps
4. Document test maintenance procedures
package provider

import (
	"context"
	"testing"
	"time"
)

func newTestState(tokensUsed, activeRequests int64) *AuthQuotaState {
	state := &AuthQuotaState{}
	state.TotalTokensUsed.Store(tokensUsed)
	state.ActiveRequests.Store(activeRequests)
	return state
}

func TestDefaultStrategy_Score(t *testing.T) {
	strategy := &DefaultStrategy{}

	config := &ProviderQuotaConfig{
		Provider:       "test",
		EstimatedLimit: 500_000,
	}

	tests := []struct {
		name           string
		state          *AuthQuotaState
		expectedBetter bool
	}{
		{
			name:           "no state has zero priority",
			state:          nil,
			expectedBetter: true,
		},
		{
			name:           "low usage has lower priority",
			state:          newTestState(100_000, 0),
			expectedBetter: true,
		},
		{
			name:           "high usage has higher priority",
			state:          newTestState(400_000, 0),
			expectedBetter: false,
		},
	}

	baseState := newTestState(200_000, 0)
	basePriority := strategy.Score(nil, baseState, config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := strategy.Score(nil, tt.state, config)

			if tt.expectedBetter && priority >= basePriority {
				t.Errorf("expected priority %d to be lower than base %d", priority, basePriority)
			}
			if !tt.expectedBetter && priority <= basePriority {
				t.Errorf("expected priority %d to be higher than base %d", priority, basePriority)
			}
		})
	}
}

func TestDefaultStrategy_Score_ActiveRequestsPenalty(t *testing.T) {
	strategy := &DefaultStrategy{}

	config := &ProviderQuotaConfig{
		Provider:       "test",
		EstimatedLimit: 500_000,
	}

	idleState := newTestState(100_000, 0)
	busyState := newTestState(100_000, 3)

	idlePriority := strategy.Score(nil, idleState, config)
	busyPriority := strategy.Score(nil, busyState, config)

	if busyPriority <= idlePriority {
		t.Errorf("busy priority %d should be higher than idle %d", busyPriority, idlePriority)
	}

	expectedMinPenalty := int64(3 * 1000)
	if busyPriority-idlePriority < expectedMinPenalty {
		t.Errorf("expected at least %d penalty for 3 active requests, got %d", expectedMinPenalty, busyPriority-idlePriority)
	}
}

func TestQuotaManager_Pick_SelectsLeastUsed(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "antigravity"}
	auth2 := &Auth{ID: "auth2", Provider: "antigravity"}
	auth3 := &Auth{ID: "auth3", Provider: "antigravity"}

	m.RecordRequestEnd("auth1", "antigravity", 1_000_000, false)
	m.RecordRequestEnd("auth2", "antigravity", 500_000, false)
	m.RecordRequestEnd("auth3", "antigravity", 10_000, false)

	selected, err := m.Pick(context.Background(), "antigravity", "claude-sonnet-4", Options{ForceRotate: true}, []*Auth{auth1, auth2, auth3})
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	if selected.ID != "auth3" {
		t.Errorf("expected auth3 (least usage), got %s", selected.ID)
	}
}

func TestQuotaManager_Pick_StickyBehavior(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "claude"}
	auth2 := &Auth{ID: "auth2", Provider: "claude"}

	selected1, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{}, []*Auth{auth1, auth2})
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	selected2, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{}, []*Auth{auth1, auth2})
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	if selected1.ID != selected2.ID {
		t.Errorf("sticky should return same auth: got %s then %s", selected1.ID, selected2.ID)
	}
}

func TestQuotaManager_Pick_NoStickyForAntigravity(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "antigravity"}
	auth2 := &Auth{ID: "auth2", Provider: "antigravity"}

	m.RecordRequestEnd("auth1", "antigravity", 400_000, false)

	selected, err := m.Pick(context.Background(), "antigravity", "gemini-2.5-pro", Options{ForceRotate: true}, []*Auth{auth1, auth2})
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	if selected.ID != "auth2" {
		t.Errorf("expected auth2 (less usage, no sticky), got %s", selected.ID)
	}
}

func TestQuotaManager_RecordQuotaHit_SetsCooldown(t *testing.T) {
	m := NewQuotaManager()

	m.RecordRequestEnd("auth1", "antigravity", 600_000, false)

	resetAfter := 30 * time.Minute
	m.RecordQuotaHit("auth1", "antigravity", "claude-sonnet-4", &resetAfter)

	state := m.GetState("auth1")
	if state == nil {
		t.Fatal("expected state to exist")
	}

	if state.CooldownUntil.IsZero() {
		t.Error("expected CooldownUntil to be set")
	}
}

func TestQuotaManager_RecordRequestStartEnd(t *testing.T) {
	m := NewQuotaManager()

	m.RecordRequestStart("auth1")

	state := m.GetState("auth1")
	if state == nil || state.ActiveRequests != 1 {
		t.Error("expected ActiveRequests to be 1 after RecordRequestStart")
	}

	m.RecordRequestEnd("auth1", "antigravity", 1000, false)

	state = m.GetState("auth1")
	if state == nil || state.ActiveRequests != 0 {
		t.Error("expected ActiveRequests to be 0 after RecordRequestEnd")
	}
	if state.TotalTokensUsed != 1000 {
		t.Errorf("expected TotalTokensUsed to be 1000, got %d", state.TotalTokensUsed)
	}
}

func TestQuotaManager_Pick_SkipsCooldown(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "antigravity"}
	auth2 := &Auth{ID: "auth2", Provider: "antigravity"}

	cooldown := 1 * time.Hour
	m.RecordQuotaHit("auth1", "antigravity", "test", &cooldown)

	selected, err := m.Pick(context.Background(), "antigravity", "test", Options{ForceRotate: true}, []*Auth{auth1, auth2})
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	if selected.ID != "auth2" {
		t.Errorf("expected auth2 (auth1 in cooldown), got %s", selected.ID)
	}
}

func TestQuotaManager_Pick_AllExhaustedReturnsError(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "antigravity"}

	cooldown := 1 * time.Hour
	m.RecordQuotaHit("auth1", "antigravity", "test", &cooldown)

	_, err := m.Pick(context.Background(), "antigravity", "test", Options{}, []*Auth{auth1})
	if err == nil {
		t.Fatal("expected error when all auths exhausted")
	}

	provErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if provErr.HTTPStatus != 429 {
		t.Errorf("expected 429 status, got %d", provErr.HTTPStatus)
	}
}

func TestQuotaManager_RecordRequestEnd_ClearsCooldownOnSuccess(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "auth1", Provider: "antigravity"}

	cooldown := 1 * time.Hour
	m.RecordQuotaHit("auth1", "antigravity", "test", &cooldown)

	state := m.GetState("auth1")
	if state == nil || state.CooldownUntil.IsZero() {
		t.Fatal("expected CooldownUntil to be set after RecordQuotaHit")
	}

	m.RecordRequestEnd("auth1", "antigravity", 1000, false)

	state = m.GetState("auth1")
	if state == nil {
		t.Fatal("expected state to exist")
	}
	if !state.CooldownUntil.IsZero() {
		t.Error("expected CooldownUntil to be cleared after successful RecordRequestEnd")
	}

	selected, err := m.Pick(context.Background(), "antigravity", "test", Options{ForceRotate: true}, []*Auth{auth1})
	if err != nil {
		t.Fatalf("Pick failed after cooldown cleared: %v", err)
	}
	if selected.ID != "auth1" {
		t.Errorf("expected auth1 after cooldown cleared, got %s", selected.ID)
	}
}

func TestQuotaConfig_GetProviderQuotaConfig(t *testing.T) {
	tests := []struct {
		provider     string
		expectSticky bool
	}{
		{"antigravity", false},
		{"claude", true},
		{"copilot", true},
		{"gemini", true},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			cfg := GetProviderQuotaConfig(tt.provider)
			if cfg.StickyEnabled != tt.expectSticky {
				t.Errorf("expected sticky %v, got %v", tt.expectSticky, cfg.StickyEnabled)
			}
		})
	}
}

func TestQuotaManager_GetStrategy(t *testing.T) {
	m := NewQuotaManager()

	tests := []struct {
		provider     string
		expectedType string
	}{
		{"antigravity", "*provider.AntigravityStrategy"},
		{"claude", "*provider.ClaudeStrategy"},
		{"copilot", "*provider.CopilotStrategy"},
		{"gemini", "*provider.GeminiStrategy"},
		{"unknown", "*provider.DefaultStrategy"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			strategy := m.getStrategy(tt.provider)
			strategyType := ""
			switch strategy.(type) {
			case *AntigravityStrategy:
				strategyType = "*provider.AntigravityStrategy"
			case *ClaudeStrategy:
				strategyType = "*provider.ClaudeStrategy"
			case *CopilotStrategy:
				strategyType = "*provider.CopilotStrategy"
			case *GeminiStrategy:
				strategyType = "*provider.GeminiStrategy"
			case *DefaultStrategy:
				strategyType = "*provider.DefaultStrategy"
			}
			if strategyType != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, strategyType)
			}
		})
	}
}

func TestAntigravityStrategyTokenPenalty(t *testing.T) {
	strategy := &AntigravityStrategy{}

	readyAuth := &Auth{
		Metadata: map[string]any{
			"access_token": "valid",
			"expired":      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
	}

	expiredAuth := &Auth{
		Metadata: map[string]any{
			"access_token": "expired",
			"expired":      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	needsRefreshAuth := &Auth{
		Metadata: map[string]any{
			"access_token": "needs-refresh",
			"expired":      time.Now().Add(3 * time.Minute).Format(time.RFC3339),
		},
	}

	readyScore := strategy.Score(readyAuth, &AuthQuotaState{}, nil)
	expiredScore := strategy.Score(expiredAuth, &AuthQuotaState{}, nil)
	needsRefreshScore := strategy.Score(needsRefreshAuth, &AuthQuotaState{}, nil)

	if expiredScore <= readyScore {
		t.Errorf("expired auth should have higher priority (selected last): expired=%d, ready=%d", expiredScore, readyScore)
	}

	if needsRefreshScore <= readyScore {
		t.Errorf("needs-refresh auth should have higher priority than ready: needsRefresh=%d, ready=%d", needsRefreshScore, readyScore)
	}

	if expiredScore <= needsRefreshScore {
		t.Errorf("expired auth should have higher priority than needs-refresh: expired=%d, needsRefresh=%d", expiredScore, needsRefreshScore)
	}

	expectedExpiredPenalty := int64(10000)
	if expiredScore < expectedExpiredPenalty {
		t.Errorf("expected expired penalty of at least %d, got %d", expectedExpiredPenalty, expiredScore)
	}

	expectedRefreshPenalty := int64(500)
	if needsRefreshScore < expectedRefreshPenalty {
		t.Errorf("expected needs-refresh penalty of at least %d, got %d", expectedRefreshPenalty, needsRefreshScore)
	}
}

// ========== Round-Robin Validation Tests ==========

// TestQuotaManager_Pick_RoundRobinDistribution verifies that requests are distributed
// across multiple accounts when ForceRotate is used (bypasses sticky sessions).
func TestQuotaManager_Pick_RoundRobinDistribution(t *testing.T) {
	m := NewQuotaManager()

	// Create 3 Claude accounts with identical state (no usage)
	auth1 := &Auth{ID: "rr-auth1", Provider: "claude"}
	auth2 := &Auth{ID: "rr-auth2", Provider: "claude"}
	auth3 := &Auth{ID: "rr-auth3", Provider: "claude"}
	auths := []*Auth{auth1, auth2, auth3}

	// Track how many times each account is selected
	selections := make(map[string]int)

	// Pick 9 times with ForceRotate to ensure round-robin distribution
	for i := 0; i < 9; i++ {
		selected, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{ForceRotate: true}, auths)
		if err != nil {
			t.Fatalf("Pick %d failed: %v", i, err)
		}
		selections[selected.ID]++
	}

	// With 9 picks and 3 accounts, each should be selected at least once
	// Due to random tie-breaking among top-3 similar scores, exact distribution varies
	for _, auth := range auths {
		if selections[auth.ID] == 0 {
			t.Errorf("Account %s was never selected in round-robin distribution", auth.ID)
		}
	}

	t.Logf("Round-robin distribution: auth1=%d, auth2=%d, auth3=%d",
		selections["rr-auth1"], selections["rr-auth2"], selections["rr-auth3"])
}

// TestQuotaManager_Pick_TwoAccountRoundRobin verifies round-robin between exactly 2 accounts.
func TestQuotaManager_Pick_TwoAccountRoundRobin(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "two-auth1", Provider: "claude"}
	auth2 := &Auth{ID: "two-auth2", Provider: "claude"}
	auths := []*Auth{auth1, auth2}

	// Track selections
	selections := make(map[string]int)

	// Pick 6 times with ForceRotate
	for i := 0; i < 6; i++ {
		selected, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{ForceRotate: true}, auths)
		if err != nil {
			t.Fatalf("Pick %d failed: %v", i, err)
		}
		selections[selected.ID]++
	}

	// Both accounts should be selected
	if selections["two-auth1"] == 0 {
		t.Error("Account two-auth1 was never selected")
	}
	if selections["two-auth2"] == 0 {
		t.Error("Account two-auth2 was never selected")
	}

	t.Logf("Two-account distribution: auth1=%d, auth2=%d",
		selections["two-auth1"], selections["two-auth2"])
}

// ========== 429 Fallback Validation Tests ==========

// TestQuotaManager_Claude429FallbackToOtherAccount verifies that when a Claude account
// receives a 429 rate limit, the next request is routed to a different account.
func TestQuotaManager_Claude429FallbackToOtherAccount(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "claude-auth1", Provider: "claude"}
	auth2 := &Auth{ID: "claude-auth2", Provider: "claude"}
	auths := []*Auth{auth1, auth2}

	// First pick should work (either account)
	firstPick, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{ForceRotate: true}, auths)
	if err != nil {
		t.Fatalf("First Pick failed: %v", err)
	}

	// Simulate 429 rate limit on the first picked account (3 hour cooldown)
	cooldown := 3 * time.Hour
	m.RecordQuotaHit(firstPick.ID, "claude", "claude-sonnet-4", &cooldown)

	// Verify the account is now in cooldown
	state := m.GetState(firstPick.ID)
	if state == nil || state.CooldownUntil.IsZero() {
		t.Fatal("Expected account to be in cooldown after 429")
	}

	// Next pick should route to the OTHER account (not the one in cooldown)
	secondPick, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{ForceRotate: true}, auths)
	if err != nil {
		t.Fatalf("Second Pick failed: %v", err)
	}

	// The second pick should NOT be the account that's in cooldown
	if secondPick.ID == firstPick.ID {
		t.Errorf("Expected request to be routed to different account after 429. First: %s, Second: %s",
			firstPick.ID, secondPick.ID)
	}

	t.Logf("429 fallback working: First pick=%s (now in cooldown), Second pick=%s", firstPick.ID, secondPick.ID)
}

// TestQuotaManager_AllClaude429ReturnsError verifies that when ALL Claude accounts
// are rate limited, an appropriate error is returned.
func TestQuotaManager_AllClaude429ReturnsError(t *testing.T) {
	m := NewQuotaManager()

	auth1 := &Auth{ID: "all429-auth1", Provider: "claude"}
	auth2 := &Auth{ID: "all429-auth2", Provider: "claude"}
	auths := []*Auth{auth1, auth2}

	// Put both accounts in cooldown (simulating 429 on both)
	cooldown := 3 * time.Hour
	m.RecordQuotaHit("all429-auth1", "claude", "claude-sonnet-4", &cooldown)
	m.RecordQuotaHit("all429-auth2", "claude", "claude-sonnet-4", &cooldown)

	// Attempt to pick - should fail with 429-like error
	_, err := m.Pick(context.Background(), "claude", "claude-sonnet-4", Options{}, auths)
	if err == nil {
		t.Fatal("Expected error when all accounts are in cooldown")
	}

	// Verify it's a 429 error with retry info
	provErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T: %v", err, err)
	}
	if provErr.HTTPStatus != 429 {
		t.Errorf("Expected 429 status, got %d", provErr.HTTPStatus)
	}

	t.Logf("All accounts in cooldown correctly returns 429: %v", err)
}

// ========== Claude Strategy 5-Hour Cooldown Tests ==========

// TestClaudeStrategy_OnQuotaHit_DefaultCooldown5Hours verifies that when no explicit
// cooldown is provided, Claude accounts default to 5 hours.
func TestClaudeStrategy_OnQuotaHit_DefaultCooldown5Hours(t *testing.T) {
	strategy := &ClaudeStrategy{}
	state := &AuthQuotaState{}

	now := time.Now()
	strategy.OnQuotaHit(state, nil)

	cooldownUntil := state.GetCooldownUntil()
	if cooldownUntil.IsZero() {
		t.Fatal("Expected CooldownUntil to be set")
	}

	actualCooldown := cooldownUntil.Sub(now)

	expected := 5 * time.Hour
	tolerance := 1 * time.Second

	if actualCooldown < expected-tolerance || actualCooldown > expected+tolerance {
		t.Errorf("Expected cooldown of ~%v, got %v", expected, actualCooldown)
	}

	t.Logf("Claude default cooldown verified: %v (expected ~%v)", actualCooldown.Round(time.Minute), expected)
}

// TestClaudeStrategy_OnQuotaHit_ExplicitCooldownOverridesDefault verifies that
// when an explicit cooldown is provided (from server's Retry-After), it's used.
func TestClaudeStrategy_OnQuotaHit_ExplicitCooldownOverridesDefault(t *testing.T) {
	strategy := &ClaudeStrategy{}
	state := &AuthQuotaState{}

	// Call OnQuotaHit with explicit 30-minute cooldown
	now := time.Now()
	explicitCooldown := 30 * time.Minute
	strategy.OnQuotaHit(state, &explicitCooldown)

	cooldownUntil := state.GetCooldownUntil()
	if cooldownUntil.IsZero() {
		t.Fatal("Expected CooldownUntil to be set")
	}

	// Calculate the actual cooldown duration
	actualCooldown := cooldownUntil.Sub(now)

	// Should be approximately 30 minutes
	tolerance := 1 * time.Second

	if actualCooldown < explicitCooldown-tolerance || actualCooldown > explicitCooldown+tolerance {
		t.Errorf("Expected cooldown of ~%v, got %v", explicitCooldown, actualCooldown)
	}

	t.Logf("Claude explicit cooldown verified: %v (expected ~%v)", actualCooldown.Round(time.Second), explicitCooldown)
}

// TestClaudeStrategy_OnQuotaHit_LearnedCooldownUsed verifies that when a previous
// cooldown was learned, it's used as the default instead of the 3-hour fallback.
func TestClaudeStrategy_OnQuotaHit_LearnedCooldownUsed(t *testing.T) {
	strategy := &ClaudeStrategy{}
	state := &AuthQuotaState{}

	// First, set a learned cooldown (simulating a previous 429 with explicit duration)
	firstCooldown := 2 * time.Hour
	strategy.OnQuotaHit(state, &firstCooldown)

	// Verify learned cooldown was stored
	learnedCooldown := state.GetLearnedCooldown()
	if learnedCooldown != firstCooldown {
		t.Errorf("Expected learned cooldown %v, got %v", firstCooldown, learnedCooldown)
	}

	// Clear the cooldown state to test the learned value
	state.SetCooldownUntil(time.Time{})

	// Now call OnQuotaHit with nil (should use learned value, not default)
	now := time.Now()
	strategy.OnQuotaHit(state, nil)

	cooldownUntil := state.GetCooldownUntil()
	actualCooldown := cooldownUntil.Sub(now)

	// Should use the learned 2-hour value, not the default 3 hours
	tolerance := 1 * time.Second

	if actualCooldown < firstCooldown-tolerance || actualCooldown > firstCooldown+tolerance {
		t.Errorf("Expected learned cooldown of ~%v, got %v (should not use default 3h)", firstCooldown, actualCooldown)
	}

	t.Logf("Claude learned cooldown verified: %v (expected ~%v)", actualCooldown.Round(time.Second), firstCooldown)
}

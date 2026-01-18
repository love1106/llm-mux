package provider

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStickyStoreSharding(t *testing.T) {
	store := NewStickyStore()
	store.Start()
	defer store.Stop()

	for i := 0; i < 100; i++ {
		key := "provider:" + string(rune('a'+i/10)) + string(rune('0'+i%10))
		store.Set(key, "auth"+string(rune('0'+i%10)))
	}

	if store.Len() != 100 {
		t.Errorf("Expected 100 entries, got %d", store.Len())
	}
}

func TestStickyStoreEviction(t *testing.T) {
	store := NewStickyStore()
	store.Start()
	defer store.Stop()

	shard := store.shards[0]

	for i := 0; i < maxEntriesPerShard+10; i++ {
		shard.mu.Lock()
		key := "key" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		shard.entries[key] = &stickyEntry{
			authID:   "auth",
			lastUsed: time.Now(),
		}
		if len(shard.entries) >= maxEntriesPerShard {
			store.evictOldest(shard, time.Now())
		}
		shard.mu.Unlock()
	}

	shard.mu.RLock()
	count := len(shard.entries)
	shard.mu.RUnlock()

	if count > maxEntriesPerShard {
		t.Errorf("Expected <= %d entries, got %d", maxEntriesPerShard, count)
	}
}

func TestConcurrentPick(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini"},
		{ID: "auth2", Provider: "gemini"},
		{ID: "auth3", Provider: "gemini"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)
				if err != nil {
					t.Errorf("Pick failed: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestForceRotate(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini"},
		{ID: "auth2", Provider: "gemini"},
	}

	first, _ := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)

	sticky, _ := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)
	if sticky.ID != first.ID {
		t.Errorf("Expected sticky session to return same auth, got %s vs %s", sticky.ID, first.ID)
	}

	rotated, _ := selector.Pick(context.Background(), "gemini", "model", Options{ForceRotate: true}, auths)
	if rotated.ID == first.ID {
		t.Error("Expected ForceRotate to select different auth")
	}
}

func TestGracefulShutdown(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()

	auths := []*Auth{{ID: "auth1", Provider: "gemini"}}
	_, err := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		selector.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not complete within timeout")
	}
}

func BenchmarkPick(b *testing.B) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini"},
		{ID: "auth2", Provider: "gemini"},
		{ID: "auth3", Provider: "gemini"},
	}

	ctx := context.Background()
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.Pick(ctx, "gemini", "model", opts, auths)
	}
}

func BenchmarkPickParallel(b *testing.B) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini"},
		{ID: "auth2", Provider: "gemini"},
		{ID: "auth3", Provider: "gemini"},
	}

	ctx := context.Background()
	opts := Options{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = selector.Pick(ctx, "gemini", "model", opts, auths)
		}
	})
}

func TestRoundRobinSelector_Pick_EmptyAuthList(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	_, err := selector.Pick(context.Background(), "gemini", "model", Options{}, []*Auth{})
	if err == nil {
		t.Fatal("expected error for empty auth list")
	}
	provErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if provErr.Code != "auth_not_found" {
		t.Errorf("expected error code 'auth_not_found', got '%s'", provErr.Code)
	}
}

func TestRoundRobinSelector_Pick_AllDisabled(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini", Disabled: true},
		{ID: "auth2", Provider: "gemini", Disabled: true},
	}

	_, err := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)
	if err == nil {
		t.Fatal("expected error when all auths disabled")
	}
}

func TestRoundRobinSelector_Pick_SkipsDisabled(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini", Disabled: true},
		{ID: "auth2", Provider: "gemini", Disabled: false},
		{ID: "auth3", Provider: "gemini", Disabled: true},
	}

	selected, err := selector.Pick(context.Background(), "gemini", "model", Options{ForceRotate: true}, auths)
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}
	if selected.ID != "auth2" {
		t.Errorf("expected auth2 (only non-disabled), got %s", selected.ID)
	}
}

func TestRoundRobinSelector_Pick_AllInCooldown(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	nextRetry := time.Now().Add(1 * time.Hour)
	auths := []*Auth{
		{
			ID:       "auth1",
			Provider: "gemini",
			ModelStates: map[string]*ModelState{
				"model": {
					Unavailable:    true,
					NextRetryAfter: nextRetry,
					Quota:          QuotaState{Exceeded: true, NextRecoverAt: nextRetry},
				},
			},
		},
		{
			ID:       "auth2",
			Provider: "gemini",
			ModelStates: map[string]*ModelState{
				"model": {
					Unavailable:    true,
					NextRetryAfter: nextRetry,
					Quota:          QuotaState{Exceeded: true, NextRecoverAt: nextRetry},
				},
			},
		},
	}

	_, err := selector.Pick(context.Background(), "gemini", "model", Options{}, auths)
	if err == nil {
		t.Fatal("expected error when all auths in cooldown")
	}
	cooldownErr, ok := err.(*modelCooldownError)
	if !ok {
		t.Fatalf("expected *modelCooldownError, got %T", err)
	}
	if cooldownErr.model != "model" {
		t.Errorf("expected model 'model', got '%s'", cooldownErr.model)
	}
}

func TestRoundRobinSelector_Pick_ModelLevelBlocking(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	nextRetry := time.Now().Add(1 * time.Hour)
	auths := []*Auth{
		{
			ID:       "auth1",
			Provider: "gemini",
			ModelStates: map[string]*ModelState{
				"model-a": {
					Unavailable:    true,
					NextRetryAfter: nextRetry,
					Quota:          QuotaState{Exceeded: true, NextRecoverAt: nextRetry},
				},
			},
		},
		{ID: "auth2", Provider: "gemini"},
	}

	selected, err := selector.Pick(context.Background(), "gemini", "model-a", Options{ForceRotate: true}, auths)
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}
	if selected.ID != "auth2" {
		t.Errorf("expected auth2 (auth1 blocked for model-a), got %s", selected.ID)
	}
}

func TestRoundRobinSelector_Pick_DisabledStatusSkipped(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "auth1", Provider: "gemini", Status: StatusDisabled},
		{ID: "auth2", Provider: "gemini", Status: StatusActive},
	}

	selected, err := selector.Pick(context.Background(), "gemini", "model", Options{ForceRotate: true}, auths)
	if err != nil {
		t.Fatalf("Pick failed: %v", err)
	}
	if selected.ID != "auth2" {
		t.Errorf("expected auth2 (auth1 has StatusDisabled), got %s", selected.ID)
	}
}

func TestIsAuthBlockedForModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		auth       *Auth
		model      string
		wantBlocked bool
		wantReason blockReason
	}{
		{
			name:        "nil auth is blocked",
			auth:        nil,
			model:       "test",
			wantBlocked: true,
			wantReason:  blockReasonOther,
		},
		{
			name:        "disabled auth is blocked",
			auth:        &Auth{ID: "a", Disabled: true},
			model:       "test",
			wantBlocked: true,
			wantReason:  blockReasonDisabled,
		},
		{
			name:        "status disabled is blocked",
			auth:        &Auth{ID: "a", Status: StatusDisabled},
			model:       "test",
			wantBlocked: true,
			wantReason:  blockReasonDisabled,
		},
		{
			name:        "normal auth is not blocked",
			auth:        &Auth{ID: "a", Status: StatusActive},
			model:       "test",
			wantBlocked: false,
			wantReason:  blockReasonNone,
		},
		{
			name: "model-level disabled is blocked",
			auth: &Auth{
				ID: "a",
				ModelStates: map[string]*ModelState{
					"blocked-model": {Status: StatusDisabled},
				},
			},
			model:       "blocked-model",
			wantBlocked: true,
			wantReason:  blockReasonDisabled,
		},
		{
			name: "model-level cooldown is blocked",
			auth: &Auth{
				ID: "a",
				ModelStates: map[string]*ModelState{
					"quota-model": {
						Unavailable:    true,
						NextRetryAfter: now.Add(1 * time.Hour),
						Quota:          QuotaState{Exceeded: true},
					},
				},
			},
			model:       "quota-model",
			wantBlocked: true,
			wantReason:  blockReasonCooldown,
		},
		{
			name: "expired cooldown is not blocked",
			auth: &Auth{
				ID: "a",
				ModelStates: map[string]*ModelState{
					"expired-model": {
						Unavailable:    true,
						NextRetryAfter: now.Add(-1 * time.Hour),
					},
				},
			},
			model:       "expired-model",
			wantBlocked: false,
			wantReason:  blockReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, reason, _ := isAuthBlockedForModel(tt.auth, tt.model, now)
			if blocked != tt.wantBlocked {
				t.Errorf("blocked = %v, want %v", blocked, tt.wantBlocked)
			}
			if reason != tt.wantReason {
				t.Errorf("reason = %v, want %v", reason, tt.wantReason)
			}
		})
	}
}

func TestModelCooldownError(t *testing.T) {
	err := newModelCooldownError("claude-sonnet", "anthropic", 30*time.Second)

	if err.StatusCode() != 429 {
		t.Errorf("expected status 429, got %d", err.StatusCode())
	}

	headers := err.Headers()
	if headers.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", headers.Get("Content-Type"))
	}
	if headers.Get("Retry-After") != "30" {
		t.Errorf("expected Retry-After 30, got %s", headers.Get("Retry-After"))
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
}

func TestRoundRobinSelector_RoundRobinOrder(t *testing.T) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := []*Auth{
		{ID: "a", Provider: "test"},
		{ID: "b", Provider: "test"},
		{ID: "c", Provider: "test"},
	}

	first, _ := selector.Pick(context.Background(), "test", "model", Options{}, auths)
	_ = first

	results := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		selected, err := selector.Pick(context.Background(), "test", "model", Options{ForceRotate: true}, auths)
		if err != nil {
			t.Fatalf("Pick failed at iteration %d: %v", i, err)
		}
		results = append(results, selected.ID)
	}

	for i := 0; i < 3; i++ {
		if results[i] != results[i+3] {
			t.Errorf("expected round-robin pattern, got %v", results)
			break
		}
	}
}

func BenchmarkPick_LargeAuthList(b *testing.B) {
	selector := &RoundRobinSelector{}
	selector.Start()
	defer selector.Stop()

	auths := make([]*Auth, 1000)
	for i := 0; i < 1000; i++ {
		auths[i] = &Auth{ID: "auth" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)), Provider: "gemini"}
	}

	ctx := context.Background()
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.Pick(ctx, "gemini", "model", opts, auths)
	}
}

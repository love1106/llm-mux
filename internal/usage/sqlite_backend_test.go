package usage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestSQLiteBackend(t *testing.T) *SQLiteBackend {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_usage.db")
	b, err := NewSQLiteBackend(dbPath, BackendConfig{
		BatchSize:     10,
		FlushInterval: time.Second,
		RetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("NewSQLiteBackend: %v", err)
	}
	t.Cleanup(func() {
		_ = b.Stop()
		_ = os.Remove(dbPath)
	})
	if err := b.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return b
}

func seedRecords(t *testing.T, b *SQLiteBackend, records []UsageRecord) {
	t.Helper()
	ctx := context.Background()
	if err := b.writeBatch(ctx, records); err != nil {
		t.Fatalf("writeBatch: %v", err)
	}
}

func TestQueryAPIKeyStats_Basic(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "sk-key-alpha", RequestedAt: now, InputTokens: 100, OutputTokens: 50, TotalTokens: 150},
		{Provider: "claude", Model: "sonnet-4", APIKey: "sk-key-alpha", RequestedAt: now, InputTokens: 200, OutputTokens: 100, TotalTokens: 300, Failed: true},
		{Provider: "openai", Model: "gpt-4o", APIKey: "sk-key-beta", RequestedAt: now, InputTokens: 500, OutputTokens: 250, TotalTokens: 750},
	})

	ctx := context.Background()
	stats, err := b.QueryAPIKeyStats(ctx, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 API key groups, got %d", len(stats))
	}

	// Ordered by total_tokens DESC: beta (750) > alpha (450)
	beta := stats[0]
	alpha := stats[1]

	if beta.APIKey != "sk-key-beta" {
		t.Errorf("expected first key sk-key-beta, got %s", beta.APIKey)
	}
	if beta.Requests != 1 || beta.SuccessCount != 1 || beta.FailureCount != 0 {
		t.Errorf("beta: requests=%d success=%d failure=%d", beta.Requests, beta.SuccessCount, beta.FailureCount)
	}
	if beta.TotalTokens != 750 {
		t.Errorf("beta total_tokens: got %d, want 750", beta.TotalTokens)
	}

	if alpha.APIKey != "sk-key-alpha" {
		t.Errorf("expected second key sk-key-alpha, got %s", alpha.APIKey)
	}
	if alpha.Requests != 2 || alpha.SuccessCount != 1 || alpha.FailureCount != 1 {
		t.Errorf("alpha: requests=%d success=%d failure=%d", alpha.Requests, alpha.SuccessCount, alpha.FailureCount)
	}
	if alpha.TotalTokens != 450 {
		t.Errorf("alpha total_tokens: got %d, want 450", alpha.TotalTokens)
	}
}

func TestQueryAPIKeyStats_Models(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "sk-abc", RequestedAt: now, TotalTokens: 100},
		{Provider: "claude", Model: "sonnet-4", APIKey: "sk-abc", RequestedAt: now, TotalTokens: 200},
		{Provider: "claude", Model: "opus-4", APIKey: "sk-abc", RequestedAt: now, TotalTokens: 50},
	})

	stats, err := b.QueryAPIKeyStats(context.Background(), now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 group, got %d", len(stats))
	}

	// Should have 2 distinct models
	if len(stats[0].Models) != 2 {
		t.Errorf("expected 2 distinct models, got %v", stats[0].Models)
	}
}

func TestQueryAPIKeyStats_EmptyKeyBecomesUnknown(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "", RequestedAt: now, TotalTokens: 100},
	})

	stats, err := b.QueryAPIKeyStats(context.Background(), now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 group, got %d", len(stats))
	}
	if stats[0].APIKey != "unknown" {
		t.Errorf("expected 'unknown' for empty key, got %q", stats[0].APIKey)
	}
}

func TestQueryAPIKeyStats_SinceFilter(t *testing.T) {
	b := newTestSQLiteBackend(t)
	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "old-key", RequestedAt: old, TotalTokens: 100},
		{Provider: "claude", Model: "opus-4", APIKey: "new-key", RequestedAt: recent, TotalTokens: 200},
	})

	// Query only last 24h
	stats, err := b.QueryAPIKeyStats(context.Background(), time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 group (only recent), got %d", len(stats))
	}
	if stats[0].APIKey != "new-key" {
		t.Errorf("expected new-key, got %s", stats[0].APIKey)
	}
}

func TestQueryAPIKeyStats_LastSeenAt(t *testing.T) {
	b := newTestSQLiteBackend(t)
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "sk-time", RequestedAt: t1, TotalTokens: 100},
		{Provider: "claude", Model: "opus-4", APIKey: "sk-time", RequestedAt: t2, TotalTokens: 200},
	})

	stats, err := b.QueryAPIKeyStats(context.Background(), t1.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 group, got %d", len(stats))
	}
	if stats[0].LastSeenAt.IsZero() {
		t.Error("expected non-zero LastSeenAt")
	}
	// LastSeenAt should be close to t2 (the more recent record)
	diff := stats[0].LastSeenAt.Sub(t2)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("LastSeenAt %v not close to expected %v", stats[0].LastSeenAt, t2)
	}
}

func TestQueryAPIKeyStats_CacheTokens(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "sk-cache", RequestedAt: now, TotalTokens: 500, CacheCreationInputTokens: 100, CacheReadInputTokens: 200},
		{Provider: "claude", Model: "opus-4", APIKey: "sk-cache", RequestedAt: now, TotalTokens: 300, CacheCreationInputTokens: 50, CacheReadInputTokens: 150},
	})

	stats, err := b.QueryAPIKeyStats(context.Background(), now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 group, got %d", len(stats))
	}
	if stats[0].CacheCreationInputTokens != 150 {
		t.Errorf("cache_creation: got %d, want 150", stats[0].CacheCreationInputTokens)
	}
	if stats[0].CacheReadInputTokens != 350 {
		t.Errorf("cache_read: got %d, want 350", stats[0].CacheReadInputTokens)
	}
}

func TestQueryAPIKeyStats_Empty(t *testing.T) {
	b := newTestSQLiteBackend(t)

	stats, err := b.QueryAPIKeyStats(context.Background(), time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 results for empty DB, got %d", len(stats))
	}
}

func TestQueryIPStats_Basic(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", ClientIP: "10.0.0.1", RequestedAt: now, TotalTokens: 500, InputTokens: 300, OutputTokens: 200},
		{Provider: "openai", Model: "gpt-4o", ClientIP: "10.0.0.1", RequestedAt: now, TotalTokens: 300, Failed: true},
		{Provider: "claude", Model: "opus-4", ClientIP: "10.0.0.2", RequestedAt: now, TotalTokens: 1000},
	})

	stats, err := b.QueryIPStats(context.Background(), now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryIPStats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 IP groups, got %d", len(stats))
	}

	// Ordered by total_tokens DESC: 10.0.0.2 (1000) > 10.0.0.1 (800)
	if stats[0].ClientIP != "10.0.0.2" {
		t.Errorf("expected first IP 10.0.0.2, got %s", stats[0].ClientIP)
	}
	if stats[0].TotalTokens != 1000 {
		t.Errorf("10.0.0.2 total_tokens: got %d, want 1000", stats[0].TotalTokens)
	}

	if stats[1].ClientIP != "10.0.0.1" {
		t.Errorf("expected second IP 10.0.0.1, got %s", stats[1].ClientIP)
	}
	if stats[1].Requests != 2 || stats[1].SuccessCount != 1 || stats[1].FailureCount != 1 {
		t.Errorf("10.0.0.1: requests=%d success=%d failure=%d", stats[1].Requests, stats[1].SuccessCount, stats[1].FailureCount)
	}
}

func TestResetAll(t *testing.T) {
	b := newTestSQLiteBackend(t)
	now := time.Now()

	seedRecords(t, b, []UsageRecord{
		{Provider: "claude", Model: "opus-4", APIKey: "sk-1", RequestedAt: now, TotalTokens: 100},
		{Provider: "claude", Model: "opus-4", APIKey: "sk-2", RequestedAt: now, TotalTokens: 200},
	})

	ctx := context.Background()
	if err := b.ResetAll(ctx); err != nil {
		t.Fatalf("ResetAll: %v", err)
	}

	stats, err := b.QueryAPIKeyStats(ctx, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("QueryAPIKeyStats after reset: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 results after reset, got %d", len(stats))
	}
}

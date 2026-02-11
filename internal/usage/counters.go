package usage

import "sync/atomic"

// Counters provides lock-free atomic counters for real-time usage metrics.
// These are updated on every request for instant dashboard access.
// Historical/detailed data is queried from the database backend.
type Counters struct {
	totalRequests        atomic.Int64
	successCount         atomic.Int64
	failureCount         atomic.Int64
	totalTokens          atomic.Int64
	cacheCreationTokens  atomic.Int64
	cacheReadTokens      atomic.Int64
}

// NewCounters creates a new counter set initialized to zero.
func NewCounters() *Counters {
	return &Counters{}
}

// Record increments counters based on request outcome.
// This method is lock-free and safe for high-concurrency use.
func (c *Counters) Record(failed bool, tokens, cacheCreation, cacheRead int64) {
	if c == nil {
		return
	}
	c.totalRequests.Add(1)
	if failed {
		c.failureCount.Add(1)
	} else {
		c.successCount.Add(1)
	}
	c.totalTokens.Add(tokens)
	c.cacheCreationTokens.Add(cacheCreation)
	c.cacheReadTokens.Add(cacheRead)
}

// Snapshot returns current counter values as an immutable snapshot.
func (c *Counters) Snapshot() CounterSnapshot {
	if c == nil {
		return CounterSnapshot{}
	}
	return CounterSnapshot{
		TotalRequests:            c.totalRequests.Load(),
		SuccessCount:             c.successCount.Load(),
		FailureCount:             c.failureCount.Load(),
		TotalTokens:              c.totalTokens.Load(),
		CacheCreationInputTokens: c.cacheCreationTokens.Load(),
		CacheReadInputTokens:     c.cacheReadTokens.Load(),
	}
}

// Reset zeroes all counters. Use with caution.
func (c *Counters) Reset() {
	if c == nil {
		return
	}
	c.totalRequests.Store(0)
	c.successCount.Store(0)
	c.failureCount.Store(0)
	c.totalTokens.Store(0)
	c.cacheCreationTokens.Store(0)
	c.cacheReadTokens.Store(0)
}

// Bootstrap sets initial counter values from historical data.
// This should be called once at startup to seed counters with
// aggregated statistics from the database.
func (c *Counters) Bootstrap(total, success, failure, tokens, cacheCreation, cacheRead int64) {
	if c == nil {
		return
	}
	c.totalRequests.Store(total)
	c.successCount.Store(success)
	c.failureCount.Store(failure)
	c.totalTokens.Store(tokens)
	c.cacheCreationTokens.Store(cacheCreation)
	c.cacheReadTokens.Store(cacheRead)
}

// CounterSnapshot holds an immutable point-in-time view of counter values.
type CounterSnapshot struct {
	TotalRequests            int64 `json:"total_requests"`
	SuccessCount             int64 `json:"success_count"`
	FailureCount             int64 `json:"failure_count"`
	TotalTokens              int64 `json:"total_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

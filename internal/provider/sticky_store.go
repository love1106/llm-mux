package provider

import (
	"hash"
	"hash/fnv"
	"sync"
	"time"
)

const (
	numStickyShards       = 16
	maxEntriesPerShard    = 256
	stickyTTL             = 5 * time.Minute
	stickyCleanupInterval = 5 * time.Minute
)

type stickyEntry struct {
	authID   string
	lastUsed time.Time
}

type stickyShard struct {
	mu      sync.RWMutex
	entries map[string]*stickyEntry
}

// StickyStore provides a sharded, TTL-based cache for sticky session affinity.
// It uses background cleanup to avoid blocking the hot path.
type StickyStore struct {
	shards   [numStickyShards]*stickyShard
	stopChan chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

var hasherPool = sync.Pool{
	New: func() any { return fnv.New64a() },
}

func hashKey(key string) uint64 {
	h := hasherPool.Get().(hash.Hash64)
	h.Reset()
	h.Write([]byte(key))
	sum := h.Sum64()
	hasherPool.Put(h)
	return sum
}

// NewStickyStore creates a new sharded sticky session store.
func NewStickyStore() *StickyStore {
	s := &StickyStore{
		stopChan: make(chan struct{}),
	}
	for i := range s.shards {
		s.shards[i] = &stickyShard{
			entries: make(map[string]*stickyEntry),
		}
	}
	return s
}

func (s *StickyStore) getShard(key string) *stickyShard {
	return s.shards[hashKey(key)%numStickyShards]
}

// Get retrieves a sticky entry if it exists and is not expired.
// Returns the authID and true if found and valid, empty string and false otherwise.
func (s *StickyStore) Get(key string) (string, bool) {
	shard := s.getShard(key)
	now := time.Now()

	shard.mu.RLock()
	entry, ok := shard.entries[key]
	if !ok || now.Sub(entry.lastUsed) >= stickyTTL {
		shard.mu.RUnlock()
		return "", false
	}
	authID := entry.authID
	shard.mu.RUnlock()

	// Update lastUsed with write lock
	shard.mu.Lock()
	if entry, ok := shard.entries[key]; ok {
		entry.lastUsed = now
	}
	shard.mu.Unlock()

	return authID, true
}

// Set stores or updates a sticky entry for the given key.
func (s *StickyStore) Set(key, authID string) {
	shard := s.getShard(key)
	now := time.Now()

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if entry, ok := shard.entries[key]; ok {
		entry.authID = authID
		entry.lastUsed = now
		return
	}

	// Evict oldest entries if shard is full
	if len(shard.entries) >= maxEntriesPerShard {
		s.evictOldest(shard, now)
	}

	shard.entries[key] = &stickyEntry{
		authID:   authID,
		lastUsed: now,
	}
}

// evictOldest removes expired entries first, then oldest if still over limit.
// Caller must hold shard.mu write lock.
func (s *StickyStore) evictOldest(shard *stickyShard, now time.Time) {
	// First pass: remove expired entries
	for key, entry := range shard.entries {
		if now.Sub(entry.lastUsed) >= stickyTTL {
			delete(shard.entries, key)
		}
	}

	// Second pass: if still over limit, remove oldest
	for len(shard.entries) >= maxEntriesPerShard {
		var oldestKey string
		var oldestTime time.Time
		for key, entry := range shard.entries {
			if oldestKey == "" || entry.lastUsed.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.lastUsed
			}
		}
		if oldestKey != "" {
			delete(shard.entries, oldestKey)
		} else {
			break
		}
	}
}

// Start launches the background cleanup goroutine.
func (s *StickyStore) Start() {
	s.wg.Add(1)
	go s.cleanupLoop()
}

// Stop gracefully shuts down the background cleanup goroutine.
func (s *StickyStore) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
	s.wg.Wait()
}

func (s *StickyStore) cleanupLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(stickyCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cleanupExpired()
		}
	}
}

func (s *StickyStore) cleanupExpired() {
	now := time.Now()
	for _, shard := range s.shards {
		shard.mu.Lock()
		for key, entry := range shard.entries {
			if now.Sub(entry.lastUsed) >= stickyTTL {
				delete(shard.entries, key)
			}
		}
		shard.mu.Unlock()
	}
}

// Len returns the total number of entries across all shards.
func (s *StickyStore) Len() int {
	total := 0
	for _, shard := range s.shards {
		shard.mu.RLock()
		total += len(shard.entries)
		shard.mu.RUnlock()
	}
	return total
}

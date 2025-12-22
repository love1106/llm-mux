package usage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// LoadRecordsFromDB loads usage records from the database for the last N days
// and rebuilds the in-memory statistics.
func LoadRecordsFromDB(db *sql.DB, retentionDays int, stats *RequestStatistics) error {
	if db == nil || stats == nil {
		return fmt.Errorf("database and statistics cannot be nil")
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT
			provider, model, api_key, auth_id, auth_index, source,
			requested_at, failed, input_tokens, output_tokens,
			reasoning_tokens, cached_tokens, total_tokens,
			COALESCE(audio_tokens, 0), COALESCE(cache_creation_input_tokens, 0),
			COALESCE(cache_read_input_tokens, 0), COALESCE(tool_use_prompt_tokens, 0)
		FROM usage_records
		WHERE requested_at >= ?
		ORDER BY requested_at ASC
	`, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	recordsLoaded := 0
	for rows.Next() {
		var record UsageRecord
		err := rows.Scan(
			&record.Provider,
			&record.Model,
			&record.APIKey,
			&record.AuthID,
			&record.AuthIndex,
			&record.Source,
			&record.RequestedAt,
			&record.Failed,
			&record.InputTokens,
			&record.OutputTokens,
			&record.ReasoningTokens,
			&record.CachedTokens,
			&record.TotalTokens,
			&record.AudioTokens,
			&record.CacheCreationInputTokens,
			&record.CacheReadInputTokens,
			&record.ToolUsePromptTokens,
		)
		if err != nil {
			log.Printf("[WARN] Failed to scan usage record: %v", err)
			continue
		}

		// Rebuild in-memory stats from loaded record
		rebuildStatsFromRecord(stats, record)
		recordsLoaded++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating records: %w", err)
	}

	log.Printf("[INFO] Loaded %d usage records from database (last %d days)", recordsLoaded, retentionDays)
	return nil
}

// rebuildStatsFromRecord updates the in-memory statistics with a loaded record.
func rebuildStatsFromRecord(stats *RequestStatistics, record UsageRecord) {
	if stats == nil {
		return
	}

	timestamp := record.RequestedAt
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	detail := TokenStats{
		PromptTokens:             record.InputTokens,
		CompletionTokens:         record.OutputTokens,
		ReasoningTokens:          record.ReasoningTokens,
		CachedTokens:             record.CachedTokens,
		TotalTokens:              record.TotalTokens,
		AudioTokens:              record.AudioTokens,
		CacheCreationInputTokens: record.CacheCreationInputTokens,
		CacheReadInputTokens:     record.CacheReadInputTokens,
		ToolUsePromptTokens:      record.ToolUsePromptTokens,
	}

	// Recalculate total if needed
	if detail.TotalTokens == 0 {
		detail.TotalTokens = detail.PromptTokens + detail.CompletionTokens
	}

	statsKey := record.APIKey
	if statsKey == "" {
		statsKey = record.Provider
	}
	if statsKey == "" {
		statsKey = "unknown"
	}

	modelName := record.Model
	if modelName == "" {
		modelName = "unknown"
	}

	dayKey := timestamp.Format("2006-01-02")
	hourKey := timestamp.Hour()

	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.totalRequests++
	if record.Failed {
		stats.failureCount++
	} else {
		stats.successCount++
	}
	stats.totalTokens += detail.TotalTokens

	// Get or create API stats
	apiStatsEntry, ok := stats.apis[statsKey]
	if !ok {
		if len(stats.apis) >= maxTrackedAPIs {
			// Skip if over capacity
			return
		}
		apiStatsEntry = &apiStats{Models: make(map[string]*modelStats)}
		stats.apis[statsKey] = apiStatsEntry
	}

	// Update API-level stats
	stats.updateAPIStats(apiStatsEntry, modelName, RequestDetail{
		Timestamp: timestamp,
		Source:    record.Source,
		AuthIndex: record.AuthIndex,
		Tokens:    detail,
		Failed:    record.Failed,
	})

	// Update time-based aggregates
	stats.requestsByDay[dayKey]++
	stats.requestsByHour[hourKey]++
	stats.tokensByDay[dayKey] += detail.TotalTokens
	stats.tokensByHour[hourKey] += detail.TotalTokens
}

package management

import "time"

// UsageStatsResponse represents the structured usage statistics response.
type UsageStatsResponse struct {
	Summary    UsageSummary                  `json:"summary"`
	ByProvider map[string]UsageProviderStats `json:"by_provider,omitempty"`
	ByAccount  map[string]UsageAccountStats  `json:"by_account,omitempty"`
	ByModel    map[string]UsageModelStats    `json:"by_model,omitempty"`
	Timeline   *UsageTimeline                `json:"timeline,omitempty"`
	Period     UsagePeriod                   `json:"period"`
}

// UsageSummary holds the aggregate usage summary.
type UsageSummary struct {
	TotalRequests int64        `json:"total_requests"`
	SuccessCount  int64        `json:"success_count"`
	FailureCount  int64        `json:"failure_count"`
	Tokens        TokenSummary `json:"tokens"`
}

// TokenSummary holds token breakdown.
type TokenSummary struct {
	Total     int64 `json:"total"`
	Input     int64 `json:"input"`
	Output    int64 `json:"output"`
	Reasoning int64 `json:"reasoning,omitempty"`
}

// UsageProviderStats represents per-provider statistics.
type UsageProviderStats struct {
	Requests     int64        `json:"requests"`
	Success      int64        `json:"success"`
	Failure      int64        `json:"failure"`
	Tokens       TokenSummary `json:"tokens"`
	AccountCount int64        `json:"accounts"`
	Models       []string     `json:"models,omitempty"`
}

// UsageAccountStats represents per-auth-account statistics.
type UsageAccountStats struct {
	Provider string       `json:"provider"`
	AuthID   string       `json:"auth_id"`
	Requests int64        `json:"requests"`
	Success  int64        `json:"success"`
	Failure  int64        `json:"failure"`
	Tokens   TokenSummary `json:"tokens"`
}

// UsageModelStats represents per-model statistics.
type UsageModelStats struct {
	Provider string       `json:"provider"`
	Requests int64        `json:"requests"`
	Success  int64        `json:"success"`
	Failure  int64        `json:"failure"`
	Tokens   TokenSummary `json:"tokens"`
}

// UsageTimeline holds time-series usage data.
type UsageTimeline struct {
	ByDay  []UsageDayStats  `json:"by_day,omitempty"`
	ByHour []UsageHourStats `json:"by_hour,omitempty"`
}

// UsageDayStats represents aggregated daily stats.
type UsageDayStats struct {
	Day      string `json:"day"`
	Requests int64  `json:"requests"`
	Tokens   int64  `json:"tokens"`
}

// UsageHourStats represents aggregated hourly stats.
type UsageHourStats struct {
	Hour     int   `json:"hour"`
	Requests int64 `json:"requests"`
	Tokens   int64 `json:"tokens"`
}

// UsagePeriod describes the time range of the statistics.
type UsagePeriod struct {
	From          time.Time `json:"from"`
	To            time.Time `json:"to"`
	RetentionDays int       `json:"retention_days"`
}

// UsageCounters holds the atomic counters (deprecated, use UsageSummary).
type UsageCounters struct {
	TotalRequests int64 `json:"total_requests"`
	SuccessCount  int64 `json:"success_count"`
	FailureCount  int64 `json:"failure_count"`
	TotalTokens   int64 `json:"total_tokens"`
}

// UsageAPIStats represents aggregated per-API stats (deprecated).
type UsageAPIStats struct {
	TotalRequests int64                         `json:"total_requests"`
	TotalTokens   int64                         `json:"total_tokens"`
	Models        map[string]UsageAPIModelStats `json:"models,omitempty"`
}

// UsageAPIModelStats represents per-model stats within an API (deprecated).
type UsageAPIModelStats struct {
	TotalRequests int64 `json:"total_requests"`
	TotalTokens   int64 `json:"total_tokens"`
}

// ConfigUpdateResponse represents the response after updating config.
type ConfigUpdateResponse struct {
	Status  string   `json:"status"`
	Changed []string `json:"changed,omitempty"`
	Value   any      `json:"value,omitempty"`
}

// LogEntryResponse represents a single log line in API response.
type LogEntryResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

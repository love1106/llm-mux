package executor

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nghyane/llm-mux/internal/translator/ir"
	"github.com/nghyane/llm-mux/internal/translator/to_ir"
	cliproxyauth "github.com/nghyane/llm-mux/sdk/cliproxy/auth"
	"github.com/nghyane/llm-mux/sdk/cliproxy/usage"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type usageReporter struct {
	provider    string
	model       string
	authID      string
	authIndex   uint64
	apiKey      string
	source      string
	requestedAt time.Time
	once        sync.Once
}

func newUsageReporter(ctx context.Context, provider, model string, auth *cliproxyauth.Auth) *usageReporter {
	apiKey := apiKeyFromContext(ctx)
	reporter := &usageReporter{
		provider:    provider,
		model:       model,
		requestedAt: time.Now(),
		apiKey:      apiKey,
		source:      resolveUsageSource(auth, apiKey),
	}
	if auth != nil {
		reporter.authID = auth.ID
		reporter.authIndex = auth.EnsureIndex()
	}
	return reporter
}

// publish publishes usage from IR Usage struct (optimized path - no duplicate parsing).
func (r *usageReporter) publish(ctx context.Context, u *ir.Usage) {
	r.publishWithOutcome(ctx, u, false)
}

func (r *usageReporter) publishFailure(ctx context.Context) {
	r.publishWithOutcome(ctx, nil, true)
}

func (r *usageReporter) trackFailure(ctx context.Context, errPtr *error) {
	if r == nil || errPtr == nil {
		return
	}
	if *errPtr != nil {
		// Don't count user errors (400 Bad Request) as failures
		if !isUserError(*errPtr) {
			r.publishFailure(ctx)
		}
	}
}

// isUserError checks if the error is a user error (400 Bad Request) that should not be counted as failure.
func isUserError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error implements StatusCode() method
	type statusCoder interface {
		StatusCode() int
	}
	if sc, ok := err.(statusCoder); ok {
		return sc.StatusCode() == 400
	}
	// Check if error implements Category() method
	type categorizer interface {
		Category() cliproxyauth.ErrorCategory
	}
	if cat, ok := err.(categorizer); ok {
		return cat.Category() == cliproxyauth.CategoryUserError
	}
	return false
}

func (r *usageReporter) publishWithOutcome(ctx context.Context, u *ir.Usage, failed bool) {
	if r == nil {
		return
	}
	// Skip empty usage records unless it's a failure
	if u == nil && !failed {
		return
	}
	if u != nil && u.TotalTokens == 0 && u.PromptTokens == 0 && u.CompletionTokens == 0 && !failed {
		return
	}
	r.once.Do(func() {
		usage.PublishRecord(ctx, usage.Record{
			Provider:    r.provider,
			Model:       r.model,
			Source:      r.source,
			APIKey:      r.apiKey,
			AuthID:      r.authID,
			AuthIndex:   r.authIndex,
			RequestedAt: r.requestedAt,
			Failed:      failed,
			Usage:       u,
		})
	})
}

// ensurePublished guarantees that a usage record is emitted exactly once.
// It is safe to call multiple times; only the first call wins due to once.Do.
// This is used to ensure request counting even when upstream responses do not
// include any usage fields (tokens), especially for streaming paths.
func (r *usageReporter) ensurePublished(ctx context.Context) {
	if r == nil {
		return
	}
	r.once.Do(func() {
		usage.PublishRecord(ctx, usage.Record{
			Provider:    r.provider,
			Model:       r.model,
			Source:      r.source,
			APIKey:      r.apiKey,
			AuthID:      r.authID,
			AuthIndex:   r.authIndex,
			RequestedAt: r.requestedAt,
			Failed:      false,
			Usage:       nil,
		})
	})
}

func apiKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	ginCtx, ok := ctx.Value("gin").(*gin.Context)
	if !ok || ginCtx == nil {
		return ""
	}
	if v, exists := ginCtx.Get("apiKey"); exists {
		switch value := v.(type) {
		case string:
			return value
		case fmt.Stringer:
			return value.String()
		default:
			return fmt.Sprintf("%v", value)
		}
	}
	return ""
}

func resolveUsageSource(auth *cliproxyauth.Auth, ctxAPIKey string) string {
	if auth != nil {
		provider := strings.TrimSpace(auth.Provider)
		if strings.EqualFold(provider, "gemini-cli") {
			if id := strings.TrimSpace(auth.ID); id != "" {
				return id
			}
		}
		if strings.EqualFold(provider, "vertex") {
			if auth.Metadata != nil {
				if projectID, ok := auth.Metadata["project_id"].(string); ok {
					if trimmed := strings.TrimSpace(projectID); trimmed != "" {
						return trimmed
					}
				}
				if project, ok := auth.Metadata["project"].(string); ok {
					if trimmed := strings.TrimSpace(project); trimmed != "" {
						return trimmed
					}
				}
			}
		}
		if _, value := auth.AccountInfo(); value != "" {
			return strings.TrimSpace(value)
		}
		if auth.Metadata != nil {
			if email, ok := auth.Metadata["email"].(string); ok {
				if trimmed := strings.TrimSpace(email); trimmed != "" {
					return trimmed
				}
			}
		}
		if auth.Attributes != nil {
			if key := strings.TrimSpace(auth.Attributes["api_key"]); key != "" {
				return key
			}
		}
	}
	if trimmed := strings.TrimSpace(ctxAPIKey); trimmed != "" {
		return trimmed
	}
	return ""
}

// =============================================================================
// Usage Parsing Functions - Return *ir.Usage for unified tracking
// =============================================================================

// parseGeminiUsageMetadata is a shared helper for Gemini/Antigravity usage parsing.
// =============================================================================
// Optimized Usage Extraction from Translated Responses
// These functions extract usage directly from translator IR parsing
// instead of duplicating JSON parsing from raw responses.
// =============================================================================

// extractUsageFromClaudeResponse extracts usage from Claude response using IR parser.
func extractUsageFromClaudeResponse(data []byte) *ir.Usage {
	_, usage, err := to_ir.ParseClaudeResponse(data)
	if err != nil {
		return nil
	}
	return usage
}

// extractUsageFromOpenAIResponse extracts usage from OpenAI response using IR parser.
func extractUsageFromOpenAIResponse(data []byte) *ir.Usage {
	_, usage, err := to_ir.ParseOpenAIResponse(data)
	if err != nil {
		return nil
	}
	return usage
}

// extractUsageFromGeminiResponse extracts usage from Gemini response using IR parser.
func extractUsageFromGeminiResponse(data []byte) *ir.Usage {
	_, _, usage, err := to_ir.ParseGeminiResponse(data)
	if err != nil {
		return nil
	}
	return usage
}

var stopChunkWithoutUsage sync.Map

func rememberStopWithoutUsage(traceID string) {
	stopChunkWithoutUsage.Store(traceID, struct{}{})
	time.AfterFunc(10*time.Minute, func() { stopChunkWithoutUsage.Delete(traceID) })
}

// FilterSSEUsageMetadata removes usageMetadata from SSE events that are not
// terminal (finishReason != "stop"). Stop chunks are left untouched. This
// function is shared between aistudio and antigravity executors.
func FilterSSEUsageMetadata(payload []byte) []byte {
	if len(payload) == 0 {
		return payload
	}

	lines := bytes.Split(payload, []byte("\n"))
	var outputLines [][]byte
	modified := false
	foundData := false

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 || !bytes.HasPrefix(trimmed, []byte("data:")) {
			outputLines = append(outputLines, line)
			continue
		}
		foundData = true
		dataIdx := bytes.Index(line, []byte("data:"))
		if dataIdx < 0 {
			outputLines = append(outputLines, line)
			continue
		}
		rawJSON := bytes.TrimSpace(line[dataIdx+5:])
		// Skip data payloads that are empty or whitespace-only
		if len(rawJSON) == 0 {
			outputLines = append(outputLines, line)
			continue
		}
		traceID := gjson.GetBytes(rawJSON, "traceId").String()
		if isStopChunkWithoutUsage(rawJSON) && traceID != "" {
			rememberStopWithoutUsage(traceID)
			modified = true
			continue
		}
		if traceID != "" {
			if _, ok := stopChunkWithoutUsage.Load(traceID); ok && hasUsageMetadata(rawJSON) {
				stopChunkWithoutUsage.Delete(traceID)
				modified = true
				continue
			}
		}

		cleaned, changed := StripUsageMetadataFromJSON(rawJSON)
		if !changed {
			outputLines = append(outputLines, line)
			continue
		}
		var rebuilt []byte
		rebuilt = append(rebuilt, line[:dataIdx]...)
		rebuilt = append(rebuilt, []byte("data:")...)
		if len(cleaned) > 0 {
			rebuilt = append(rebuilt, ' ')
			rebuilt = append(rebuilt, cleaned...)
		}
		outputLines = append(outputLines, rebuilt)
		modified = true
	}
	if !modified {
		if !foundData {
			// Handle payloads that are raw JSON without SSE data: prefix.
			trimmed := bytes.TrimSpace(payload)
			cleaned, changed := StripUsageMetadataFromJSON(trimmed)
			if !changed {
				return payload
			}
			return cleaned
		}
		return payload
	}
	return bytes.Join(outputLines, []byte("\n"))
}

// StripUsageMetadataFromJSON drops usageMetadata unless finishReason is present (terminal).
// It handles both formats:
// - Aistudio: candidates.0.finishReason
// - Antigravity: response.candidates.0.finishReason
func StripUsageMetadataFromJSON(rawJSON []byte) ([]byte, bool) {
	jsonBytes := bytes.TrimSpace(rawJSON)
	if len(jsonBytes) == 0 || !gjson.ValidBytes(jsonBytes) {
		return rawJSON, false
	}

	// Check for finishReason in both aistudio and antigravity formats
	finishReason := gjson.GetBytes(jsonBytes, "candidates.0.finishReason")
	if !finishReason.Exists() {
		finishReason = gjson.GetBytes(jsonBytes, "response.candidates.0.finishReason")
	}
	terminalReason := finishReason.Exists() && strings.TrimSpace(finishReason.String()) != ""

	usageMetadata := gjson.GetBytes(jsonBytes, "usageMetadata")
	if !usageMetadata.Exists() {
		usageMetadata = gjson.GetBytes(jsonBytes, "response.usageMetadata")
	}

	// Terminal chunk: keep as-is.
	if terminalReason {
		return rawJSON, false
	}

	// Nothing to strip
	if !usageMetadata.Exists() {
		return rawJSON, false
	}

	// Remove usageMetadata from both possible locations
	cleaned := jsonBytes
	var changed bool

	if gjson.GetBytes(cleaned, "usageMetadata").Exists() {
		cleaned, _ = sjson.DeleteBytes(cleaned, "usageMetadata")
		changed = true
	}

	if gjson.GetBytes(cleaned, "response.usageMetadata").Exists() {
		cleaned, _ = sjson.DeleteBytes(cleaned, "response.usageMetadata")
		changed = true
	}

	return cleaned, changed
}

func hasUsageMetadata(jsonBytes []byte) bool {
	if len(jsonBytes) == 0 || !gjson.ValidBytes(jsonBytes) {
		return false
	}
	if gjson.GetBytes(jsonBytes, "usageMetadata").Exists() {
		return true
	}
	if gjson.GetBytes(jsonBytes, "response.usageMetadata").Exists() {
		return true
	}
	return false
}

func isStopChunkWithoutUsage(jsonBytes []byte) bool {
	if len(jsonBytes) == 0 || !gjson.ValidBytes(jsonBytes) {
		return false
	}
	finishReason := gjson.GetBytes(jsonBytes, "candidates.0.finishReason")
	if !finishReason.Exists() {
		finishReason = gjson.GetBytes(jsonBytes, "response.candidates.0.finishReason")
	}
	trimmed := strings.TrimSpace(finishReason.String())
	if !finishReason.Exists() || trimmed == "" {
		return false
	}
	return !hasUsageMetadata(jsonBytes)
}

func jsonPayload(line []byte) []byte {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return nil
	}
	if bytes.Equal(trimmed, []byte("[DONE]")) {
		return nil
	}
	if bytes.HasPrefix(trimmed, []byte("event:")) {
		return nil
	}
	if bytes.HasPrefix(trimmed, []byte("data:")) {
		trimmed = bytes.TrimSpace(trimmed[len("data:"):])
	}
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil
	}
	return trimmed
}

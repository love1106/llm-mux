package util

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	GeminiThinkingBudgetMetadataKey  = "gemini_thinking_budget"
	GeminiIncludeThoughtsMetadataKey = "gemini_include_thoughts"
	GeminiOriginalModelMetadataKey   = "gemini_original_model"
)

func ParseGeminiThinkingSuffix(model string) (string, *int, *bool, bool) {
	if model == "" {
		return model, nil, nil, false
	}
	lower := strings.ToLower(model)
	if !strings.HasPrefix(lower, "gemini-") {
		return model, nil, nil, false
	}

	if strings.HasSuffix(lower, "-nothinking") {
		base := model[:len(model)-len("-nothinking")]
		budgetValue := 0
		if strings.HasPrefix(lower, "gemini-2.5-pro") {
			budgetValue = 128
		}
		include := false
		return base, &budgetValue, &include, true
	}

	// Handle "-reasoning" suffix: enables thinking with dynamic budget (-1)
	// Maps: gemini-2.5-flash-reasoning -> gemini-2.5-flash with thinkingBudget=-1
	if strings.HasSuffix(lower, "-reasoning") {
		base := model[:len(model)-len("-reasoning")]
		budgetValue := -1 // Dynamic budget
		include := true
		return base, &budgetValue, &include, true
	}

	idx := strings.LastIndex(lower, "-thinking-")
	if idx == -1 {
		return model, nil, nil, false
	}

	digits := model[idx+len("-thinking-"):]
	if digits == "" {
		return model, nil, nil, false
	}
	end := len(digits)
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			end = i
			break
		}
	}
	if end == 0 {
		return model, nil, nil, false
	}
	valueStr := digits[:end]
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return model, nil, nil, false
	}
	base := model[:idx]
	budgetValue := value
	return base, &budgetValue, nil, true
}

func NormalizeGeminiThinkingModel(modelName string) (string, map[string]any) {
	baseModel, budget, include, matched := ParseGeminiThinkingSuffix(modelName)
	if !matched {
		return baseModel, nil
	}
	metadata := map[string]any{
		GeminiOriginalModelMetadataKey: modelName,
	}
	if budget != nil {
		metadata[GeminiThinkingBudgetMetadataKey] = *budget
	}
	if include != nil {
		metadata[GeminiIncludeThoughtsMetadataKey] = *include
	}
	return baseModel, metadata
}

func ApplyGeminiThinkingConfig(body []byte, budget *int, includeThoughts *bool) []byte {
	if budget == nil && includeThoughts == nil {
		return body
	}
	updated := body
	if budget != nil {
		valuePath := "generationConfig.thinkingConfig.thinkingBudget"
		rewritten, err := sjson.SetBytes(updated, valuePath, *budget)
		if err == nil {
			updated = rewritten
		}
	}
	if includeThoughts != nil {
		valuePath := "generationConfig.thinkingConfig.include_thoughts"
		rewritten, err := sjson.SetBytes(updated, valuePath, *includeThoughts)
		if err == nil {
			updated = rewritten
		}
	}
	return updated
}

func ApplyGeminiCLIThinkingConfig(body []byte, budget *int, includeThoughts *bool) []byte {
	if budget == nil && includeThoughts == nil {
		return body
	}
	updated := body
	if budget != nil {
		valuePath := "request.generationConfig.thinkingConfig.thinkingBudget"
		rewritten, err := sjson.SetBytes(updated, valuePath, *budget)
		if err == nil {
			updated = rewritten
		}
	}
	if includeThoughts != nil {
		valuePath := "request.generationConfig.thinkingConfig.include_thoughts"
		rewritten, err := sjson.SetBytes(updated, valuePath, *includeThoughts)
		if err == nil {
			updated = rewritten
		}
	}
	return updated
}

func GeminiThinkingFromMetadata(metadata map[string]any) (*int, *bool, bool) {
	if len(metadata) == 0 {
		return nil, nil, false
	}
	var (
		budgetPtr  *int
		includePtr *bool
		matched    bool
	)
	if rawBudget, ok := metadata[GeminiThinkingBudgetMetadataKey]; ok {
		switch v := rawBudget.(type) {
		case int:
			budget := v
			budgetPtr = &budget
			matched = true
		case int32:
			budget := int(v)
			budgetPtr = &budget
			matched = true
		case int64:
			budget := int(v)
			budgetPtr = &budget
			matched = true
		case float64:
			budget := int(v)
			budgetPtr = &budget
			matched = true
		case json.Number:
			if val, err := v.Int64(); err == nil {
				budget := int(val)
				budgetPtr = &budget
				matched = true
			}
		}
	}
	if rawInclude, ok := metadata[GeminiIncludeThoughtsMetadataKey]; ok {
		switch v := rawInclude.(type) {
		case bool:
			include := v
			includePtr = &include
			matched = true
		case string:
			if parsed, err := strconv.ParseBool(v); err == nil {
				include := parsed
				includePtr = &include
				matched = true
			}
		case json.Number:
			if val, err := v.Int64(); err == nil {
				include := val != 0
				includePtr = &include
				matched = true
			}
		case int:
			include := v != 0
			includePtr = &include
			matched = true
		case int32:
			include := v != 0
			includePtr = &include
			matched = true
		case int64:
			include := v != 0
			includePtr = &include
			matched = true
		case float64:
			include := v != 0
			includePtr = &include
			matched = true
		}
	}
	return budgetPtr, includePtr, matched
}

// StripThinkingConfigIfUnsupported removes thinkingConfig from the request body
// when the target model does not advertise Thinking capability. It cleans both
// standard Gemini and Gemini CLI JSON envelopes. This acts as a final safety net
// in case upstream injected thinking for an unsupported model.
func StripThinkingConfigIfUnsupported(model string, body []byte) []byte {
	if ModelSupportsThinking(model) || len(body) == 0 {
		return body
	}
	updated := body
	// Gemini CLI path
	updated, _ = sjson.DeleteBytes(updated, "request.generationConfig.thinkingConfig")
	// Standard Gemini path
	updated, _ = sjson.DeleteBytes(updated, "generationConfig.thinkingConfig")
	return updated
}

// ConvertThinkingLevelToBudget checks for "generationConfig.thinkingConfig.thinkingLevel"
// and converts it to "thinkingBudget".
// "high" -> 32768
// "low" -> 128
// It removes "thinkingLevel" after conversion.
func ConvertThinkingLevelToBudget(body []byte) []byte {
	levelPath := "generationConfig.thinkingConfig.thinkingLevel"
	res := gjson.GetBytes(body, levelPath)
	if !res.Exists() {
		return body
	}

	level := strings.ToLower(res.String())
	var budget int
	switch level {
	case "high":
		budget = 32768
	case "low":
		budget = 128
	default:
		// If unknown level, we might just leave it or default.
		// User only specified high and low. We'll assume we shouldn't touch it if it's something else,
		// or maybe we should just remove the invalid level?
		// For safety adhering to strict instructions: "If high... if low...".
		// If it's something else, the upstream might fail anyway if we leave it,
		// but let's just delete the level if we processed it.
		// Actually, let's check if we need to do anything for other values.
		// For now, only handle high/low.
		return body
	}

	// Set budget
	budgetPath := "generationConfig.thinkingConfig.thinkingBudget"
	updated, err := sjson.SetBytes(body, budgetPath, budget)
	if err != nil {
		return body
	}

	// Remove level
	updated, err = sjson.DeleteBytes(updated, levelPath)
	if err != nil {
		return body
	}
	return updated
}

// ConvertBudgetToThinkingLevel converts thinkingBudget to thinking_level for Gemini 3 models.
// Gemini 3 uses thinking_level (MINIMAL/LOW/MEDIUM/HIGH) instead of thinkingBudget.
//
// Mapping for Gemini 3 Flash (supports all levels):
//   - budget ≤ 128   → MINIMAL
//   - budget ≤ 1024  → LOW
//   - budget ≤ 8192  → MEDIUM
//   - budget > 8192  → HIGH
//
// Mapping for Gemini 3 Pro (no MINIMAL/MEDIUM):
//   - budget ≤ 1024  → LOW
//   - budget > 1024  → HIGH
func ConvertBudgetToThinkingLevel(model string, budget int) string {
	isFlash := strings.Contains(strings.ToLower(model), "flash")

	switch {
	case budget <= 128:
		if isFlash {
			return "MINIMAL"
		}
		return "LOW" // Pro doesn't have MINIMAL
	case budget <= 1024:
		return "LOW"
	case budget <= 8192:
		if isFlash {
			return "MEDIUM"
		}
		return "HIGH" // Pro doesn't have MEDIUM
	default:
		return "HIGH"
	}
}

// GetDefaultThinkingLevel returns the default thinking level for a Gemini 3 model.
// Flash defaults to MEDIUM, Pro defaults to HIGH.
func GetDefaultThinkingLevel(model string) string {
	if strings.Contains(strings.ToLower(model), "flash") {
		return "MEDIUM"
	}
	return "HIGH"
}

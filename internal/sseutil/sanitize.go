package sseutil

import (
	"strings"

	"github.com/nghyane/llm-mux/internal/json"
	"github.com/tidwall/gjson"
)

// SanitizeUndefinedValues removes "[undefined]" placeholder values from JSON payloads.
// Some clients send "[undefined]" as a placeholder for missing values which breaks
// provider APIs. This function recursively removes such entries.
func SanitizeUndefinedValues(payload []byte) []byte {
	if !strings.Contains(string(payload), "[undefined]") {
		return payload
	}
	result := gjson.ParseBytes(payload)
	if !result.IsObject() && !result.IsArray() {
		return payload
	}
	cleaned := cleanUndefinedRecursive(result.Value())
	if cleaned == nil {
		return payload
	}
	out, err := json.Marshal(cleaned)
	if err != nil {
		return payload
	}
	return out
}

// cleanUndefinedRecursive recursively removes "[undefined]" values from nested structures.
func cleanUndefinedRecursive(v any) any {
	switch val := v.(type) {
	case map[string]any:
		cleaned := make(map[string]any)
		for k, child := range val {
			if str, ok := child.(string); ok && str == "[undefined]" {
				continue
			}
			if cleanedChild := cleanUndefinedRecursive(child); cleanedChild != nil {
				cleaned[k] = cleanedChild
			}
		}
		if len(cleaned) == 0 {
			return nil
		}
		return cleaned
	case []any:
		var cleaned []any
		for _, item := range val {
			if str, ok := item.(string); ok && str == "[undefined]" {
				continue
			}
			if cleanedItem := cleanUndefinedRecursive(item); cleanedItem != nil {
				cleaned = append(cleaned, cleanedItem)
			}
		}
		return cleaned
	default:
		return v
	}
}

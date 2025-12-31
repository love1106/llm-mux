package ir

import (
	"testing"
)

func TestValidateAntigravityPayload_RemovesUnsupportedFieldsForClaude(t *testing.T) {
	payload := []byte(`{
		"model": "claude-3-5-sonnet",
		"userAgent": "antigravity",
		"project": "test-project",
		"requestId": "test-request",
		"request": {
			"contents": [{"role": "user", "parts": [{"text": "hello"}]}],
			"safetySettings": [{"category": "HARM_CATEGORY_HATE_SPEECH", "threshold": "OFF"}],
			"cachedContent": "projects/test/cachedContents/123",
			"labels": {"env": "test"},
			"generationConfig": {
				"temperature": 0.7,
				"maxOutputTokens": 1000
			}
		}
	}`)

	result, report := ValidateAntigravityPayload(payload, "claude-3-5-sonnet")

	if !report.Changed {
		t.Error("expected changes for Claude model with Gemini-specific fields")
	}

	resultStr := string(result)

	if contains(resultStr, "safetySettings") {
		t.Error("safetySettings should be removed for Claude")
	}
	if contains(resultStr, "cachedContent") {
		t.Error("cachedContent should be removed for Claude")
	}
	if contains(resultStr, "labels") {
		t.Error("labels should be removed for Claude")
	}
	if !contains(resultStr, "temperature") {
		t.Error("temperature should be preserved")
	}
	if !contains(resultStr, "maxOutputTokens") {
		t.Error("maxOutputTokens should be preserved")
	}
}

func TestValidateAntigravityPayload_PreservesFieldsForGemini(t *testing.T) {
	payload := []byte(`{
		"model": "gemini-2.5-pro",
		"userAgent": "antigravity",
		"project": "test-project",
		"requestId": "test-request",
		"request": {
			"contents": [{"role": "user", "parts": [{"text": "hello"}]}],
			"safetySettings": [{"category": "HARM_CATEGORY_HATE_SPEECH", "threshold": "OFF"}],
			"cachedContent": "projects/test/cachedContents/123",
			"generationConfig": {
				"temperature": 0.7,
				"candidateCount": 2,
				"responseMimeType": "application/json"
			}
		}
	}`)

	result, _ := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	resultStr := string(result)

	if !contains(resultStr, "safetySettings") {
		t.Error("safetySettings should be preserved for Gemini")
	}
	if !contains(resultStr, "cachedContent") {
		t.Error("cachedContent should be preserved for Gemini")
	}
	if !contains(resultStr, "candidateCount") {
		t.Error("candidateCount should be preserved for Gemini")
	}
}

func TestValidateAntigravityPayload_RemovesNullValues(t *testing.T) {
	payload := []byte(`{
		"model": "gemini-2.5-pro",
		"project": "test",
		"request": {
			"contents": [{"role": "user", "parts": [{"text": "hello"}]}],
			"generationConfig": {
				"temperature": null,
				"maxOutputTokens": 1000
			}
		}
	}`)

	result, report := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	t.Logf("Result: %s", string(result))
	t.Logf("Report.Changed: %v, Entries: %+v", report.Changed, report.Entries)

	if !report.Changed {
		t.Error("expected changes when removing null values")
	}

	resultStr := string(result)
	if contains(resultStr, "null") {
		t.Error("null values should be removed")
	}
	if !contains(resultStr, "maxOutputTokens") {
		t.Error("valid fields should be preserved")
	}
}

func TestValidateAntigravityPayload_RemovesUnknownFields(t *testing.T) {
	payload := []byte(`{
		"model": "gemini-2.5-pro",
		"project": "test",
		"unknownField": "should be removed",
		"request": {
			"contents": [{"role": "user", "parts": [{"text": "hello"}]}],
			"invalidField": "also removed",
			"generationConfig": {
				"temperature": 0.5,
				"notAValidField": 123
			}
		}
	}`)

	result, report := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	if !report.Changed {
		t.Error("expected changes for unknown fields")
	}

	resultStr := string(result)
	if contains(resultStr, "unknownField") {
		t.Error("unknownField should be removed")
	}
	if contains(resultStr, "invalidField") {
		t.Error("invalidField should be removed")
	}
	if contains(resultStr, "notAValidField") {
		t.Error("notAValidField should be removed from generationConfig")
	}
}

func TestValidateAntigravityPayload_RemovesUndefinedStrings(t *testing.T) {
	payload := []byte(`{
		"model": "gemini-2.5-pro",
		"project": "test",
		"request": {
			"contents": [{"role": "user", "parts": [{"text": "[undefined]"}]}],
			"generationConfig": {
				"temperature": 0.5
			}
		}
	}`)

	result, report := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	if !report.Changed {
		t.Error("expected changes for undefined values")
	}

	resultStr := string(result)
	if contains(resultStr, "[undefined]") {
		t.Error("[undefined] strings should be removed")
	}
}

func TestValidateAntigravityPayload_HandlesEmptyPayload(t *testing.T) {
	payload := []byte(`{}`)

	result, report := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	if report.Changed {
		t.Error("empty payload should not change")
	}
	if string(result) != "{}" {
		t.Errorf("expected {}, got %s", string(result))
	}
}

func TestValidateAntigravityPayload_HandlesInvalidJSON(t *testing.T) {
	payload := []byte(`{invalid json}`)

	result, report := ValidateAntigravityPayload(payload, "gemini-2.5-pro")

	if report.Changed {
		t.Error("invalid JSON should not be changed")
	}
	if string(result) != string(payload) {
		t.Error("invalid JSON should be returned as-is")
	}
}

func TestGetRequestSpecForModel_ReturnsCorrectSpec(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		{"claude-3-5-sonnet", "VertexClaudeRequest"},
		{"claude-3-opus", "VertexClaudeRequest"},
		{"Claude-3-Haiku", "VertexClaudeRequest"},
		{"gemini-2.5-pro", "VertexGeminiRequest"},
		{"gemini-3-flash", "VertexGeminiRequest"},
		{"random-model", "VertexGeminiRequest"},
	}

	for _, tt := range tests {
		spec := GetRequestSpecForModel(tt.model)
		if spec.Name != tt.expected {
			t.Errorf("GetRequestSpecForModel(%q) = %q, want %q", tt.model, spec.Name, tt.expected)
		}
	}
}

func TestValidationReport_LogDebug(t *testing.T) {
	report := &ValidationReport{
		SpecName: "test",
		Changed:  true,
		Entries: []ValidationEntry{
			{Path: "request.safetySettings", Action: ActionRemoved, Reason: "not in whitelist"},
		},
	}

	report.LogDebug()
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

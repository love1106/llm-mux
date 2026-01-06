// Package cloudcode provides utilities for Cloud Code API providers (Antigravity, Gemini CLI).
// These providers use a specific envelope format for request/response wrapping.
package cloudcode

import (
	"bytes"

	"github.com/nghyane/llm-mux/internal/sseutil"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// RequestEnvelope wraps a Gemini payload in the Cloud Code request format.
// Input:  {"contents": [...], "generationConfig": {...}}
// Output: {"request": {"contents": [...], "generationConfig": {...}}}
func RequestEnvelope(payload []byte) []byte {
	return sseutil.WrapEnvelope(payload)
}

// ResponseUnwrap extracts the inner response from Cloud Code envelope.
// Input:  {"response": {"candidates": [...]}}
// Output: {"candidates": [...]}
func ResponseUnwrap(payload []byte) []byte {
	return sseutil.UnwrapEnvelope(payload)
}

// StreamPreprocessor returns a preprocessor function for Cloud Code SSE streams.
// It unwraps the envelope from each streaming chunk.
func StreamPreprocessor(basePreprocessor func([]byte) ([]byte, bool)) func([]byte) ([]byte, bool) {
	return func(line []byte) ([]byte, bool) {
		payload, skip := basePreprocessor(line)
		if skip || payload == nil {
			return nil, true
		}
		return sseutil.UnwrapEnvelope(payload), false
	}
}

// IsEnvelopeWrapped checks if the payload is wrapped in a Cloud Code envelope.
func IsEnvelopeWrapped(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	parsed := gjson.ParseBytes(trimmed)
	return parsed.Get("request").Exists() || parsed.Get("response").Exists()
}

// EnsureRequestEnvelope wraps payload only if not already wrapped.
func EnsureRequestEnvelope(payload []byte) []byte {
	if IsEnvelopeWrapped(payload) {
		return payload
	}
	return RequestEnvelope(payload)
}

// EnsureResponseUnwrap unwraps only if wrapped.
func EnsureResponseUnwrap(payload []byte) []byte {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return payload
	}
	parsed := gjson.ParseBytes(trimmed)
	if response := parsed.Get("response"); response.Exists() && response.IsObject() {
		return []byte(response.Raw)
	}
	return payload
}

// AddEnvelopeFields adds Cloud Code specific fields to a wrapped request.
// Used by Antigravity to add project, model, requestId etc.
func AddEnvelopeFields(payload []byte, fields map[string]any) []byte {
	result := payload
	for key, value := range fields {
		updated, err := sjson.SetBytes(result, key, value)
		if err == nil {
			result = updated
		}
	}
	return result
}

// DeleteEnvelopeFields removes fields from a wrapped request.
func DeleteEnvelopeFields(payload []byte, keys ...string) []byte {
	result := payload
	for _, key := range keys {
		updated, err := sjson.DeleteBytes(result, key)
		if err == nil {
			result = updated
		}
	}
	return result
}

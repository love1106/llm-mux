package ir

import (
	"github.com/nghyane/llm-mux/internal/json"
	log "github.com/nghyane/llm-mux/internal/logging"
)

type ValidationAction string

const (
	ActionRemoved   ValidationAction = "removed"
	ActionNullified ValidationAction = "nullified"
	ActionFixed     ValidationAction = "fixed"
)

type ValidationEntry struct {
	Path   string
	Action ValidationAction
	Reason string
}

type ValidationReport struct {
	SpecName string
	Entries  []ValidationEntry
	Changed  bool
}

func (r *ValidationReport) AddEntry(path string, action ValidationAction, reason string) {
	r.Entries = append(r.Entries, ValidationEntry{Path: path, Action: action, Reason: reason})
	r.Changed = true
}

func (r *ValidationReport) LogDebug() {
	if !r.Changed || len(r.Entries) == 0 {
		return
	}
	log.Debugf("payload_validator [%s]: %d changes applied", r.SpecName, len(r.Entries))
	for _, e := range r.Entries {
		log.Debugf("  %s: %s (%s)", e.Action, e.Path, e.Reason)
	}
}

func ValidateAndSanitizePayload(payload []byte, spec *PayloadSpec) ([]byte, *ValidationReport) {
	if spec == nil || len(payload) == 0 {
		return payload, &ValidationReport{SpecName: "none", Changed: false}
	}

	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return payload, &ValidationReport{SpecName: spec.Name, Changed: false}
	}

	report := &ValidationReport{SpecName: spec.Name}
	cleaned := sanitizeObject(data, spec.Fields, "", report)

	if !report.Changed {
		return payload, report
	}

	result, err := json.Marshal(cleaned)
	if err != nil {
		return payload, report
	}
	return result, report
}

func ValidateAntigravityPayload(payload []byte, modelName string) ([]byte, *ValidationReport) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return payload, &ValidationReport{SpecName: "antigravity", Changed: false}
	}

	report := &ValidationReport{SpecName: "antigravity"}

	wrapperCleaned := sanitizeObject(data, AntigravityWrapperSpec.Fields, "", report)

	if req, ok := wrapperCleaned["request"].(map[string]any); ok {
		requestSpec := GetRequestSpecForModel(modelName)
		reqCleaned := sanitizeObject(req, requestSpec.Fields, "request", report)
		wrapperCleaned["request"] = reqCleaned
	}

	if !report.Changed {
		return payload, report
	}

	result, err := json.Marshal(wrapperCleaned)
	if err != nil {
		return payload, report
	}
	return result, report
}

func sanitizeObject(data map[string]any, allowedFields map[string]*FieldSpec, path string, report *ValidationReport) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any, len(data))

	for key, value := range data {
		fieldPath := key
		if path != "" {
			fieldPath = path + "." + key
		}

		spec, allowed := allowedFields[key]
		if !allowed {
			report.AddEntry(fieldPath, ActionRemoved, "not in whitelist")
			continue
		}

		if value == nil {
			if spec.Nullable {
				result[key] = nil
			} else {
				report.AddEntry(fieldPath, ActionRemoved, "null not allowed")
			}
			continue
		}

		sanitizedValue := sanitizeValue(value, spec, fieldPath, report)
		if sanitizedValue != nil {
			result[key] = sanitizedValue
		}
	}

	return result
}

func sanitizeValue(value any, spec *FieldSpec, path string, report *ValidationReport) any {
	if value == nil {
		return nil
	}

	if spec.Type == FieldTypeAny {
		return sanitizeAnyValue(value, path, report)
	}

	switch spec.Type {
	case FieldTypeObject:
		if obj, ok := value.(map[string]any); ok {
			if spec.Children != nil {
				return sanitizeObject(obj, spec.Children, path, report)
			}
			return sanitizeAnyValue(obj, path, report)
		}
		report.AddEntry(path, ActionRemoved, "expected object")
		return nil

	case FieldTypeArray:
		if arr, ok := value.([]any); ok {
			return sanitizeArray(arr, spec.Items, path, report)
		}
		report.AddEntry(path, ActionRemoved, "expected array")
		return nil

	case FieldTypeString:
		if str, ok := value.(string); ok {
			return str
		}
		report.AddEntry(path, ActionRemoved, "expected string")
		return nil

	case FieldTypeNumber:
		switch v := value.(type) {
		case float64, float32, int, int32, int64, uint, uint32, uint64:
			return v
		}
		report.AddEntry(path, ActionRemoved, "expected number")
		return nil

	case FieldTypeBoolean:
		if b, ok := value.(bool); ok {
			return b
		}
		report.AddEntry(path, ActionRemoved, "expected boolean")
		return nil
	}

	return value
}

func sanitizeArray(arr []any, itemSpec *FieldSpec, path string, report *ValidationReport) []any {
	if len(arr) == 0 {
		return arr
	}

	result := make([]any, 0, len(arr))
	for i, item := range arr {
		if item == nil {
			continue
		}

		itemPath := path + "[" + itoa(i) + "]"

		if itemSpec != nil {
			sanitized := sanitizeValue(item, itemSpec, itemPath, report)
			if sanitized != nil {
				result = append(result, sanitized)
			}
		} else {
			sanitized := sanitizeAnyValue(item, itemPath, report)
			if sanitized != nil {
				result = append(result, sanitized)
			}
		}
	}
	return result
}

func sanitizeAnyValue(value any, path string, report *ValidationReport) any {
	switch v := value.(type) {
	case map[string]any:
		return sanitizeAnyObject(v, path, report)
	case []any:
		return sanitizeAnyArray(v, path, report)
	case string:
		if v == "[undefined]" || v == "undefined" {
			report.AddEntry(path, ActionRemoved, "undefined value")
			return nil
		}
		return v
	default:
		return v
	}
}

func sanitizeAnyObject(obj map[string]any, path string, report *ValidationReport) map[string]any {
	result := make(map[string]any, len(obj))
	for k, v := range obj {
		fieldPath := k
		if path != "" {
			fieldPath = path + "." + k
		}

		if v == nil {
			report.AddEntry(fieldPath, ActionRemoved, "null value in untyped object")
			continue
		}

		sanitized := sanitizeAnyValue(v, fieldPath, report)
		if sanitized != nil {
			result[k] = sanitized
		}
	}
	return result
}

func sanitizeAnyArray(arr []any, path string, report *ValidationReport) []any {
	result := make([]any, 0, len(arr))
	for i, item := range arr {
		if item == nil {
			continue
		}
		itemPath := path + "[" + itoa(i) + "]"
		sanitized := sanitizeAnyValue(item, itemPath, report)
		if sanitized != nil {
			result = append(result, sanitized)
		}
	}
	return result
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	digits := make([]byte, 0, 10)
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

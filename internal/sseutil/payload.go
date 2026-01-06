package sseutil

import (
	"strings"

	"github.com/nghyane/llm-mux/internal/config"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ApplyPayloadConfig applies configuration rules (defaults and overrides) to payload.
func ApplyPayloadConfig(cfg *config.Config, model string, payload []byte) []byte {
	return ApplyPayloadConfigWithRoot(cfg, model, "", "", payload)
}

// ApplyPayloadConfigWithRoot applies configuration with custom root path and protocol.
func ApplyPayloadConfigWithRoot(cfg *config.Config, model, protocol, root string, payload []byte) []byte {
	if cfg == nil || len(payload) == 0 {
		return payload
	}
	rules := cfg.Payload
	if len(rules.Default) == 0 && len(rules.Override) == 0 {
		return payload
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return payload
	}
	out := payload

	// Apply defaults (only if path doesn't exist)
	for i := range rules.Default {
		rule := &rules.Default[i]
		if !payloadRuleMatchesModel(rule, model, protocol) {
			continue
		}
		for path, value := range rule.Params {
			fullPath := buildPayloadPath(root, path)
			if fullPath == "" {
				continue
			}
			if gjson.GetBytes(out, fullPath).Exists() {
				continue
			}
			updated, errSet := sjson.SetBytes(out, fullPath, value)
			if errSet != nil {
				continue
			}
			out = updated
		}
	}

	// Apply overrides (always set)
	for i := range rules.Override {
		rule := &rules.Override[i]
		if !payloadRuleMatchesModel(rule, model, protocol) {
			continue
		}
		for path, value := range rule.Params {
			fullPath := buildPayloadPath(root, path)
			if fullPath == "" {
				continue
			}
			updated, errSet := sjson.SetBytes(out, fullPath, value)
			if errSet != nil {
				continue
			}
			out = updated
		}
	}
	return out
}

func payloadRuleMatchesModel(rule *config.PayloadRule, model, protocol string) bool {
	if rule == nil || len(rule.Models) == 0 {
		return false
	}
	for _, entry := range rule.Models {
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			continue
		}
		if ep := strings.TrimSpace(entry.Protocol); ep != "" && protocol != "" && !strings.EqualFold(ep, protocol) {
			continue
		}
		if MatchModelPattern(name, model) {
			return true
		}
	}
	return false
}

func buildPayloadPath(root, path string) string {
	r := strings.TrimSpace(root)
	p := strings.TrimSpace(path)
	if r == "" {
		return p
	}
	if p == "" {
		return r
	}
	return r + "." + strings.TrimPrefix(p, ".")
}

// MatchModelPattern checks if model matches a glob-style pattern.
// Supports: exact match, "*" (all), "*suffix", "prefix*", "*contains*"
func MatchModelPattern(pattern, model string) bool {
	pattern = strings.TrimSpace(pattern)
	model = strings.TrimSpace(model)
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}

	// Wildcard matching using simple state machine
	pi, si := 0, 0
	starIdx := -1
	matchIdx := 0
	for si < len(model) {
		if pi < len(pattern) && pattern[pi] == model[si] {
			pi++
			si++
			continue
		}
		if pi < len(pattern) && pattern[pi] == '*' {
			starIdx = pi
			matchIdx = si
			pi++
			continue
		}
		if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
			continue
		}
		return false
	}
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}

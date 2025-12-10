// Package registry provides model family definitions for cross-provider routing.
// Model families allow clients to use a canonical model name (e.g., "claude-sonnet-4-5")
// and have it automatically routed to the appropriate provider-specific model ID.
package registry

import (
	"math/rand"
	"sort"
	"sync"
)

// FamilyMember represents a provider-specific model within a family.
type FamilyMember struct {
	Provider string // Provider type (e.g., "kiro", "antigravity", "claude")
	ModelID  string // Provider-specific model ID
	Priority int    // Priority level (1 = highest). Same priority = load balanced
}

// ModelFamilies maps canonical model names to their provider-specific variants.
// Only define families where model IDs DIFFER between providers.
// Models with same ID across providers don't need family definitions.
var ModelFamilies = map[string][]FamilyMember{
	// Claude Sonnet 4.5 family - IDs differ between providers
	"claude-sonnet-4-5": {
		{Provider: "kiro", ModelID: "claude-sonnet-4-5", Priority: 1},
		{Provider: "antigravity", ModelID: "gemini-claude-sonnet-4-5", Priority: 1},
		{Provider: "claude", ModelID: "claude-sonnet-4-5-20250929", Priority: 2},
	},
	"claude-sonnet-4-5-thinking": {
		{Provider: "claude", ModelID: "claude-sonnet-4-5-thinking", Priority: 1},
		{Provider: "antigravity", ModelID: "gemini-claude-sonnet-4-5-thinking", Priority: 2},
	},

	// Claude Opus 4.5 family
	"claude-opus-4-5": {
		{Provider: "claude", ModelID: "claude-opus-4-5-20251101", Priority: 1},
		{Provider: "kiro", ModelID: "claude-opus-4-5-20251101", Priority: 2},
	},
	"claude-opus-4-5-thinking": {
		{Provider: "antigravity", ModelID: "gemini-claude-opus-4-5-thinking", Priority: 1},
		{Provider: "claude", ModelID: "claude-opus-4-5-thinking", Priority: 2},
	},

	// Claude Sonnet 4 family
	"claude-sonnet-4": {
		{Provider: "kiro", ModelID: "claude-sonnet-4-20250514", Priority: 1},
		{Provider: "claude", ModelID: "claude-sonnet-4-20250514", Priority: 2},
	},

	// Claude 3.7 Sonnet family
	"claude-3-7-sonnet": {
		{Provider: "kiro", ModelID: "claude-3-7-sonnet-20250219", Priority: 1},
		{Provider: "claude", ModelID: "claude-3-7-sonnet-20250219", Priority: 2},
	},

	// GPT-5.1 Codex Max family
	"gpt-5.1-codex-max": {
		{Provider: "github-copilot", ModelID: "gpt-5.1-codex-max", Priority: 1},
		{Provider: "openai", ModelID: "gpt-5.1-codex-max", Priority: 2},
	},

	// Note: Gemini models removed - they have same ID across providers,
	// so normal routing works without family translation.
}

// Pre-built indexes for O(1) lookup - initialized once
var (
	translationIndex map[string]map[string]string // [canonical][provider] -> modelID
	indexOnce        sync.Once
)

func buildIndexes() {
	translationIndex = make(map[string]map[string]string, len(ModelFamilies))
	for canonical, members := range ModelFamilies {
		providerMap := make(map[string]string, len(members))
		for _, m := range members {
			providerMap[m.Provider] = m.ModelID
		}
		translationIndex[canonical] = providerMap
	}
}

func ensureIndexes() {
	indexOnce.Do(buildIndexes)
}

// IsCanonicalID checks if the given ID is a canonical family name.
func IsCanonicalID(modelID string) bool {
	_, ok := ModelFamilies[modelID]
	return ok
}

// ResolveAllProviders returns all available providers for a canonical model,
// sorted by priority. Load balancing happens at execution layer.
func ResolveAllProviders(canonicalID string, availableProviders []string) (providers []string, found bool) {
	family, ok := ModelFamilies[canonicalID]
	if !ok {
		return nil, false
	}

	// Fast path: single provider available
	if len(availableProviders) == 1 {
		for _, member := range family {
			if member.Provider == availableProviders[0] {
				return []string{member.Provider}, true
			}
		}
		return nil, false
	}

	// Create set for O(1) lookup
	availableSet := make(map[string]struct{}, len(availableProviders))
	for _, p := range availableProviders {
		availableSet[p] = struct{}{}
	}

	// Single pass: collect and group by priority
	type priorityGroup struct {
		priority int
		members  []string
	}
	groups := make([]priorityGroup, 0, 3) // Most families have 2-3 priority levels
	groupIdx := make(map[int]int)         // priority -> index in groups

	for _, member := range family {
		if _, ok := availableSet[member.Provider]; !ok {
			continue
		}
		if idx, exists := groupIdx[member.Priority]; exists {
			groups[idx].members = append(groups[idx].members, member.Provider)
		} else {
			groupIdx[member.Priority] = len(groups)
			groups = append(groups, priorityGroup{
				priority: member.Priority,
				members:  []string{member.Provider},
			})
		}
	}

	if len(groups) == 0 {
		return nil, false
	}

	// Sort by priority (typically 2-3 groups, fast)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].priority < groups[j].priority
	})

	// Build result with shuffle for same priority
	result := make([]string, 0, len(family))
	for _, g := range groups {
		if len(g.members) > 1 {
			rand.Shuffle(len(g.members), func(i, j int) {
				g.members[i], g.members[j] = g.members[j], g.members[i]
			})
		}
		result = append(result, g.members...)
	}

	return result, true
}

// TranslateModelForProvider translates a canonical model ID to the provider-specific ID.
// Uses pre-built index for O(1) lookup.
func TranslateModelForProvider(canonicalID, provider string) string {
	ensureIndexes()

	if providerMap, ok := translationIndex[canonicalID]; ok {
		if modelID, ok := providerMap[provider]; ok {
			return modelID
		}
	}
	return canonicalID
}

// GetFamilyMembers returns all members of a model family, sorted by priority.
// Returns nil if the family doesn't exist.
func GetFamilyMembers(canonicalID string) []FamilyMember {
	family, ok := ModelFamilies[canonicalID]
	if !ok {
		return nil
	}

	// Return a copy sorted by priority
	sorted := make([]FamilyMember, len(family))
	copy(sorted, family)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	return sorted
}

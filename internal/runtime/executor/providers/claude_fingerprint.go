package providers

import (
	"hash/fnv"

	"github.com/nghyane/llm-mux/internal/provider"
	"github.com/nghyane/llm-mux/internal/runtime/executor"
)

// claudeFingerprint groups the identifying headers that should vary per
// account to reduce fingerprint correlation when multiplexing subscriptions.
type claudeFingerprint struct {
	UserAgent      string
	PackageVersion string // X-Stainless-Package-Version
	RuntimeVersion string // X-Stainless-Runtime-Version (Node.js version)
	OS             string // X-Stainless-Os
	Arch           string // X-Stainless-Arch
}

// defaultClaudeFingerprints contains realistic Claude CLI client fingerprints
// representing different environments. Each account is assigned one
// deterministically based on its ID so the same account always presents
// the same identity.
var defaultClaudeFingerprints = []claudeFingerprint{
	{
		UserAgent:      "claude-cli/1.0.83 (external, cli)",
		PackageVersion: "0.55.1",
		RuntimeVersion: "v24.3.0",
		OS:             "MacOS",
		Arch:           "arm64",
	},
	{
		UserAgent:      "claude-cli/1.0.83 (external, cli)",
		PackageVersion: "0.55.1",
		RuntimeVersion: "v22.12.0",
		OS:             "Linux",
		Arch:           "x64",
	},
	{
		UserAgent:      "claude-cli/1.0.82 (external, cli)",
		PackageVersion: "0.54.2",
		RuntimeVersion: "v22.17.0",
		OS:             "MacOS",
		Arch:           "x64",
	},
	{
		UserAgent:      "claude-cli/1.0.80 (external, cli)",
		PackageVersion: "0.53.0",
		RuntimeVersion: "v20.18.0",
		OS:             "Windows",
		Arch:           "x64",
	},
	{
		UserAgent:      "claude-cli/1.0.83 (external, cli)",
		PackageVersion: "0.55.1",
		RuntimeVersion: "v22.11.0",
		OS:             "Linux",
		Arch:           "arm64",
	},
}

// resolveClaudeFingerprint returns the fingerprint for the given auth entry.
// Resolution order:
//  1. Per-field overrides from auth Attributes (user_agent, stainless_os, etc.)
//  2. Per-field overrides from auth Metadata
//  3. Deterministic preset selection based on auth ID hash
func resolveClaudeFingerprint(auth *provider.Auth) claudeFingerprint {
	fp := selectFingerprintByAuthID(auth)

	if auth == nil {
		return fp
	}

	// Allow per-field overrides from auth attributes / metadata.
	if ua := executor.AttrStringValue(auth.Attributes, "user_agent"); ua != "" {
		fp.UserAgent = ua
	} else if ua := executor.MetaStringValue(auth.Metadata, "user_agent"); ua != "" {
		fp.UserAgent = ua
	}
	if v := executor.AttrStringValue(auth.Attributes, "stainless_os"); v != "" {
		fp.OS = v
	}
	if v := executor.AttrStringValue(auth.Attributes, "stainless_arch"); v != "" {
		fp.Arch = v
	}
	if v := executor.AttrStringValue(auth.Attributes, "stainless_package_version"); v != "" {
		fp.PackageVersion = v
	}
	if v := executor.AttrStringValue(auth.Attributes, "stainless_runtime_version"); v != "" {
		fp.RuntimeVersion = v
	}

	return fp
}

// selectFingerprintByAuthID deterministically picks a preset using a hash
// of the auth ID so the same credential always presents the same identity.
func selectFingerprintByAuthID(auth *provider.Auth) claudeFingerprint {
	if auth == nil || auth.ID == "" {
		return defaultClaudeFingerprints[0]
	}
	h := fnv.New32a()
	h.Write([]byte(auth.ID))
	idx := int(h.Sum32()) % len(defaultClaudeFingerprints)
	return defaultClaudeFingerprints[idx]
}
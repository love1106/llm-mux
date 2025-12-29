package executor

import (
	"strings"
	"time"

	"github.com/nghyane/llm-mux/sdk/cliproxy/auth"
)

type CredExtractorConfig struct {
	MetadataTokenKey    string
	MetadataURLKey      string
	TokenPrefix         string
	URLTransformFunc    func(string) string
	CheckNestedTokenMap bool
	TrimWhitespace      bool
}

func ExtractCreds(a *auth.Auth, cfg CredExtractorConfig) (token, url string) {
	if a == nil {
		return "", ""
	}

	trim := func(s string) string {
		if cfg.TrimWhitespace {
			return strings.TrimSpace(s)
		}
		return s
	}

	if a.Attributes != nil {
		token = trim(a.Attributes["api_key"])
		url = trim(a.Attributes["base_url"])
	}

	if token == "" && a.Metadata != nil {
		if v, ok := a.Metadata[cfg.MetadataTokenKey].(string); ok {
			token = trim(v)
		}

		if token == "" && cfg.CheckNestedTokenMap {
			if tokenMap, ok := a.Metadata["token"].(map[string]any); ok {
				if v, ok := tokenMap["access_token"].(string); ok && v != "" {
					token = trim(v)
				}
			}
		}
	}

	if url == "" && a.Metadata != nil && cfg.MetadataURLKey != "" {
		if v, ok := a.Metadata[cfg.MetadataURLKey].(string); ok {
			url = trim(v)
			if cfg.URLTransformFunc != nil {
				url = cfg.URLTransformFunc(url)
			}
		}
	}

	if token != "" && cfg.TokenPrefix != "" {
		token = cfg.TokenPrefix + token
	}

	return token, url
}

func ExtractRefreshToken(auth *auth.Auth) (string, bool) {
	if auth == nil || auth.Metadata == nil {
		return "", false
	}
	if v, ok := auth.Metadata["refresh_token"].(string); ok && strings.TrimSpace(v) != "" {
		return v, true
	}
	return "", false
}

func UpdateRefreshMetadata(auth *auth.Auth, updates map[string]any, providerType string) {
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	for k, v := range updates {
		if v != nil && v != "" {
			auth.Metadata[k] = v
		}
	}
	auth.Metadata["type"] = providerType
	auth.Metadata["last_refresh"] = time.Now().Format(time.RFC3339)
}

var (
	ClaudeCredsConfig = CredExtractorConfig{
		MetadataTokenKey: "access_token",
	}

	CodexCredsConfig = CredExtractorConfig{
		MetadataTokenKey: "access_token",
	}

	ClineCredsConfig = CredExtractorConfig{
		MetadataTokenKey: "access_token",
		TokenPrefix:      "workos:",
	}

	QwenCredsConfig = CredExtractorConfig{
		MetadataTokenKey: "access_token",
		MetadataURLKey:   "resource_url",
		URLTransformFunc: func(url string) string {
			return "https://" + url + "/v1"
		},
	}

	IFlowCredsConfig = CredExtractorConfig{
		MetadataTokenKey: "api_key",
		MetadataURLKey:   "base_url",
		TrimWhitespace:   true,
	}

	GeminiCredsConfig = CredExtractorConfig{
		MetadataTokenKey:    "access_token",
		CheckNestedTokenMap: true,
	}
)

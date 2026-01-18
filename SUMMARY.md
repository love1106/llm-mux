# llm-mux Architecture Summary

## Overview

llm-mux is an AI Gateway that converts subscription-based LLM accounts (Claude Pro, GitHub Copilot, Gemini, etc.) into standard OpenAI-compatible API endpoints. It allows developers to use their existing subscriptions with any OpenAI-compatible tool.

## Core Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Client App     │────▶│   llm-mux       │────▶│  LLM Provider   │
│  (Cursor, etc.) │◀────│   :8317         │◀────│  (Claude, etc.) │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                              │
                              ▼
                        ┌───────────┐
                        │ Auth Files│
                        │ (tokens)  │
                        └───────────┘
```

## Directory Structure

```
llm-mux/
├── cmd/server/          # Entry point
├── internal/
│   ├── api/             # HTTP server & routes
│   │   ├── handlers/    # Format-specific handlers
│   │   │   ├── format/  # OpenAI, Gemini, Ollama, Claude
│   │   │   └── management/  # Admin API
│   │   ├── middleware/  # Auth, logging, rate limiting
│   │   └── modules/     # Amp CLI integration
│   ├── auth/            # Provider authentication
│   │   ├── claude/      # Anthropic OAuth
│   │   ├── copilot/     # GitHub device flow
│   │   ├── gemini/      # Google OAuth
│   │   ├── codex/       # OpenAI Codex OAuth
│   │   └── login/       # Unified login manager
│   ├── provider/        # Provider orchestration
│   │   ├── manager.go   # Request routing & retry
│   │   ├── quota_manager.go  # Rate limit tracking
│   │   └── auth_registry.go  # Auth entry management
│   ├── runtime/executor/    # Request execution
│   │   └── providers/   # Provider-specific executors
│   ├── translator/      # Format conversion (IR-based)
│   └── config/          # YAML configuration
└── docs/                # Documentation
```

## Key Components

### 1. Authentication (`internal/auth/`)

Each provider has its own OAuth implementation:

| Provider | Auth Method | Files |
|----------|-------------|-------|
| Claude | OAuth2 + PKCE | `claude/anthropic_auth.go` |
| Copilot | GitHub Device Flow | `copilot/auth.go` |
| Gemini | Google OAuth2 | `gemini/gemini_auth.go` |
| Codex | OpenAI OAuth2 | `codex/openai_auth.go` |

**Flow**:
1. `llm-mux login <provider>` starts OAuth
2. Local callback server receives token
3. Token persisted to `~/.llm-mux/auth/<provider>.json`
4. Token auto-refreshed when expired

### 2. API Server (`internal/api/`)

Exposes multiple API formats on single port (default `:8317`):

| Endpoint | Format | Handler |
|----------|--------|---------|
| `/v1/chat/completions` | OpenAI | `openai_handlers.go` |
| `/v1/messages` | Anthropic | `code_handlers.go` |
| `/v1beta/models/*` | Gemini | `gemini_handlers.go` |
| `/api/chat` | Ollama | `ollama_handlers.go` |

### 3. Provider Manager (`internal/provider/manager.go`)

Orchestrates request execution:

```go
// Simplified flow
func (m *Manager) Execute(ctx, req, opts) (Response, error) {
    // 1. Select provider based on model & availability
    providers := m.selector.SelectProviders(model)

    // 2. Try each provider with retry logic
    for _, provider := range providers {
        auth := m.registry.PickAuth(provider, model)
        resp, err := m.executor.Execute(ctx, auth, req)
        if err == nil {
            return resp, nil
        }
        // Handle quota errors, retry next provider
    }
}
```

### 4. Quota Management (`internal/provider/quota_manager.go`)

Tracks rate limits per auth entry:
- Monitors token usage
- Implements cooldown on quota exceeded
- Supports learned limits from actual usage

### 5. Format Translation (`internal/translator/`)

Converts between API formats using Intermediate Representation (IR):

```
OpenAI Request → IR → Claude Request
                  ↓
Claude Response → IR → OpenAI Response
```

## Configuration

`config.yaml`:
```yaml
port: 8317
auth-dir: ~/.llm-mux/auth
debug: false

# Provider-specific settings
providers:
  - name: gemini
    type: gemini
  - name: claude
    type: anthropic

# Request routing
routing:
  default-provider: gemini
```

## Request Lifecycle

1. **Receive**: Client sends OpenAI-format request
2. **Authenticate**: Validate client API key (if configured)
3. **Route**: Select provider based on model name
4. **Translate**: Convert to provider's native format
5. **Execute**: Send to provider with stored OAuth token
6. **Retry**: On failure, try next available auth/provider
7. **Respond**: Translate response back to client format

## Multi-Account Load Balancing

```
                    ┌─── claude-account-1.json
Model: claude-4 ───▶├─── claude-account-2.json
                    └─── claude-account-3.json
```

When one account hits quota, automatically switches to next available.

## Security Model

- OAuth tokens stored locally with `0600` permissions
- No credentials transmitted externally
- Optional API key authentication for proxy access
- Management API protected by secret key

## RUN
./llm-mux serve --debug
./llm-mux serve

## Account Store
~/.config/llm-mux/auth/
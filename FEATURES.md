# LLM-MUX Feature List

> AI Gateway for Subscription-Based LLMs - Turn your Claude Pro, GitHub Copilot, and Gemini subscriptions into standard LLM APIs.

## 1. Multi-Provider Support

| Provider | Auth Type | Description |
|----------|-----------|-------------|
| **Claude** (Anthropic) | OAuth | Claude Pro/Max subscription |
| **GitHub Copilot** | OAuth | GitHub Copilot subscription |
| **Gemini** (Google AI Studio) | OAuth | Google AI Studio access |
| **Gemini CLI** | OAuth | Gemini CLI integration |
| **Antigravity** (Google) | OAuth | Google Antigravity service |
| **Codex** (OpenAI) | OAuth | OpenAI Codex access |
| **Qwen** | OAuth | Alibaba Qwen models |
| **Kiro** | OAuth | Kiro AI service |
| **iFlow** | OAuth | iFlow AI service |
| **Cline** | OAuth | Cline AI service |
| **Vertex AI** | API Key | Google Cloud Vertex AI |
| **OpenAI Compatible** | API Key | Any OpenAI-compatible endpoint |

## 2. Multi-Format API Endpoints

### OpenAI Format (`/v1`)
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions
- `POST /v1/responses` - Response API
- `GET /v1/models` - List models

### Anthropic/Claude Format (`/v1`)
- `POST /v1/messages` - Claude messages API
- `POST /v1/messages/count_tokens` - Token counting

### Gemini Format (`/v1beta`)
- `POST /v1beta/models/:action` - Gemini actions
- `GET /v1beta/models/:action` - Gemini queries
- `GET /v1beta/models` - List Gemini models

### Ollama Format (`/api`, `/ollama/api`)
- `POST /api/chat` - Ollama chat
- `POST /api/generate` - Ollama generate
- `GET /api/tags` - List Ollama models
- `POST /api/show` - Model info
- `GET /api/version` - API version

## 3. Management API

Base path: `/v1/management/*` (requires management key)

### Configuration
- `GET/PUT /config` - Runtime config
- `GET/PUT /config.yaml` - Full YAML config

### Authentication
- `GET/POST/DELETE /auth-files` - Manage auth files
- `POST /auth-files/refresh` - Force token refresh
- `POST /auth-files/import` - Import credentials
- `GET/PUT/PATCH/DELETE /api-keys` - API key management

### OAuth
- `POST /oauth/start` - Start OAuth flow
- `GET /oauth/status/:state` - Check OAuth status
- `POST /oauth/cancel/:state` - Cancel OAuth

### Providers
- `GET/PUT/DELETE /providers` - Provider management
- `POST /vertex/import` - Import Vertex credentials

### Monitoring
- `GET /usage` - Usage statistics
- `GET/PUT /usage-statistics-enabled` - Toggle usage tracking
- `GET/DELETE /logs` - Server logs
- `GET /request-error-logs` - Error logs

### Settings
- `GET/PUT /debug` - Debug mode
- `GET/PUT /logging-to-file` - File logging
- `GET/PUT/DELETE /proxy-url` - Proxy configuration
- `GET/PUT /request-retry` - Retry settings
- `GET/PUT /max-retry-interval` - Max retry interval
- `GET/PUT /ws-auth` - WebSocket auth
- `GET/PUT /request-log` - Request logging

### Quota Management
- `GET/PUT /quota-exceeded/switch-project` - Auto switch project
- `GET/PUT /quota-exceeded/switch-preview-model` - Auto switch to preview
- `GET/PUT/PATCH/DELETE /oauth-excluded-models` - Excluded models

## 4. Core Features

### Load Balancing
- Multi-account rotation
- Weighted selection based on usage
- Automatic failover on quota exceeded
- Per-auth cooldown periods

### Token Management
- Background auto-refresh (configurable lead time)
- On-demand refresh during requests
- Unified refresh path (single source of truth)
- OAuth token rotation handling

### Format Translation
- IR (Intermediate Representation) based conversion
- Seamless format translation between providers
- Support for streaming and non-streaming

### Resilience
- Configurable retry with exponential backoff
- Streaming circuit breaker
- Request size limits (DoS protection)
- Graceful degradation

### Hot Reload
- File watcher for config changes
- Auth file hot-reload
- No restart required for config updates

### WebSocket Support
- Real-time streaming via WebSocket
- Configurable WebSocket auth
- Session management

### Usage Tracking
- SQLite backend (default)
- PostgreSQL backend (optional)
- Token counting and pricing
- Per-provider statistics

### Extended Thinking
- Claude extended thinking mode support
- Configurable thinking parameters

## 5. CLI Commands

| Command | Description |
|---------|-------------|
| `llm-mux` | Start server (default) |
| `llm-mux serve` | Start server explicitly |
| `llm-mux login <provider>` | OAuth login to provider |
| `llm-mux env` | Show environment info |
| `llm-mux import` | Import credentials |
| `llm-mux service install/start/stop` | System service management |
| `llm-mux update` | Self-update binary |
| `llm-mux version` | Show version |

## 6. Configuration

### Server
- `port` - Server port (default: 8317)
- `tls.enable/cert/key` - HTTPS configuration
- `debug` - Debug mode
- `logging-to-file` - Log to file

### Authentication
- `auth-dir` - Auth files directory
- `api-keys` - Client API keys
- `disable-auth` - Disable authentication
- `ws-auth` - Require WebSocket auth

### Limits
- `max-request-size` - Max request body (default: 50MB)
- `max-response-size` - Max response body (default: 100MB)
- `stream-timeout` - Streaming timeout

### Retry
- `request-retry` - Number of retries
- `max-retry-interval` - Max retry wait time

### Quota
- `quota-window` - Quota tracking window
- `quota-exceeded.switch-project` - Auto switch on quota hit
- `quota-exceeded.switch-preview-model` - Fallback to preview
- `disable-cooling` - Disable cooldown periods

### Usage
- `usage.enabled` - Enable usage tracking
- `usage.database-path` - SQLite path
- `usage.postgres-url` - PostgreSQL URL

### Proxy
- `proxy-url` - HTTP/HTTPS proxy for outbound requests

### AMP
- `ampcode.*` - Amp CLI compatibility settings

## 7. Integrations

| Tool | Status | Notes |
|------|--------|-------|
| Claude Code | ✅ Full | Bypass OAuth supported |
| OpenCode | ✅ Full | Native support |
| Cursor | ✅ Full | OpenAI compatible |
| Aider | ✅ Full | OpenAI compatible |
| Cline | ✅ Full | OpenAI/Anthropic compatible |
| Continue | ✅ Full | OpenAI compatible |
| LangChain | ✅ Full | OpenAI compatible |
| Open WebUI | ✅ Full | Ollama compatible |
| Amp CLI | ✅ Full | Native AMP module |

## 8. Security

- **API Key Authentication** - Multiple keys, per-request auth
- **Management Key** - Separate management endpoint protection
- **TLS Support** - HTTPS with custom certificates
- **Request Size Limits** - DoS/OOM protection
- **Rate Limiting** - Per-auth cooldown on quota hits
- **Secure Token Storage** - File-based with restricted permissions

## 9. Deployment

- **Binary** - Single binary, no dependencies
- **Docker** - Official Docker image
- **System Service** - systemd integration
- **Auto-Update** - Built-in self-update

---

*Last updated: 2026-01-25*

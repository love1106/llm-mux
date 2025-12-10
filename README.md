# llm-mux

Multi-provider LLM gateway with unified OpenAI-compatible API.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap nghyane/tap
brew install llm-mux

# Start as service
brew services start llm-mux
```

### Docker

```bash
docker pull nghyane/llm-mux
docker run -p 8318:8318 -v ./config.yaml:/llm-mux/config.yaml nghyane/llm-mux
```

### Build from source

```bash
go build -o llm-mux ./cmd/server/
./llm-mux -config config.yaml
```

## Features

- **Multi-provider support** - Gemini, Claude, OpenAI, Vertex AI, and more
- **Unified API** - OpenAI-compatible endpoints for all providers
- **OAuth authentication** - Gemini CLI, AI Studio, Antigravity, Claude, Codex
- **IR-based translation** - Canonical intermediate representation for clean format conversion
- **Load balancing** - Intelligent provider selection with performance tracking
- **Streaming** - SSE and NDJSON streaming support

## Configuration

```yaml
port: 8318
auth-dir: "~/.config/llm-mux/auth"
use-canonical-translator: true

# API keys (optional - can also use OAuth)
api-keys:
  - "your-api-key"
```

## Supported Providers

| Provider | Auth Method | Models |
|----------|-------------|--------|
| Gemini CLI | OAuth | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite, gemini-3-pro-preview |
| AI Studio | OAuth | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite, gemini-3-pro-preview, gemini-*-image |
| Antigravity | OAuth | gemini-claude-sonnet-4-5, gemini-claude-sonnet-4-5-thinking, gemini-claude-opus-4-5-thinking, + Gemini models |
| Claude | OAuth | claude-sonnet-4-5, claude-opus-4-5, claude-3-7-sonnet |
| Codex | OAuth | gpt-5.1, gpt-5.1-codex, gpt-5.1-codex-max |
| Kiro | OAuth | claude-sonnet-4-5, claude-opus-4-5 (via Amazon Q) |
| GitHub Copilot | OAuth | gpt-4.1, gpt-4o, gpt-5-mini, gpt-5.1-codex-max |
| iFlow | OAuth | qwen3-coder-plus, deepseek-r1, kimi-k2, and more |
| Vertex AI | API Key | Gemini models |
| OpenAI Compatible | API Key | Any OpenAI-compatible API |

## API Endpoints

```
POST /v1/chat/completions              # OpenAI Chat API
POST /v1/completions                   # OpenAI Completions API
GET  /v1/models                        # List available models
POST /v1beta/models/{model}:generate   # Gemini API
POST /api/chat                         # Ollama-compatible API
```

## Architecture

```
    OpenAI ─────┐                       ┌───── Gemini CLI
    Claude ─────┤                       ├───── AI Studio
    Ollama ─────┼─────► Canonical ◄─────┼───── Antigravity
    Gemini ─────┤       IR              ├───── Claude
      Kiro ─────┤                       ├───── Codex
     Cline ─────┘                       └───── Vertex
```

**Hub-and-spoke design** - All formats convert through unified IR, minimizing code duplication.

## Authentication

```bash
# OAuth login for providers
llm-mux --login              # Gemini CLI
llm-mux --antigravity-login  # Antigravity (Claude via Google)
llm-mux --claude-login       # Claude
llm-mux --codex-login        # OpenAI Codex
llm-mux --kiro-login         # Kiro (Amazon Q)
llm-mux --copilot-login      # GitHub Copilot
llm-mux --qwen-login         # Qwen
llm-mux --iflow-login        # iFlow
llm-mux --cline-login        # Cline
```

## SDK Usage

```go
import "github.com/nghyane/llm-mux/sdk/cliproxy"

svc, _ := cliproxy.NewBuilder().
    WithConfig(cfg).
    Build()

svc.Run(ctx)
```

## License

MIT License - see [LICENSE](LICENSE)

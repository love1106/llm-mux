# llm-mux

[![GitHub stars](https://img.shields.io/github/stars/nghyane/llm-mux?style=social)](https://github.com/nghyane/llm-mux)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)](https://github.com/nghyane/llm-mux)
[![Go Version](https://img.shields.io/badge/go-1.24-cyan)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-available-blue)](https://hub.docker.com/r/nghyane/llm-mux)

**Turn your existing AI subscriptions into API endpoints.** A local gateway that converts Claude Pro, GitHub Copilot, and Google Cloud Code subscriptions into OpenAI-compatible APIs.

> No separate API billing. No per-token charges. Use the subscriptions you already pay for.

---

## The Problem

You're paying for AI subscriptions but can't use them programmatically:

| Subscription | What You Pay | What You Get |
|:-------------|:-------------|:-------------|
| **Claude Pro/Max** | $20-100/month | Web/app access only |
| **GitHub Copilot** | $10-19/month | IDE integration only |
| **Google Cloud Code** | Free tier | Gemini in IDE only |

Meanwhile, API access costs **extra** — often $20+/month per provider, billed per token.

## The Solution

llm-mux runs locally and converts your subscription OAuth tokens into standard API endpoints:

```mermaid
graph LR
    subgraph "Your Tools"
        A[Cursor]
        B[Cline]
        C[Aider]
        D[Python/LangChain]
    end
    
    subgraph "llm-mux :8318"
        E[IR Translator]
    end
    
    subgraph "Your Subscriptions"
        F[Claude Pro/Max]
        G[GitHub Copilot]
        H[Google Cloud Code]
        I[More Providers]
    end
    
    A --> E
    B --> E
    C --> E
    D --> E
    E -->|OAuth| F
    E -->|OAuth| G
    E -->|OAuth| H
    E -->|OAuth| I
```

---

## Key Features

| Feature | Description |
|:--------|:------------|
| **Subscription-to-API** | Convert Claude, Copilot, and Gemini subscriptions into REST APIs |
| **Universal Format** | Single OpenAI-compatible endpoint for all providers |
| **Auto Token Refresh** | OAuth tokens refresh automatically — no manual intervention |
| **Load Balancing** | Distribute requests across multiple accounts |
| **Local & Private** | Runs on your machine; credentials never leave your system |
| **Multi-Format Support** | OpenAI, Claude, Gemini, and Ollama API formats |

---

## Installation

| Platform | Command |
|:---------|:--------|
| **macOS / Linux** | `curl -fsSL https://raw.githubusercontent.com/nghyane/llm-mux/main/install.sh \| bash` |
| **Windows** | `irm https://raw.githubusercontent.com/nghyane/llm-mux/main/install.ps1 \| iex` |
| **Docker** | `docker run -p 8318:8318 -v ~/.config/llm-mux:/root/.config/llm-mux nghyane/llm-mux` |

---

## Quick Start

### 1. Initialize

```bash
llm-mux --init
```

### 2. Connect Your Subscriptions

```bash
# Google (Gemini 2.5/3.0 via Cloud Code)
llm-mux --antigravity-login

# Claude Pro/Max subscription
llm-mux --claude-login

# GitHub Copilot subscription
llm-mux --copilot-login
```

### 3. Use the API

Point any OpenAI-compatible tool to `http://localhost:8318/v1`:

```bash
curl http://localhost:8318/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-pro",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Available Models

### Gemini (via Google OAuth)

| Model | Context | Login Command |
|:------|:--------|:--------------|
| `gemini-2.5-pro` | 1M tokens | `--login` or `--antigravity-login` |
| `gemini-2.5-flash` | 1M tokens | `--login` or `--antigravity-login` |
| `gemini-3-flash` | 1M tokens | `--antigravity-login` |
| `gemini-3-pro-preview` | 1M tokens | `--antigravity-login` |

### Claude (via Claude Pro/Max Subscription)

| Model | Context | Login Command |
|:------|:--------|:--------------|
| `claude-sonnet-4-20250514` | 200K tokens | `--claude-login` |
| `claude-opus-4-5-20251101` | 200K tokens | `--claude-login` |
| `claude-sonnet-4-5-thinking` | 200K tokens | `--claude-login` |

### GPT (via GitHub Copilot Subscription)

| Model | Context | Login Command |
|:------|:--------|:--------------|
| `gpt-4.1` | 128K tokens | `--copilot-login` |
| `gpt-4o` | 128K tokens | `--copilot-login` |
| `gpt-5` | 128K tokens | `--copilot-login` |
| `gpt-5.1-codex-max` | 128K tokens | `--copilot-login` |

### Other Providers

| Provider | Models | Login Command |
|:---------|:-------|:--------------|
| **OpenAI Codex** | GPT-5 series | `--codex-login` |
| **Qwen** | `qwen3-coder-plus`, `qwen3-coder-flash` | `--qwen-login` |
| **Amazon Q / Kiro** | Various | `--kiro-login` |

> List all available models: `curl http://localhost:8318/v1/models`

---

## Tool Integration

### Cursor

```
Settings → Models → OpenAI API Base URL → http://localhost:8318/v1
```

### Cline (VS Code)

```
Settings → API Provider → OpenAI Compatible → http://localhost:8318/v1
```

### Aider

```bash
aider --openai-api-base http://localhost:8318/v1 --model gemini-2.5-pro
```

### Continue.dev

```json
{
  "models": [{
    "provider": "openai",
    "model": "claude-sonnet-4-20250514",
    "apiBase": "http://localhost:8318/v1"
  }]
}
```

### Windsurf

```
Settings → AI Provider → Custom OpenAI → http://localhost:8318/v1
```

### Python / OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8318/v1",
    api_key="unused"  # Any non-empty string works
)

response = client.chat.completions.create(
    model="gemini-2.5-pro",
    messages=[{"role": "user", "content": "Explain quantum computing"}]
)

print(response.choices[0].message.content)
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8318/v1",
    model="claude-sonnet-4-20250514",
    api_key="unused"
)

response = llm.invoke("Write a haiku about programming")
```

---

## API Endpoints

llm-mux supports multiple API formats for maximum compatibility:

| Format | Endpoint | Use Case |
|:-------|:---------|:---------|
| **OpenAI** | `POST /v1/chat/completions` | Cursor, LangChain, most tools |
| **Claude** | `POST /v1/messages` | Anthropic SDK, Claude-native tools |
| **Gemini** | `POST /v1beta/models/:model:generateContent` | Google AI tools |
| **Ollama** | `POST /api/chat` | Ollama-compatible clients |
| **Responses** | `POST /v1/responses` | OpenAI Responses API format |

---

## All Login Commands

```bash
# Google Gemini
llm-mux --login              # Gemini CLI (2.5 Pro/Flash)
llm-mux --antigravity-login  # Google Cloud Code (2.5/3.0 series)

# Anthropic
llm-mux --claude-login       # Claude Pro/Max subscription

# OpenAI / GitHub
llm-mux --codex-login        # OpenAI Codex CLI
llm-mux --copilot-login      # GitHub Copilot (GPT-4.1, GPT-5, GPT-5.1)

# Amazon
llm-mux --kiro-login         # Amazon Q Developer / Kiro

# Others
llm-mux --qwen-login         # Qwen (Alibaba Cloud)
llm-mux --iflow-login        # iFlow
llm-mux --cline-login        # Cline API
```

---

## Advanced Configuration

<details>
<summary><strong>Load Balancing Multiple Accounts</strong></summary>

Add multiple accounts per provider to distribute load and increase quota:

```bash
# Login with multiple Google accounts
llm-mux --login              # Account 1
llm-mux --login              # Account 2 (different Google account)
llm-mux --antigravity-login  # Account 3
```

llm-mux automatically:
- Rotates requests across accounts
- Retries failed requests on alternate accounts
- Tracks quota usage per credential

</details>

<details>
<summary><strong>Management API & Web UI</strong></summary>

Access the web management interface at `http://localhost:8318/v0/management/`

Or use the REST API:

```bash
# Initialize and get your management key
llm-mux --init

# List all credentials
curl http://localhost:8318/v0/management/auths \
  -H "X-Management-Key: YOUR_KEY"

# Trigger OAuth flow via API
curl -X POST http://localhost:8318/v0/management/oauth/start \
  -H "X-Management-Key: YOUR_KEY" \
  -d '{"provider": "claude"}'
```

</details>

<details>
<summary><strong>Sync Credentials Across Machines</strong></summary>

Use GitStore to sync credentials via a private Git repository:

```bash
export GITSTORE_GIT_URL=https://github.com/you/llm-mux-credentials.git
export GITSTORE_GIT_USERNAME=your_username
export GITSTORE_GIT_TOKEN=your_personal_access_token
```

</details>

<details>
<summary><strong>Custom API Keys & Providers</strong></summary>

Add traditional API keys alongside OAuth credentials in `~/.config/llm-mux/config.yaml`:

```yaml
# Gemini API keys
gemini-api-key:
  - api-key: "AIzaSy..."
    base-url: "https://generativelanguage.googleapis.com"

# Claude API keys
claude-api-key:
  - api-key: "sk-ant-..."

# OpenAI-compatible providers (OpenRouter, etc.)
openai-compatibility:
  - name: "openrouter"
    base-url: "https://openrouter.ai/api/v1"
    api-key-entries:
      - api-key: "sk-or-..."
    models:
      - name: "anthropic/claude-3.5-sonnet"
        alias: "openrouter-claude"
```

</details>

<details>
<summary><strong>Service Management</strong></summary>

| Action | macOS | Linux | Windows |
|:-------|:------|:------|:--------|
| Start | `launchctl start com.llm-mux` | `systemctl --user start llm-mux` | `Start-ScheduledTask "llm-mux"` |
| Stop | `launchctl stop com.llm-mux` | `systemctl --user stop llm-mux` | `Stop-ScheduledTask "llm-mux"` |
| Logs | `tail -f ~/Library/Logs/llm-mux.log` | `journalctl --user -u llm-mux -f` | Event Viewer |

</details>

<details>
<summary><strong>Proxy Configuration</strong></summary>

Configure a proxy for all outbound requests:

```yaml
# In ~/.config/llm-mux/config.yaml
proxy-url: "socks5://user:pass@proxy.example.com:1080"
```

Or per-provider:

```yaml
gemini-api-key:
  - api-key: "AIzaSy..."
    proxy-url: "http://proxy.example.com:8080"
```

</details>

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        Your Application                         │
│                   (Cursor, Aider, Python, etc.)                 │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼ Intermediate Representation (IR)
┌─────────────────────────────────────────────────────────────────┐
│                         llm-mux :8318                           │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────────┐   │
│  │   Request   │───▶│  Translator  │───▶│  Provider Router │   │
│  │   Parser    │    │     (IR)     │    │  & Load Balancer │   │
│  └─────────────┘    └──────────────┘    └──────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                               │
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                  ▼
     ┌────────────┐     ┌────────────┐     ┌────────────┐
     │   Claude   │     │   Gemini   │     │  Copilot   │
     │    API     │     │    API     │     │    API     │
     └────────────┘     └────────────┘     └────────────┘
```

**Intermediate Representation (IR):** llm-mux translates requests through a unified format, enabling 2N translations instead of N² — one translation to IR, one from IR to each provider.

---

## Troubleshooting

| Issue | Solution |
|:------|:---------|
| "No credentials found" | Run the appropriate `--login` command |
| "Rate limit exceeded" | Add more accounts or wait for quota reset |
| "Model not found" | Check available models with `curl http://localhost:8318/v1/models` |
| OAuth window doesn't open | Check if port 8318 is available; try `llm-mux --port 8319` |

---

## Security Considerations

- **Local-only by default:** Management endpoints only accept localhost connections
- **Credentials stored locally:** OAuth tokens are saved in `~/.config/llm-mux/auth/`
- **No telemetry:** llm-mux does not collect or transmit usage data
- **Open source:** Full source code available for audit

---

## Keywords

**Subscription Conversion:** claude-pro, github-copilot, antigravity, subscription-to-api, gemini-2.5, gpt-5  
**AI Tools:** cursor, cline, aider, continue.dev, windsurf, vscode  
**API Integration:** openai-compatible, api-gateway, local-proxy, oauth  
**AI Models:** claude-sonnet-4, gemini-3-flash, gpt-5.1-codex-max  
**Use Cases:** ai-coding, developer-tools, langchain, python-sdk

---

## License

MIT License — see [LICENSE](LICENSE)

---

## Contributing

Contributions are welcome. Please open an issue first to discuss proposed changes.

---

<p align="center">
  <strong>If llm-mux saves you money, consider giving it a star.</strong>
  <br><br>
  <a href="https://github.com/nghyane/llm-mux">GitHub</a> ·
  <a href="https://github.com/nghyane/llm-mux/issues">Issues</a> ·
  <a href="https://github.com/nghyane/llm-mux/discussions">Discussions</a>
</p>

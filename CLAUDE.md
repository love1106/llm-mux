# CLAUDE.md

Project-specific instructions for Claude Code.

## Project Purpose

llm-mux is an AI Gateway that proxies and rotates multiple LLM subscription accounts (Claude Pro/Max, GitHub Copilot, Gemini, etc.) for clients. It allows multiple users/tools to share subscription-based LLM access through a unified API.

Key use case: Route requests from tools like Claude Code, OpenCode, Cursor, Aider through this proxy to leverage existing subscriptions instead of paying for API access.

**Current status**: Works for simple API requests. Compatibility with Claude Code / OpenCode (URL override) needs verification.

## Development Server

When asked to "start server", "start servers", or "run the server":
- Start Go API: `./llm-mux serve` or `go run . serve`
- Start UI dev server: `cd ui && npm run dev`
- **DO NOT** use Docker Compose unless explicitly requested

## Ports

- API: 8317
- UI (dev): 8318
- Management key: read from `~/.config/llm-mux/credentials.json` (do NOT set env var)

## Build Commands

```bash
# Build Go binary
go build -o llm-mux ./cmd/server

# Build and start server (dev)
make dev

# Build UI
cd ui && npm run build

# Run tests
go test ./...
```

## Project Structure

- `internal/` - Go backend code
- `ui/` - React frontend (Vite + TypeScript)
- `auth/` - OAuth token storage directory

# Phase 2: Go Server Embedding

## Context
Integrate the React build output into the Go server using embed.FS to serve the Management UI at `/ui/*`.

## Overview
Configure the Go server to embed and serve the React SPA, handling client-side routing correctly with index.html fallback for all non-asset routes.

## Requirements
- Use Go 1.16+ embed package
- Serve UI at `/ui/*` path prefix
- Fallback to index.html for client-side routes
- Preserve existing API routes
- Support development and production modes

## Implementation Steps

### 1. Create Embed Configuration
Create `internal/server/ui_embed.go`:
```go
package server

import (
    "embed"
    "io/fs"
    "net/http"
    "path"
    "strings"
)

//go:embed all:dist-ui
var uiFiles embed.FS

// GetUIFileSystem returns the embedded UI filesystem
func GetUIFileSystem() (fs.FS, error) {
    return fs.Sub(uiFiles, "dist-ui")
}
```

### 2. Create UI Handler
Create `internal/server/ui_handler.go`:
```go
package server

import (
    "fmt"
    "io"
    "io/fs"
    "mime"
    "net/http"
    "path/filepath"
    "strings"
)

type UIHandler struct {
    fileSystem fs.FS
    indexHTML  []byte
}

func NewUIHandler() (*UIHandler, error) {
    fsys, err := GetUIFileSystem()
    if err != nil {
        return nil, fmt.Errorf("failed to get UI filesystem: %w", err)
    }

    // Pre-load index.html for SPA fallback
    indexHTML, err := fs.ReadFile(fsys, "index.html")
    if err != nil {
        return nil, fmt.Errorf("failed to read index.html: %w", err)
    }

    return &UIHandler{
        fileSystem: fsys,
        indexHTML:  indexHTML,
    }, nil
}

func (h *UIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Remove /ui prefix
    uiPath := strings.TrimPrefix(r.URL.Path, "/ui")
    if uiPath == "" || uiPath == "/" {
        uiPath = "/index.html"
    }

    // Try to serve the file
    file, err := h.fileSystem.Open(strings.TrimPrefix(uiPath, "/"))
    if err != nil {
        // If file not found, serve index.html for SPA routing
        h.serveIndex(w, r)
        return
    }
    defer file.Close()

    // Get file info
    stat, err := file.Stat()
    if err != nil || stat.IsDir() {
        h.serveIndex(w, r)
        return
    }

    // Set content type
    ext := filepath.Ext(uiPath)
    contentType := mime.TypeByExtension(ext)
    if contentType != "" {
        w.Header().Set("Content-Type", contentType)
    }

    // Set cache headers for assets
    if strings.Contains(uiPath, "/assets/") {
        w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
    }

    // Serve the file
    io.Copy(w, file)
}

func (h *UIHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(h.indexHTML)
}
```

### 3. Update Main Server Router
Update server initialization to include UI handler:
```go
func (s *Server) setupRoutes() {
    // Existing API routes
    s.router.PathPrefix("/v1/management").Handler(s.managementHandler)
    s.router.PathPrefix("/v1/chat").Handler(s.chatHandler)

    // UI handler (must be after API routes)
    if !s.config.DisableUI {
        uiHandler, err := NewUIHandler()
        if err != nil {
            log.Printf("Warning: UI disabled due to error: %v", err)
        } else {
            s.router.PathPrefix("/ui").Handler(
                http.StripPrefix("/ui", uiHandler),
            )

            // Redirect root to UI
            s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
            })
        }
    }
}
```

### 4. Build Script Integration
Create `scripts/build-ui.sh`:
```bash
#!/bin/bash
set -e

echo "Building Management UI..."
cd ui

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm ci
fi

# Build the UI
npm run build

# The output will be in dist-ui/ as configured in vite.config.ts
echo "UI build complete in dist-ui/"
```

### 5. Update Makefile
Add UI build to main build process:
```makefile
.PHONY: ui
ui:
    @bash scripts/build-ui.sh

.PHONY: build
build: ui
    go build -o bin/llm-mux ./cmd/server

.PHONY: build-all
build-all: ui
    GOOS=linux GOARCH=amd64 go build -o bin/llm-mux-linux-amd64 ./cmd/server
    GOOS=darwin GOARCH=amd64 go build -o bin/llm-mux-darwin-amd64 ./cmd/server
    GOOS=windows GOARCH=amd64 go build -o bin/llm-mux-windows-amd64.exe ./cmd/server

.PHONY: dev
dev:
    # Run UI dev server in background
    cd ui && npm run dev &
    # Run Go server with hot reload
    air
```

### 6. Development Mode Support
Create `internal/server/ui_dev.go`:
```go
//go:build dev

package server

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

// In development, proxy to Vite dev server
func NewUIHandler() (http.Handler, error) {
    target, _ := url.Parse("http://localhost:5173")
    proxy := httputil.NewSingleHostReverseProxy(target)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.URL.Path = strings.TrimPrefix(r.URL.Path, "/ui")
        proxy.ServeHTTP(w, r)
    }), nil
}
```

### 7. CI/CD Integration
Update `.github/workflows/build.yml`:
```yaml
- name: Setup Node.js
  uses: actions/setup-node@v3
  with:
    node-version: '20'
    cache: 'npm'
    cache-dependency-path: ui/package-lock.json

- name: Build UI
  run: |
    cd ui
    npm ci
    npm run build

- name: Build Go binary
  run: go build -o llm-mux ./cmd/server
```

## Todo List
- [ ] Create embed.FS configuration for dist-ui directory
- [ ] Implement UIHandler with SPA routing logic
- [ ] Add cache headers for static assets
- [ ] Update main server router with UI routes
- [ ] Create build scripts for UI compilation
- [ ] Update Makefile with UI build targets
- [ ] Add development mode proxy support
- [ ] Configure CI/CD for UI builds
- [ ] Test embedded UI serving

## Success Criteria
- [ ] UI accessible at http://localhost:8317/ui
- [ ] Client-side routing works (refresh on any route)
- [ ] Static assets served with correct MIME types
- [ ] Assets cached with immutable headers
- [ ] Development mode proxies to Vite dev server
- [ ] Production build embeds UI in single binary
- [ ] Root path redirects to /ui/
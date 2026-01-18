# Embedding Vite React SPA in Go Server: Research Report

## Vite Build Output Structure (dist/)
- Default structure:
  ```
  dist/
  ├── assets/
  │   ├── vendor.[hash].js
  │   └── main.[hash].js
  └── index.html
  ```
- Supports customizing chunk and asset file names
- Root-level index.html always generated

## Go embed.FS Serving Strategy
### Routing Middleware Pattern
1. Check if requested file exists
2. Serve static files (CSS, JS) directly
3. Fallback to index.html for SPA routes

### Code Structure Example
```go
func serveIndex(fs embed.FS) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Serve index.html for SPA routes
        if _, err := fs.Open(r.URL.Path); os.IsNotExist(err) {
            indexContent, _ := fs.ReadFile("index.html")
            w.Write(indexContent)
        }
    }
}
```

## Production Optimization Tips
- Use Vite's built-in tree-shaking
- Configure base path to '/'
- Minify and compress static assets
- Use content hashing for cache busting

## Unresolved Questions
- Best practices for handling large asset files
- Performance implications of embed.FS vs file system

## Recommended Next Steps
1. Prototype Go server with embedded Vite build
2. Benchmark server performance
3. Test various routing scenarios
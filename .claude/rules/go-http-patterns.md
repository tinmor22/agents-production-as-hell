# Go HTTP Patterns

Canonical patterns for HTTP handlers, middleware, and server lifecycle. Copy-paste these exactly.

## Handler signature

```go
// internal/handler/<resource>.go
func (h *Handler) MethodResource(w http.ResponseWriter, r *http.Request) {
    var req RequestType
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    result, err := h.svc.DoSomething(r.Context(), req)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Health endpoint (required on every HTTP server)

```go
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "dev"})
}
```

## Router setup

```go
// internal/handler/router.go
func NewRouter(h *Handler, cfg config.Config) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/health", h.Health)
    // Register all routes here
    return middleware.CORS([]string{"http://localhost:5173", cfg.Origin})(mux)
}
```

- Use Go 1.22+ method-pattern routing: `"GET /path"`, `"POST /path"`, `"GET /path/{id}"`.
- Never use `http.DefaultServeMux`.
- Always wire CORS when a frontend exists.

## CORS middleware

```go
// internal/middleware/cors.go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            for _, allowed := range allowedOrigins {
                if origin == allowed {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    break
                }
            }
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Graceful shutdown (required on every HTTP server)

```go
// In cmd/<name>/main.go, after srv.ListenAndServe() goroutine:
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Printf("shutdown error: %v", err)
}
```

## Frontend embedding (single binary)

```go
// internal/handler/static.go
//go:embed all:../../web/dist
var webDist embed.FS

var webFS, _ = fs.Sub(webDist, "web/dist")
```

Wire in router: `mux.Handle("/", http.FileServer(http.FS(webFS)))`

## HTTP handler test pattern

```go
// internal/handler/<resource>_test.go
func TestHandler_ListItems(t *testing.T) {
    svc := &mockService{items: []Item{{ID: 1, Name: "test"}}}
    h := NewHandler(svc)
    req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
    w := httptest.NewRecorder()
    h.ListItems(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("want 200, got %d", w.Code)
    }
    var got []Item
    json.NewDecoder(w.Body).Decode(&got)
    // assert
}
```

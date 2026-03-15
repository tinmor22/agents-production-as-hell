# Go Coding Standards

These rules apply to all Go code written in this repository. No exceptions.

## Error handling

- Always wrap errors with context: `fmt.Errorf("<context>: %w", err)`
- Return typed sentinel errors for expected conditions: `var ErrNotFound = errors.New("<entity>: not found")`
- Never swallow errors silently. Never `_ = err`.

## Package structure

- `cmd/` — entry points only. One binary per subdirectory. No business logic.
- `internal/` — all reusable logic. Never import across service boundaries.
- `internal/config/` — env var loading. Only place where `os.Getenv` is allowed.
- `internal/handler/` — HTTP handlers. No direct DB calls — only service calls.
- `internal/<domain>/` — business logic per domain. No HTTP, no DB imports.
- `internal/db/` — database layer. One `<Entity>Store` per entity.
- Never use `http.DefaultServeMux`. Always use a named `http.NewServeMux()`.

## Functions

- Maximum 40 lines per function. Split if longer.
- No global mutable state outside `main.go`.
- Inject all dependencies via constructors: `func NewService(db *sql.DB) *Service`
- No direct `os.Getenv` calls outside `config.go`.
- No magic strings or magic numbers — use typed constants.

## Testing

- Table-driven tests for any function with more than 2 cases.
- DB tests: guard with `if testing.Short() { t.Skip() }` at the top.
- DB tests: open connection from `TEST_DSN` env var via a `testDB(t *testing.T) *sql.DB` helper.
- HTTP handler tests: use `httptest.NewRequest` + `httptest.NewRecorder`. Never start a real server.
- Never delete or weaken an assertion to make a test pass. Fix the implementation.
- Maximum 3 fix iterations per failing test before documenting as `tech_debt`.

## Go version and tooling

- Use Go 1.22+ method-pattern routing: `mux.HandleFunc("GET /path", handler)`.
- Run `go mod tidy` after every `go get` or new import addition.
- Run `go vet ./...` — must produce zero output.
- Use `make build` / `make test` / `make vet` if `Makefile` exists.

## HTTP responses

- Always set `Content-Type: application/json` before writing response body.
- Success: `json.NewEncoder(w).Encode(resp)`.
- Error: `{"error": "<message>"}` with appropriate HTTP status code.
- Decode request body: `json.NewDecoder(r.Body).Decode(&req)`.

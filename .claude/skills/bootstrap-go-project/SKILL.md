---
name: bootstrap-go-project
description: Scaffolds a new Go project from scratch: go.mod, Makefile, shared types, config package, entry point, and optionally database + React frontend. Use only once per project on the first Viktor run.
argument-hint: [module-name]
allowed-tools: Read, Write, Edit, Bash, Glob
user-invocable: false
---

Bootstrap the Go project infrastructure for module **$ARGUMENTS**.

**Check before every step — skip anything that already exists. Never re-initialize working state.**

Derive the binary name from the module path: last path segment of `$ARGUMENTS` (e.g. `github.com/acme/myapp` → `myapp`). Use this as `<name>` throughout.

---

## Step 1 — Go module

```bash
Glob("go.mod")
```
- Exists → read it, confirm module name matches `$ARGUMENTS`. If it matches, continue. If it doesn't match, stop and warn the user.
- Missing → `Bash("go mod init $ARGUMENTS")`

---

## Step 2 — .gitignore

```bash
Glob(".gitignore")
```
- Exists → `Read` it and `Edit` to append `.env` if not already present.
- Missing → create:

```
.env
bin/
*.test
```

---

## Step 3 — Makefile

```bash
Glob("Makefile")
```
- Exists → skip.
- Missing → create, replacing `<name>` with the binary name:

```makefile
APP := <name>

.PHONY: build test vet run dev frontend-install frontend-build frontend-test

build:
	go build -o bin/$(APP) ./cmd/$(APP)/...

test:
	go test ./...

vet:
	go vet ./...

run:
	./bin/$(APP)

dev:
	go run ./cmd/$(APP)/... serve

frontend-install:
	cd web && npm install

frontend-build:
	cd web && npm run build

frontend-test:
	cd web && npm test -- --run
```

---

## Step 4 — Shared types

```bash
Glob("internal/types/")
```

Write every type from `codebase_blueprint.key_types` to its exact package path. Rules:
- No `...` placeholders. No `// TODO`. Every field typed and tagged.
- JSON tags must match Go field names in snake_case: `json:"field_name"`.
- If `key_types` is empty, create `internal/types/types.go` with a comment placeholder.

---

## Step 5 — Config package

```bash
Glob("internal/config/config.go")
```
- Exists → skip.
- Missing → create. Include `DSN` only if `tech_stack.storage` is not `"none"` or `"filesystem"`:

```go
package config

import "os"

type Config struct {
	Port   string
	DSN    string // omit if storage = none/filesystem
	Env    string // "development" | "production"
	Origin string // allowed CORS origin for production
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return Config{
		Port:   port,
		DSN:    os.Getenv("DSN"),
		Env:    env,
		Origin: os.Getenv("ORIGIN"),
	}, nil
}
```

---

## Step 6 — Database (skip entirely if `tech_stack.storage` is `"none"` or `"filesystem"`)

Install deps first:
```bash
Bash("go get github.com/pressly/goose/v3 github.com/lib/pq github.com/joho/godotenv")
```

Then check and create each file only if missing:

**`migrations/001_init.sql`**
```sql
-- +goose Up
-- +goose StatementBegin
-- initial schema — add tables here
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- drop tables here
-- +goose StatementEnd
```

**`internal/db/db.go`**
```go
package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return db, nil
}
```

**`internal/db/migrate.go`**
```go
package db

import (
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	goose.SetDialect("postgres")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	return nil
}
```

**`docker-compose.yml`**
```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: app
      POSTGRES_DB: app
    ports:
      - "5432:5432"
```

**`.env.example`**
```
PORT=8080
ENV=development
DSN=postgres://app:app@localhost:5432/app?sslmode=disable
ORIGIN=http://localhost:5173
```

---

## Step 7 — Entry point

```bash
Glob("cmd/*/main.go")
```
- Exists → skip.
- Missing → create `cmd/<name>/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"<module>/internal/config"
	"<module>/internal/handler"
	// "<module>/internal/db" — uncomment if using DB
)

func main() {
	// Load .env in development only
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// db, err := db.Connect(cfg.DSN)
	// if err != nil { log.Fatalf("db: %v", err) }
	// if err := db.RunMigrations(db, "migrations"); err != nil {
	//     log.Fatalf("migrations: %v", err)
	// }

	h := handler.NewHandler(/* db */)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler.NewRouter(h, cfg),
	}

	fmt.Printf("listening on :%s (env=%s)\n", cfg.Port, cfg.Env)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
```

Replace `<module>` with the actual module path from `go.mod`.

---

## Step 8 — React frontend (skip entirely if no frontend component in `design_context.components`)

```bash
Glob("web/package.json")
```
- Exists → skip scaffold, but check deps below.
- Missing → scaffold:

```bash
Bash("npm create vite@latest web -- --template react-ts --yes")
Bash("cd web && npm install")
Bash("cd web && npm install @tanstack/react-query")
Bash("cd web && npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom")
```

Create these files only if missing:

**`web/src/api/client.ts`**
```ts
const BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as { error?: string }).error ?? res.statusText);
  }
  return res.json() as Promise<T>;
}
```

**`web/src/setupTests.ts`**
```ts
import "@testing-library/jest-dom";
```

**`web/.env.example`**
```
VITE_API_URL=http://localhost:8080
```

**`web/.env.local`** (copy from `.env.example` if missing — gitignored by Vite by default)

Add vitest config to `web/vite.config.ts` using `Edit` — append inside the `defineConfig({})` object:
```ts
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/setupTests.ts"],
  },
```

---

## Step 9 — Environment files

```bash
Glob(".env")
```
- Exists → skip.
- Missing → create from `.env.example` with local defaults:

```
PORT=8080
ENV=development
DSN=postgres://app:app@localhost:5432/app?sslmode=disable
ORIGIN=http://localhost:5173
```

---

## Step 10 — Install deps and verify

```bash
Bash("go mod tidy")
Bash("go build ./...")
```

Both must exit 0. Fix every compiler error before continuing.

If frontend was scaffolded:
```bash
Bash("cd web && npm run build")
```
Must exit 0.

---

## Step 11 — Smoke test (if HTTP server)

```bash
Bash("go run ./cmd/<name>/... serve &")
# capture PID
Bash("sleep 1 && curl -sf http://localhost:8080/api/health")
Bash("kill <PID>")
```

Must return `{"status":"ok",...}`. If connection refused: fix the server before declaring bootstrap complete.

---

## Done — report

Print a summary:

```
✓ Bootstrap complete: <module>
  go.mod:      <module> (go <version>)
  Makefile:    created | existed
  Types:       <N> types written
  Config:      internal/config/config.go
  Entry point: cmd/<name>/main.go
  Database:    yes (goose + postgres) | skipped
  Frontend:    yes (Vite + React + TS) | skipped
  Health:      GET /api/health → {"status":"ok"}

Next: run `make dev` to start, `make test` to verify.
```

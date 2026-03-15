---
name: explore-codebase-structure
description: Maps the existing codebase before writing any code. Use at the start of every implementation task to discover what already exists vs what needs to be created.
allowed-tools: Read, Glob, Grep, Bash
user-invocable: false
---

Explore the codebase and produce a structured map before writing a single line of code.

## Steps

1. **Go backend**
   - `Glob("**/*.go")` — list all Go source files.
   - `Read("go.mod")` — confirm module name and Go version. If missing, note: `go.mod absent`.
   - `Bash("go list ./...")` — list all packages (skip if `go.mod` missing).
   - `Glob("Makefile")` — if exists, read it. Use `make build` / `make test` / `make vet` instead of raw `go` commands.

2. **Migrations**
   - `Glob("migrations/**")` — list existing migration files.
   - Note the highest `NNN` prefix found (e.g. `003`). Next migration = `NNN + 1`.

3. **Frontend**
   - `Glob("web/package.json")` or `Glob("frontend/package.json")` — check for React setup.
   - If found: `Read("web/package.json")` — note framework, Vite version, test runner.
   - `Glob("web/src/**/*.tsx")` — list existing components and pages.

4. **Config & environment**
   - `Glob(".env.example")` — list expected env vars.
   - `Glob("internal/config/**")` — check if config package exists.

5. **What exists vs what is needed**
   - For each planned file/package: check with `Glob` or `Read` whether it already exists.
   - If it exists and looks correct: mark as **skip**.
   - If it exists but is incomplete: mark as **patch** (use `Edit`).
   - If it is absent: mark as **create** (use `Write`).

## Output

Produce a brief inventory before proceeding:

```
Module:        <module_name> (go <version>)
Makefile:      yes | no
Packages:      <list>
Migrations:    <count> existing, next = <NNN+1>
Frontend:      yes (<framework>) | no
Config pkg:    yes | no
---
Files to skip:   <list>
Files to patch:  <list>
Files to create: <list>
```

## Rules

- Never write a file without checking this inventory first.
- If `go.mod` is missing, run `go mod init <module_name>` before any other step.
- If `web/package.json` is missing and frontend is needed, scaffold with Vite before any frontend work.
- Prefer `Edit` over `Write` for any file that already exists.

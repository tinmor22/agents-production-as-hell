# Viktor Re-entrance Rules

Viktor may be run multiple times against the same codebase (resume, re-run, incremental slice). These rules prevent overwriting working code and ensure idempotency.

## Before writing any file

1. `Glob` or `Read` to check if the file already exists.
2. **File is correct and complete** → skip entirely. Do not touch it.
3. **File exists but incomplete or wrong** → use `Edit` to patch the specific section. Never `Write` the whole file.
4. **File is absent** → use `Write` to create it.

**Rule: prefer `Edit` over `Write` for any file that already exists.**

## Never re-initialize

| Condition | Command to skip |
|---|---|
| `go.mod` exists | `go mod init` |
| `web/package.json` exists | `npm create vite` |
| `migrations/001_*.sql` exists | recreating init migration |
| `go test ./...` passes for a package | touching that package |

## Migrations

- Before creating a new migration: `Glob("migrations/**")`, find the highest `NNN`. Next = `NNN + 1`.
- Never duplicate a migration number.
- Never modify an existing migration file — add a new one instead.

## Test loop

- If `go test ./...` already passes for a package: do not touch it. Move to the failing slice.
- If a previous run left a partial file: read it first, understand what's done, then patch with `Edit`.

## Failure recovery protocol

| Failure type | Response |
|---|---|
| Build error | Read exact error message. Fix the specific line. Rebuild. Do not rewrite the whole file. |
| Test failure | Read test output carefully. Fix the implementation, not the test (unless the test is provably wrong). |
| Import cycle | Move shared types to `internal/types`. Restructure packages. |
| Missing dependency | `go get <pkg>` then `go mod tidy`. |
| React type error | Check TypeScript types match Go JSON tags exactly. |

**Maximum 3 fix iterations per failure.** After 3: document in `tech_debt`, set `test_results.passed: false`, stop. Do not loop indefinitely.

## Idempotency checklist

Before starting any slice, verify:
- [ ] Explored codebase with `Glob`/`Read` (see `explore-codebase-structure` skill).
- [ ] Identified which files are skip / patch / create.
- [ ] Migration numbering confirmed (next NNN noted).
- [ ] Existing passing tests identified — will not be touched.

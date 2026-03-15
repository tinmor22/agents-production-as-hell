---
name: run-dod-check
description: Runs every Definition of Done shell command for a milestone slice and verifies all pass. Use after implementing each slice in the Viktor pipeline.
argument-hint: [slice-id]
allowed-tools: Read, Bash
user-invocable: false
---

Run the Definition of Done checks for slice **$ARGUMENTS** and verify all items pass.

## Steps

1. **Load the DoD** — read the `definition_of_done` array from the slice in Omar's plan. Each item is an exact shell command with an expected outcome.

2. **Run each command** in order via `Bash`. Capture exit code and output.

3. **Evaluate**:
   - Exit code 0 + expected output pattern → **PASS**
   - Any other result → **FAIL**

4. **On failure**:
   - Show the exact command, exit code, and last 20 lines of output.
   - Attempt to fix the root cause (read the error, fix the specific line, re-run).
   - **Maximum 3 fix iterations per failing check.** After 3: mark as `tech_debt`, set `test_results.passed: false`, stop.
   - Do NOT rewrite entire files to fix a single error.

5. **Report**:

```
DoD Check — slice: <slice-id>
✓ go build ./...
✓ go test ./...
✓ go vet ./...
✗ curl http://localhost:8080/api/health → connection refused
  Fix attempt 1/3: ...
```

## Standard checks (always run)

Beyond the slice-specific DoD, always verify:

| Command | Pass condition |
|---|---|
| `go build ./...` | Exit 0 |
| `go test ./...` | Exit 0, no FAIL lines |
| `go vet ./...` | Exit 0, no output |
| `go mod tidy` | Exit 0 (run after adding any new import) |
| `cd web && npm run build` | Exit 0 (if frontend was touched) |
| `cd web && npm test -- --run` | Exit 0 (if frontend was touched) |

## Health check (if HTTP server exists)

```bash
go run ./cmd/<name>/... serve &
SERVER_PID=$!
sleep 1
curl -s http://localhost:8080/api/health
kill $SERVER_PID
```

Expected: `{"status":"ok",...}`. If connection refused: server is broken — fix before marking slice done.

## Rules

- Use `make build` / `make test` / `make vet` if `Makefile` exists, raw `go` commands otherwise.
- Never mark a slice done if any DoD item fails.
- Never fix a test by deleting or weakening the assertion — fix the implementation.
- After 3 failed iterations: document in `tech_debt` with exact error message, location, and severity.

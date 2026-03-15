---
name: save-run-artifact
description: Saves a pipeline agent output to the run directory with consistent naming (runs/YYYY-MM-DD-NNN/NN-agent-output.json). Use after every agent completes in the orchestrator pipeline.
argument-hint: [run-dir] [stage-num] [agent-name]
allowed-tools: Read, Write, Glob
user-invocable: false
---

Save a pipeline artifact to disk with the correct naming convention.

## Arguments

- `$ARGUMENTS[0]` — run directory path (e.g. `runs/2026-03-13-001`)
- `$ARGUMENTS[1]` — two-digit stage number (e.g. `01`, `07`)
- `$ARGUMENTS[2]` — agent name slug (e.g. `nora`, `viktor`, `user-selection`)

If arguments are not provided, infer them from context (current run directory, current stage, current agent).

## Naming convention

| Pattern | Example |
|---|---|
| `{run-dir}/{NN}-{agent-name}-output.json` | `runs/2026-03-13-001/07-viktor-output.json` |
| Special: initial input | `runs/2026-03-13-001/00-input.json` |
| Special: user selection | `runs/2026-03-13-001/02a-user-selection.json` |

Full artifact map:

| File | Agent |
|---|---|
| `00-input.json` | Initial user input |
| `01-nora-output.json` | Nora |
| `02-leo-output.json` | Leo |
| `02a-user-selection.json` | User selection |
| `03-maya-output.json` | Maya |
| `04-sam-output.json` | Sam |
| `05-dani-output.json` | Dani |
| `06-omar-output.json` | Omar |
| `07-viktor-output.json` | Viktor |
| `08-priya-output.json` | Priya |
| `09-nate-output.json` | Nate |
| `10-ada-output.json` | Ada |

## Steps

1. Construct the file path: `{run-dir}/{NN}-{agent-name}-output.json`
2. Serialize the output as pretty-printed JSON (2-space indent).
3. Write the file using the Write tool.
4. Confirm: `✓ Saved {file-path}`

## Run directory setup (stage 00 only)

When saving the initial input (`00-input.json`), first create the run directory:

1. Use Glob to list existing directories under `runs/` matching `runs/YYYY-MM-DD-*`.
2. Filter to today's date (`YYYY-MM-DD`).
3. Find the highest existing sequence number; increment by 1 (zero-padded to 3 digits). Start at `001` if none exist.
4. The run directory is `runs/YYYY-MM-DD-NNN/`.
5. Write `runs/YYYY-MM-DD-NNN/00-input.json` with the initial input JSON.
6. Return the run directory path — the orchestrator must store this as `RUN_DIR` for all subsequent stages.

## Rules

- Never overwrite an existing artifact without confirming with the user.
- Always pretty-print JSON — never write minified output.
- If the run directory does not exist, create it implicitly via the Write tool path.
- Stage numbers must always be zero-padded to 2 digits (`01`, not `1`).
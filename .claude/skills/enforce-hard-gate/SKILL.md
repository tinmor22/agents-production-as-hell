---
name: enforce-hard-gate
description: Checks a pipeline agent's output against its hard gate requirements. Re-invokes the agent if the gate fails. Use after Sam, Omar, Viktor, and Nate in the orchestrator pipeline.
argument-hint: [agent-name]
allowed-tools: Read
user-invocable: false
---

Enforce the hard gate for **$ARGUMENTS** before the pipeline proceeds.

## Steps

1. **Identify the gate rules** for the agent (see table below).
2. **Check each required field** in the agent's output JSON.
3. **If all rules pass**: confirm `✓ Hard gate passed for <agent>` and continue.
4. **If any rule fails**:
   - Show the user exactly which field failed and why.
   - Do NOT proceed to the next stage.
   - Re-invoke the same agent with the same input.
   - After re-run, check the gate again. Repeat until it passes.

## Hard gate rules

### Sam (3_solver_sam) — after stage 04
| Field | Rule |
|---|---|
| `decision.target_user` | Non-empty string |
| `metrics.north_star` | Non-empty string |

Failure message:
```
✗ Hard gate failed: Sam
- decision.target_user: missing or empty — Sam must define a specific target user with role + context
- metrics.north_star: missing or empty — Sam must define a measurable north star metric
```

### Omar (5_planner_omar) — after stage 06
| Field | Rule |
|---|---|
| `plan.milestones` | Non-empty array (at least 1 milestone) |

Failure message:
```
✗ Hard gate failed: Omar
- plan.milestones: empty or missing — Omar must produce at least one milestone with tasks
```

### Viktor (6_coder_viktor) — after stage 07
| Field | Rule |
|---|---|
| `implementation.test_results.passed` | Must be boolean `true` |

Failure message:
```
✗ Hard gate failed: Viktor
- implementation.test_results.passed: false — tests must pass before proceeding
  Test output: <last 20 lines of implementation.test_results.output>
```

When re-invoking Viktor after a test failure, append the test output to his input:
```json
{ ...same input..., "error_context": { "test_output": "..." } }
```

### Nate (8_deployer_nate) — after stage 09
| Field | Rule |
|---|---|
| `deployment.deploy_command` | Non-empty string |

Failure message:
```
✗ Hard gate failed: Nate
- deployment.deploy_command: missing or empty — Nate must produce a single deploy command
```

## Rules

- Hard gates are non-negotiable. Never skip or soft-fail them.
- Always show the raw failing field value to the user so they can diagnose the issue.
- Maximum 3 re-run attempts per gate. If still failing after 3 attempts, stop the pipeline and ask the user how to proceed.
- Do not modify the agent's input between re-runs unless the agent explicitly failed due to missing context (Viktor test failure is the only exception — append `error_context`).

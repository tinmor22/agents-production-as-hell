package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// DoingAgent spawns `claude --print` with the project dir as working directory.
// Claude Code's built-in tools (Read, Write, Edit, Bash) handle all file I/O.
// Used for: Viktor, Priya, Nate, Rosa.
type DoingAgent struct {
	id           int
	name         string
	systemPrompt string
	model        string
}

func NewDoingAgent(id int, name, systemPrompt, model string) *DoingAgent {
	return &DoingAgent{id: id, name: name, systemPrompt: systemPrompt, model: model}
}

func (a *DoingAgent) StageID() int { return a.id }
func (a *DoingAgent) Name() string { return a.name }

func (a *DoingAgent) Run(ctx context.Context, rc RunContext) (json.RawMessage, error) {
	task := buildDoingTask(a.systemPrompt, rc.Input)

	cmd := exec.CommandContext(ctx, "claude",
		"-p", task,
		"--model", a.model,
		"--output-format", "json",
		"--allowedTools", "Read,Write,Edit,Bash,Glob,Grep",
	)
	cmd.Env = cleanEnv(os.Environ())

	// Doing agents operate inside the target project directory.
	if rc.ProjectDir != "" {
		cmd.Dir = rc.ProjectDir
	}

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("claude subprocess error: %w", err)
	}

	text := extractResultText(out)
	result := extractJSON([]byte(text))
	if result == nil {
		return nil, fmt.Errorf("no valid JSON found in claude output:\n%s", string(out))
	}
	return result, nil
}

func buildDoingTask(systemPrompt string, input json.RawMessage) string {
	return fmt.Sprintf("%s\n\nYour input:\n%s\n\nImplement the work described above in this project directory.\nWhen done, output ONLY valid JSON matching the output schema. No prose, no markdown wrapper.", systemPrompt, string(input))
}

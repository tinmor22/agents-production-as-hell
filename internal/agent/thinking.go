package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ThinkingAgent executes a single `claude --print` call and returns JSON.
// Used for: Nora, Leo, Maya, Sam, Iris, Omar, Ada.
type ThinkingAgent struct {
	id           int
	name         string
	systemPrompt string
	model        string
}

func NewThinkingAgent(id int, name, systemPrompt, model string) *ThinkingAgent {
	return &ThinkingAgent{id: id, name: name, systemPrompt: systemPrompt, model: model}
}

func (a *ThinkingAgent) StageID() int  { return a.id }
func (a *ThinkingAgent) Name() string  { return a.name }

func (a *ThinkingAgent) Run(ctx context.Context, rc RunContext) (json.RawMessage, error) {
	task := buildThinkingTask(a.systemPrompt, rc.Input)
	cmd := exec.CommandContext(ctx, "claude",
		"-p", task,
		"--model", a.model,
		"--output-format", "json",
	)
	cmd.Env = cleanEnv(os.Environ())

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("claude subprocess error: %w", err)
	}

	// --output-format json wraps the response: {"type":"result","subtype":"success","result":"<text>"}
	// Extract the inner text first, then find JSON within it.
	text := extractResultText(out)
	result := extractJSON([]byte(text))
	if result == nil {
		return nil, fmt.Errorf("no valid JSON found in claude output:\n%s", string(out))
	}
	return result, nil
}

func buildThinkingTask(systemPrompt string, input json.RawMessage) string {
	return fmt.Sprintf("%s\n\nInput:\n%s\n\nRespond ONLY with valid JSON matching the output schema. No prose, no markdown wrapper.", systemPrompt, string(input))
}

var jsonBlockRe = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\{.*?\\}|\\[.*?\\])\\s*```")

// extractResultText unwraps the `--output-format json` envelope from claude -p.
// The envelope looks like: {"type":"result","subtype":"success","result":"<text>"}
// Falls back to returning the raw bytes as string if parsing fails.
func extractResultText(out []byte) string {
	var envelope struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(out, &envelope); err == nil && envelope.Result != "" {
		return envelope.Result
	}
	return string(out)
}

// extractJSON finds the first JSON object or array in the output.
// It first tries to find a fenced ```json block, then falls back to raw scan.
func extractJSON(out []byte) json.RawMessage {
	text := strings.TrimSpace(string(out))

	// Try ```json ... ``` block first.
	if m := jsonBlockRe.FindSubmatch([]byte(text)); m != nil {
		candidate := strings.TrimSpace(string(m[1]))
		if json.Valid([]byte(candidate)) {
			return json.RawMessage(candidate)
		}
	}

	// Fall back: find the first { or [ and try to parse from there.
	for i, ch := range text {
		if ch == '{' || ch == '[' {
			candidate := text[i:]
			// Try progressively shorter substrings from the right.
			for j := len(candidate); j > 0; j-- {
				if json.Valid([]byte(candidate[:j])) {
					return json.RawMessage(candidate[:j])
				}
			}
			break
		}
	}
	return nil
}

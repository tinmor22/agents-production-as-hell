package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExtractJSON_ValidObject(t *testing.T) {
	input := []byte(`{"problems": [{"title": "test"}]}`)
	got := extractJSON(input)
	if got == nil {
		t.Fatal("extractJSON returned nil for valid JSON object")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(got, &m); err != nil {
		t.Errorf("extractJSON result is not valid JSON: %v", err)
	}
}

func TestExtractJSON_WithProse(t *testing.T) {
	input := []byte(`Here is the result: {"problems": []}`)
	got := extractJSON(input)
	if got == nil {
		t.Fatal("extractJSON returned nil when JSON is embedded in prose")
	}
}

func TestExtractJSON_FencedBlock(t *testing.T) {
	input := []byte("```json\n{\"problems\": []}\n```")
	got := extractJSON(input)
	if got == nil {
		t.Fatal("extractJSON returned nil for fenced JSON block")
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := []byte("no json here at all")
	got := extractJSON(input)
	if got != nil {
		t.Errorf("extractJSON returned non-nil for non-JSON input: %s", string(got))
	}
}

func TestBuildThinkingTask(t *testing.T) {
	systemPrompt := "You are Nora."
	input := json.RawMessage(`{"topic": "test"}`)
	task := buildThinkingTask(systemPrompt, input)
	if !strings.Contains(task, systemPrompt) {
		t.Error("task does not contain system prompt")
	}
	if !strings.Contains(task, `{"topic": "test"}`) {
		t.Error("task does not contain input JSON")
	}
}

func TestExtractResultText_ValidEnvelope(t *testing.T) {
	envelope := []byte(`{"type":"result","subtype":"success","result":"{\"problems\":[]}"}`)
	got := extractResultText(envelope)
	want := `{"problems":[]}`
	if got != want {
		t.Errorf("extractResultText = %q, want %q", got, want)
	}
}

func TestExtractResultText_Fallback(t *testing.T) {
	raw := []byte("not an envelope")
	got := extractResultText(raw)
	if got != "not an envelope" {
		t.Errorf("extractResultText fallback = %q, want %q", got, "not an envelope")
	}
}

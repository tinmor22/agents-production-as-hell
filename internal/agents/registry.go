// Package agents wires each pipeline stage to its agent type, model, and system prompt.
package agents

import (
	"fmt"

	"github.com/tinmor22/agents-production-as-hell/internal/agent"
	"github.com/tinmor22/agents-production-as-hell/prompts"
)

const (
	ModelSonnet = "claude-sonnet-4-6"
	ModelOpus   = "claude-opus-4-6"
)

func loadPrompt(name string) string {
	data, err := prompts.FS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("missing prompt file %s: %v", name, err))
	}
	return string(data)
}

// All returns the ordered list of pipeline agents: Nora → Ada.
func All() []agent.Agent {
	return []agent.Agent{
		NewNora(),
		NewLeo(),
		NewMaya(),
		NewSam(),
		NewIris(),
		NewOmar(),
		NewViktor(),
		NewPriya(),
		NewNate(),
		NewRosa(),
		NewAda(),
	}
}

package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewSam() agent.Agent {
	return agent.NewThinkingAgent(4, "sam", loadPrompt("solver_sam.md"), ModelSonnet)
}

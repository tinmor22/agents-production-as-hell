package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewNora() agent.Agent {
	return agent.NewThinkingAgent(1, "nora", loadPrompt("problematic_nora.md"), ModelSonnet)
}

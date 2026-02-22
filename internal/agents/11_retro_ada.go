package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewAda() agent.Agent {
	return agent.NewThinkingAgent(11, "ada", loadPrompt("retro_ada.md"), ModelSonnet)
}

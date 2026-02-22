package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewOmar() agent.Agent {
	return agent.NewThinkingAgent(6, "omar", loadPrompt("planner_omar.md"), ModelSonnet)
}

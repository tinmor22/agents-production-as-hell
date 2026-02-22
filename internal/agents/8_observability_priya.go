package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewPriya() agent.Agent {
	return agent.NewDoingAgent(8, "priya", loadPrompt("observability_priya.md"), ModelOpus)
}

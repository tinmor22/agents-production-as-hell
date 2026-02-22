package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewRosa() agent.Agent {
	return agent.NewDoingAgent(10, "rosa", loadPrompt("maintainer_rosa.md"), ModelOpus)
}

package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewIris() agent.Agent {
	return agent.NewThinkingAgent(5, "iris", loadPrompt("designer_iris.md"), ModelSonnet)
}

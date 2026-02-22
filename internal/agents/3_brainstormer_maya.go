package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewMaya() agent.Agent {
	return agent.NewThinkingAgent(3, "maya", loadPrompt("brainstormer_maya.md"), ModelSonnet)
}

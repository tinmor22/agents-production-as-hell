package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewLeo() agent.Agent {
	return agent.NewThinkingAgent(2, "leo", loadPrompt("dreamer_leo.md"), ModelSonnet)
}

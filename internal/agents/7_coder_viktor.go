package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewViktor() agent.Agent {
	return agent.NewDoingAgent(7, "viktor", loadPrompt("coder_viktor.md"), ModelOpus)
}

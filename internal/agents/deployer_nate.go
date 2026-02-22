package agents

import "github.com/tinmor22/agents-production-as-hell/internal/agent"

func NewNate() agent.Agent {
	return agent.NewDoingAgent(9, "nate", loadPrompt("deployer_nate.md"), ModelOpus)
}

package agent

import (
	"context"
	"encoding/json"
)

// Agent is implemented by every pipeline stage.
type Agent interface {
	StageID() int
	Name() string
	Run(ctx context.Context, rc RunContext) (json.RawMessage, error)
}

// RunContext carries everything an agent needs for one execution.
type RunContext struct {
	// Input is the artifact produced by the previous stage (or the initial input JSON).
	Input json.RawMessage

	// RunDir is the path to runs/YYYY-MM-DD-NNN/ where artifacts are persisted.
	RunDir string

	// ProjectDir is the target project directory; Doing agents operate here.
	ProjectDir string
}

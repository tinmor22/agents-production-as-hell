package agents

import "testing"

func TestNewNora(t *testing.T) {
	a := NewNora()
	if got := a.StageID(); got != 1 {
		t.Errorf("StageID() = %d, want 1", got)
	}
	if got := a.Name(); got != "nora" {
		t.Errorf("Name() = %q, want %q", got, "nora")
	}
}

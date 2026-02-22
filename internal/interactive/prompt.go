package interactive

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"
)

// Action represents what the user chose after reviewing a stage output.
type Action int

const (
	ActionApprove Action = iota
	ActionEdit
	ActionReRun
	ActionQuit
)

// GateResult holds the result of hard-gate validation for a stage.
type GateResult struct {
	Passed  bool
	Message string
}

// Prompt displays the post-stage interactive menu and waits for a keypress.
// artifactPath is the file the user can edit on disk.
func Prompt(agentName string, output json.RawMessage, gate GateResult, artifactPath string) Action {
	printDivider()
	printSummary(agentName, output, gate)
	printMenu(agentName)

	for {
		ch := readKey()
		switch strings.ToLower(string(ch)) {
		case "a":
			if !gate.Passed {
				fmt.Printf("\n  Hard gate failed: %s\n  Cannot approve until gate passes. [R]e-run or [E]dit.\n\n", gate.Message)
				printMenu(agentName)
				continue
			}
			fmt.Println("\n  → Approved")
			return ActionApprove
		case "e":
			openEditor(artifactPath)
			fmt.Printf("\n  Waiting for you to save %s...\n", artifactPath)
			fmt.Print("  Press Enter when done: ")
			waitForEnter()
			return ActionEdit
		case "r":
			fmt.Println("\n  → Re-running")
			return ActionReRun
		case "q":
			fmt.Println("\n  → Quit")
			return ActionQuit
		}
	}
}

func printDivider() {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
}

func printSummary(agentName string, output json.RawMessage, gate GateResult) {
	// Count top-level keys or array elements as a rough summary.
	var summary string
	var m map[string]json.RawMessage
	var arr []json.RawMessage
	if err := json.Unmarshal(output, &m); err == nil {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		summary = fmt.Sprintf("keys: %s", strings.Join(keys, ", "))
	} else if err := json.Unmarshal(output, &arr); err == nil {
		summary = fmt.Sprintf("%d items", len(arr))
	} else {
		summary = fmt.Sprintf("%d bytes", len(output))
	}

	gateIcon := "✓"
	if !gate.Passed {
		gateIcon = "✗"
	}

	fmt.Printf("  ✓ %s completed — %s\n", strings.Title(agentName), summary)
	fmt.Printf("  Hard gate: %s %s\n", gateIcon, gate.Message)
	fmt.Println()
}

func printMenu(agentName string) {
	fmt.Printf("  [A] Approve   [E] Edit artifact   [R] Re-run %s   [Q] Quit\n", strings.Title(agentName))
	fmt.Print("  > ")
}

// readKey reads a single keypress without requiring Enter (raw terminal mode).
func readKey() []byte {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback: read a line
		var line string
		fmt.Scanln(&line)
		if len(line) > 0 {
			return []byte{line[0]}
		}
		return []byte("q")
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	return buf
}

func waitForEnter() {
	buf := make([]byte, 1)
	for {
		n, _ := os.Stdin.Read(buf)
		if n > 0 && (buf[0] == '\n' || buf[0] == '\r') {
			break
		}
	}
}

// openEditor opens the artifact file in the user's $EDITOR (or a sensible default).
func openEditor(path string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		switch runtime.GOOS {
		case "darwin":
			editor = "open"
		default:
			editor = "nano"
		}
	}

	// Small delay so the terminal restores before the editor opens.
	time.Sleep(50 * time.Millisecond)

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // ignore error — user may close without saving
}

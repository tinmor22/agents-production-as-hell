package agent

import "strings"

// cleanEnv returns os.Environ() with CLAUDECODE removed so that
// spawned `claude` subprocesses are not rejected as nested sessions.
func cleanEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, "CLAUDECODE=") {
			continue
		}
		out = append(out, e)
	}
	return out
}

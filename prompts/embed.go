// Package prompts embeds all agent system prompt files at compile time.
package prompts

import "embed"

//go:embed *.md
var FS embed.FS

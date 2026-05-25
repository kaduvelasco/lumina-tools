package llm

import (
	"context"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// LLM describes an AI CLI tool managed by lumina.
type LLM struct {
	Name string
	Cmd  string // binary name checked via `which`
}

// Catalogue lists all LLM CLIs managed by lumina.
var Catalogue = []LLM{
	{Name: "Claude Code CLI", Cmd: "claude"},
	{Name: "Antigravity CLI", Cmd: "agy"},
	{Name: "Codex CLI", Cmd: "codex"},
	{Name: "OpenCode CLI", Cmd: "opencode"},
}

// InstalledMap returns which LLMs are currently installed (by Name).
func InstalledMap(ctx context.Context, exe *executor.Executor) map[string]bool {
	result := make(map[string]bool, len(Catalogue))
	for _, l := range Catalogue {
		if _, err := exe.Output(ctx, executor.Options{}, "which", l.Cmd); err == nil {
			result[l.Name] = true
		}
	}
	return result
}

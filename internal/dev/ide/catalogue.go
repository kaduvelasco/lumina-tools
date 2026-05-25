package ide

import (
	"context"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// IDE describes an editor managed by lumina.
type IDE struct {
	Name string
	Cmd  string
}

// Catalogue lists all IDEs managed by lumina.
var Catalogue = []IDE{
	{Name: "Zed Editor", Cmd: "zed"},
	{Name: "Windsurf", Cmd: "windsurf"},
	{Name: "VS Code", Cmd: "code"},
	{Name: "VSCodium", Cmd: "codium"},
}

// InstalledMap returns which IDEs are currently installed (by Name).
func InstalledMap(ctx context.Context, exe *executor.Executor) map[string]bool {
	result := make(map[string]bool, len(Catalogue))
	for _, e := range Catalogue {
		if _, err := exe.Output(ctx, executor.Options{}, "which", e.Cmd); err == nil {
			result[e.Name] = true
		}
	}
	return result
}

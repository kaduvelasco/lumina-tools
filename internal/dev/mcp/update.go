package mcp

import (
	"context"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Update reinstalls all currently installed MCP servers to their latest versions.
func Update(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	servers, err := Catalogue()
	if err != nil || len(servers) == 0 {
		return nil
	}

	installed := InstalledMap(ctx, exe, servers)
	npmPath := os.Getenv("PATH")

	for _, s := range servers {
		if !installed[s.Name] {
			continue
		}
		ui.Info(stdout, "Atualizando "+s.Name+"...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"env", "PATH="+npmPath, "npm", "install", "-g", s.Package,
		); err != nil {
			ui.Warning(stdout, "Falha ao atualizar "+s.Name+": "+err.Error())
		} else {
			ui.Success(stdout, s.Name+" atualizado.")
		}
	}
	return nil
}

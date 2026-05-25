package llm

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Update reinstalls all currently installed LLM CLIs to their latest versions.
func Update(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	installed := InstalledMap(ctx, exe)

	var pending []LLM
	for _, l := range Catalogue {
		if installed[l.Name] {
			pending = append(pending, l)
		}
	}
	if len(pending) == 0 {
		return nil
	}

	if !nodeAvailable(ctx, exe) {
		ui.Info(stdout, "Node.js não encontrado. Instalando via nvm...")
		if err := installNode(ctx, exe, stdout); err != nil {
			ui.Warning(stdout, "Falha ao instalar Node.js: "+err.Error())
			return nil
		}
	}

	for _, l := range pending {
		ui.Info(stdout, "Atualizando "+l.Name+"...")
		if err := installOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha ao atualizar "+l.Name+": "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" atualizado.")
		}
	}
	return nil
}

package terminal

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Update reinstalls all currently installed terminal emulators.
func Update(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	installed := InstalledMap(ctx, exe)
	family := distro.Detect()

	for _, t := range Catalogue {
		if !installed[t.Name] {
			continue
		}
		ui.Info(stdout, "Atualizando "+t.Name+"...")
		if err := installOne(ctx, exe, stdout, t, family); err != nil {
			ui.Warning(stdout, "Falha ao atualizar "+t.Name+": "+err.Error())
		} else {
			ui.Success(stdout, t.Name+" atualizado.")
		}
	}
	return nil
}

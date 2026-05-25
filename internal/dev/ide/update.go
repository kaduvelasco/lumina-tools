package ide

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Update reinstalls all currently installed IDEs to their latest versions.
func Update(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	installed := InstalledMap(ctx, exe)
	family := distro.Detect()

	for _, e := range Catalogue {
		if !installed[e.Name] {
			continue
		}
		ui.Info(stdout, "Atualizando "+e.Name+"...")
		if err := installOne(ctx, exe, stdout, e, family); err != nil {
			ui.Warning(stdout, "Falha ao atualizar "+e.Name+": "+err.Error())
		} else {
			ui.Success(stdout, e.Name+" atualizado.")
		}
	}
	return nil
}

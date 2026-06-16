package gnome

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// ManageCinnamonThemes shows a multi-select for Cinnamon GTK themes and applies the diff.
func ManageCinnamonThemes(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	catalogue, err := loadCinnamonThemeCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de temas Cinnamon: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	return manageThemesFrom(ctx, exe, stdin, stdout, catalogue, "Customizar Cinnamon — Temas GTK")
}

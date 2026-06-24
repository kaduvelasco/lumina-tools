package prereqs

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Select shows a multiselect of all prerequisite groups.
// Items already installed start selected. Deselecting uninstalls; selecting installs.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar pré-requisitos de desenvolvimento")
	ui.Info(stdout, "Verificando pré-requisitos instalados...")

	installed := InstalledMap(ctx, exe)

	items := make([]ui.SelectItem, len(Catalogue))
	for i, p := range Catalogue {
		items[i] = ui.SelectItem{Label: p.Name, ID: p.ID, Selected: installed[p.Name]}
	}

	return ui.RunManagedSelect(ctx, stdin, stdout,
		"Instalar pré-requisitos de desenvolvimento",
		items,
		installed,
		func(i int) error { return installOne(ctx, exe, stdout, Catalogue[i]) },
		func(i int) error { return uninstallOne(ctx, exe, stdout, Catalogue[i]) },
		"Instalação de pré-requisitos concluída.",
	)
}

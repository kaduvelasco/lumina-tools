package terminal

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Select shows a multiselect of all terminal emulators.
// Items already installed start selected. Deselecting one uninstalls it;
// selecting one installs it.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Terminais")
	ui.Info(stdout, "Verificando terminais instalados...")

	installed := InstalledMap(ctx, exe)
	family := distro.Detect()

	items := make([]ui.SelectItem, len(Catalogue))
	for i, t := range Catalogue {
		items[i] = ui.SelectItem{Label: t.Name, ID: t.Cmd, Selected: installed[t.Name]}
	}

	return ui.RunManagedSelect(ctx, stdin, stdout,
		"DevStuff :: Gerenciar Terminais",
		items,
		installed,
		func(i int) error { return installOne(ctx, exe, stdout, Catalogue[i], family) },
		func(i int) error { return uninstallOne(ctx, exe, stdout, Catalogue[i], family) },
		"Gerenciamento de terminais concluído.",
	)
}

package ide

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Select shows a multiselect of all IDEs.
// Items already installed start selected. Deselecting one uninstalls it;
// selecting one installs it.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar IDEs")
	ui.Info(stdout, "Verificando IDEs instalados...")

	installed := InstalledMap(ctx, exe)
	family := distro.Detect()

	items := make([]ui.SelectItem, len(Catalogue))
	for i, e := range Catalogue {
		items[i] = ui.SelectItem{Label: e.Name, ID: e.Cmd, Selected: installed[e.Name]}
	}

	return ui.RunManagedSelect(ctx, stdin, stdout,
		"DevStuff :: Gerenciar IDEs",
		items,
		installed,
		func(i int) error { return installOne(ctx, exe, stdout, Catalogue[i], family) },
		func(i int) error { return uninstallOne(ctx, exe, stdout, Catalogue[i], family) },
		"Gerenciamento de IDEs concluído.",
	)
}

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

	items := make([]ui.SelectItem, len(Catalogue))
	for i, e := range Catalogue {
		items[i] = ui.SelectItem{Label: e.Name, ID: e.Cmd, Selected: installed[e.Name]}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var toInstall, toRemove []IDE
	for i, item := range finalItems {
		e := Catalogue[i]
		switch {
		case item.Selected && !installed[e.Name]:
			toInstall = append(toInstall, e)
		case !item.Selected && installed[e.Name]:
			toRemove = append(toRemove, e)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "DevStuff :: Gerenciar IDEs")
	family := distro.Detect()

	for _, e := range toRemove {
		ui.Info(stdout, "Desinstalando "+e.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, e, family); err != nil {
			ui.Warning(stdout, "Falha ao remover "+e.Name+": "+err.Error())
		} else {
			ui.Success(stdout, e.Name+" removido.")
		}
	}

	for _, e := range toInstall {
		ui.Info(stdout, "Instalando "+e.Name+"...")
		if err := installOne(ctx, exe, stdout, e, family); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+e.Name+": "+err.Error())
		} else {
			ui.Success(stdout, e.Name+" instalado.")
		}
	}

	ui.Success(stdout, "Gerenciamento de IDEs concluído.")
	ui.WaitEnter(stdout)
	return nil
}

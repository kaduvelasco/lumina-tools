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

	items := make([]ui.SelectItem, len(Catalogue))
	for i, t := range Catalogue {
		items[i] = ui.SelectItem{Label: t.Name, ID: t.Cmd, Selected: installed[t.Name]}
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

	var toInstall, toRemove []Terminal
	for i, item := range finalItems {
		t := Catalogue[i]
		switch {
		case item.Selected && !installed[t.Name]:
			toInstall = append(toInstall, t)
		case !item.Selected && installed[t.Name]:
			toRemove = append(toRemove, t)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Terminais")
	family := distro.Detect()

	for _, t := range toRemove {
		ui.Info(stdout, "Desinstalando "+t.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, t, family); err != nil {
			ui.Warning(stdout, "Falha ao remover "+t.Name+": "+err.Error())
		} else {
			ui.Success(stdout, t.Name+" removido.")
		}
	}

	for _, t := range toInstall {
		ui.Info(stdout, "Instalando "+t.Name+"...")
		if err := installOne(ctx, exe, stdout, t, family); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+t.Name+": "+err.Error())
		} else {
			ui.Success(stdout, t.Name+" instalado.")
		}
	}

	ui.Success(stdout, "Gerenciamento de terminais concluído.")
	ui.WaitEnter(stdout)
	return nil
}

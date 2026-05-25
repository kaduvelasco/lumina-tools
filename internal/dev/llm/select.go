package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Select shows a multiselect of all LLM CLIs.
// Items already installed start selected. Deselecting one uninstalls it;
// selecting one installs it. Follows the templates.Select pattern.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar CLIs LLM")
	ui.Info(stdout, "Verificando CLIs instalados...")

	installed := InstalledMap(ctx, exe)

	items := make([]ui.SelectItem, len(Catalogue))
	for i, l := range Catalogue {
		items[i] = ui.SelectItem{Label: l.Name, ID: l.Cmd, Selected: installed[l.Name]}
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

	var toInstall, toRemove []LLM
	for i, item := range finalItems {
		l := Catalogue[i]
		switch {
		case item.Selected && !installed[l.Name]:
			toInstall = append(toInstall, l)
		case !item.Selected && installed[l.Name]:
			toRemove = append(toRemove, l)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "DevStuff :: Gerenciar CLIs LLM")

	if len(toInstall) > 0 && !nodeAvailable(ctx, exe) {
		ui.Info(stdout, "Node.js não encontrado. Instalando via nvm...")
		if err := installNode(ctx, exe, stdout); err != nil {
			ui.Err(stdout, "Falha ao instalar Node.js: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("instalar node: %w", err)
		}
	}

	for _, l := range toRemove {
		ui.Info(stdout, "Desinstalando "+l.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha ao remover "+l.Name+": "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" removido.")
		}
	}

	for _, l := range toInstall {
		ui.Info(stdout, "Instalando "+l.Name+"...")
		if err := installOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+l.Name+": "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" instalado.")
		}
	}

	ui.Success(stdout, "Gerenciamento de CLIs concluído.")
	ui.WaitEnter(stdout)
	return nil
}

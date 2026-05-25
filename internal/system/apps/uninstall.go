package apps

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// SelectUninstall shows an interactive list of installed apps and uninstalls the selected ones.
func SelectUninstall(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Desinstalar Aplicativos")
	ui.Info(stdout, "Verificando aplicativos instalados...")

	installed := InstalledIDs(ctx, exe)
	if len(installed) == 0 {
		ui.Info(stdout, "Nenhum aplicativo Flatpak instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	// Build list from catalogue intersected with installed IDs.
	var items []ui.SelectItem
	for _, a := range Catalogue {
		if installed[a.FlatID] {
			items = append(items, ui.SelectItem{Label: a.Name, ID: a.FlatID})
		}
	}
	// Include installed apps not in the catalogue.
	inCatalogue := make(map[string]bool)
	for _, a := range Catalogue {
		inCatalogue[a.FlatID] = true
	}
	for id := range installed {
		if !inCatalogue[id] {
			items = append(items, ui.SelectItem{Label: id, ID: id})
		}
	}

	if len(items) == 0 {
		ui.Info(stdout, "Nenhum aplicativo disponivel para desinstalar.")
		ui.WaitEnter(stdout)
		return nil
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operacao cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var selected []string
	for _, item := range finalItems {
		if item.Selected {
			selected = append(selected, item.ID)
		}
	}

	if len(selected) == 0 {
		ui.Info(stdout, "Nenhum aplicativo selecionado.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "Desinstalar Aplicativos")
	if err := Uninstall(ctx, exe, stdout, selected); err != nil {
		ui.Err(stdout, "Erro durante a desinstalacao: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Desinstalacao concluida com sucesso!")
	ui.WaitEnter(stdout)
	return nil
}

// Uninstall removes the Flatpak apps identified by flatIDs.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, flatIDs []string) error {
	if len(flatIDs) == 0 {
		fmt.Fprintln(stdout, "Nenhum aplicativo selecionado.")
		return nil
	}

	ui.Info(stdout, fmt.Sprintf("Desinstalando %d aplicativo(s)...", len(flatIDs)))
	var failed []string
	for _, id := range flatIDs {
		ui.Info(stdout, "Desinstalando: "+id)
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"flatpak", "uninstall", config.FlatpakFlag(), "-y", id,
		); err != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao desinstalar %s: %v", id, err))
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d aplicativo(s) nao removido(s): %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}

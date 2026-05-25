package apps

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// InstalledIDs returns a set of Flatpak app IDs currently installed on the system.
func InstalledIDs(ctx context.Context, exe *executor.Executor) map[string]bool {
	out, err := exe.Output(ctx, executor.Options{}, "flatpak", "list", config.FlatpakFlag(), "--app", "--columns=application")
	if err != nil {
		return map[string]bool{}
	}
	result := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		id := strings.TrimSpace(line)
		if id != "" {
			result[id] = true
		}
	}
	return result
}

// EnsureFlatpak checks whether flatpak is available and offers to install it if not.
// Returns an error if flatpak is missing and cannot be installed.
func EnsureFlatpak(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "flatpak"); err == nil {
		return nil
	}
	fmt.Fprintln(stdout, "→ Flatpak não encontrado. Instalando...")
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch distro.Detect() {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "install", "-y", "--", "flatpak"); err != nil {
			return fmt.Errorf("instalar flatpak: %w", err)
		}
	case distro.Fedora:
		if err := exe.Run(ctx, opts, "dnf", "install", "-y", "--", "flatpak"); err != nil {
			return fmt.Errorf("instalar flatpak: %w", err)
		}
	case distro.Arch:
		if err := exe.Run(ctx, opts, "pacman", "-S", "--noconfirm", "--", "flatpak"); err != nil {
			return fmt.Errorf("instalar flatpak: %w", err)
		}
	default:
		return fmt.Errorf("instale o flatpak manualmente nesta distribuição")
	}
	scope := config.FlatpakFlag()
	return exe.Run(ctx,
		executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout},
		"flatpak", "remote-add", scope, "--if-not-exists", "flathub",
		"https://dl.flathub.org/repo/flathub.flatpakrepo",
	)
}

// SelectInstall shows an interactive list of non-installed apps and installs the selected ones.
func SelectInstall(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar Aplicativos")
	ui.Info(stdout, "Verificando aplicativos instalados...")

	installed := InstalledIDs(ctx, exe)

	var items []ui.SelectItem
	for _, a := range Catalogue {
		if !installed[a.FlatID] {
			items = append(items, ui.SelectItem{Label: a.Name, ID: a.FlatID})
		}
	}

	if len(items) == 0 {
		ui.Success(stdout, "Todos os aplicativos do catalogo ja estao instalados.")
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

	ui.PrintHeader(stdout, "Instalar Aplicativos")
	if err := Install(ctx, exe, stdout, selected); err != nil {
		ui.Err(stdout, "Erro durante a instalacao: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Instalacao concluida com sucesso!")
	ui.WaitEnter(stdout)
	return nil
}

// Install installs the Flatpak apps identified by flatIDs from Flathub.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer, flatIDs []string) error {
	if len(flatIDs) == 0 {
		fmt.Fprintln(stdout, "Nenhum aplicativo selecionado.")
		return nil
	}

	if err := EnsureFlatpak(ctx, exe, stdout); err != nil {
		return err
	}

	ui.Info(stdout, fmt.Sprintf("Instalando %d aplicativo(s)...", len(flatIDs)))
	var failed []string
	for _, id := range flatIDs {
		ui.Info(stdout, "Instalando: "+id)
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"flatpak", "install", config.FlatpakFlag(), "-y", "flathub", id,
		); err != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", id, err))
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d aplicativo(s) nao instalado(s): %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}

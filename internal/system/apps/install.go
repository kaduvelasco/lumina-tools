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

// installedByScope returns app IDs installed in a specific Flatpak scope.
func installedByScope(ctx context.Context, exe *executor.Executor, scope string) map[string]bool {
	out, err := exe.Output(ctx, executor.Options{}, "flatpak", "list", scope, "--app", "--columns=application")
	if err != nil {
		return map[string]bool{}
	}
	result := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		if id := strings.TrimSpace(line); id != "" {
			result[id] = true
		}
	}
	return result
}

// listAll returns a map from app ID to its scope flag ("--system" or "--user"),
// querying both scopes in a single pass. System takes precedence over user.
func listAll(ctx context.Context, exe *executor.Executor) map[string]string {
	result := make(map[string]string)
	for id := range installedByScope(ctx, exe, "--user") {
		result[id] = "--user"
	}
	for id := range installedByScope(ctx, exe, "--system") {
		result[id] = "--system"
	}
	return result
}

// InstalledIDs returns all Flatpak app IDs installed in either system or user scope.
func InstalledIDs(ctx context.Context, exe *executor.Executor) map[string]bool {
	all := listAll(ctx, exe)
	out := make(map[string]bool, len(all))
	for id := range all {
		out[id] = true
	}
	return out
}

// InstalledScopeMap returns a map from app ID to its scope flag ("--system" or "--user").
// When an app is installed in both scopes, "--system" takes precedence.
func InstalledScopeMap(ctx context.Context, exe *executor.Executor) map[string]string {
	return listAll(ctx, exe)
}

// EnsureFlatpak checks whether flatpak is available and offers to install it if not.
// Returns an error if flatpak is missing and cannot be installed.
func EnsureFlatpak(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "flatpak"); err == nil {
		return nil
	}
	ui.Info(stdout, "Flatpak não encontrado. Instalando...")
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
		executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}},
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
		ui.Info(stdout, "Nenhum aplicativo selecionado.")
		return nil
	}

	if err := EnsureFlatpak(ctx, exe, stdout); err != nil {
		return err
	}

	appByID := make(map[string]App, len(Catalogue))
	for _, a := range Catalogue {
		appByID[a.FlatID] = a
	}

	scope := config.FlatpakFlag()
	ui.Info(stdout, fmt.Sprintf("Instalando %d aplicativo(s)...", len(flatIDs)))
	var failed []string
	for _, id := range flatIDs {
		ui.Info(stdout, "Instalando: "+id)
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}},
			"flatpak", "install", "--noninteractive", scope, "-y", "flathub", id,
		); err != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", id, err))
			failed = append(failed, id)
		} else if app, ok := appByID[id]; ok && len(app.FlatpakOverride) > 0 {
			overrideArgs := append([]string{"override", scope, id}, app.FlatpakOverride...)
			if err := exe.Run(ctx,
				executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout},
				"flatpak", overrideArgs...,
			); err != nil {
				ui.Warning(stdout, fmt.Sprintf("Override do %s falhou: %v", id, err))
			}
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d aplicativo(s) nao instalado(s): %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}

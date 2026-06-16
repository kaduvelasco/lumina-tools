package gnome

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// ErrNotGnome is returned when the current session is not running GNOME.
var ErrNotGnome = errors.New("a área de trabalho atual não é GNOME — esta opção requer GNOME")

// isGnome reports whether the current desktop session is GNOME.
func isGnome() bool {
	for _, env := range []string{"XDG_CURRENT_DESKTOP", "DESKTOP_SESSION", "GDMSESSION"} {
		v := strings.ToLower(os.Getenv(env))
		if strings.Contains(v, "gnome") || v == "ubuntu" {
			return true
		}
	}
	return false
}

func themesDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".themes"), nil
}

func iconsDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "share", "icons"), nil
}

// globExists reports whether any path matches the given pattern.
func globExists(pattern string) bool {
	matches, _ := filepath.Glob(pattern)
	return len(matches) > 0
}

// shellQuote wraps s in single quotes, escaping any existing single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// manageCatalogue implements the shared select→diff→install/remove flow used by
// cursors and icons. T is the catalogue entry type; callers provide accessor
// functions so the helper stays type-agnostic without requiring an interface.
func manageCatalogue[T any](
	ctx         context.Context,
	exe         *executor.Executor,
	stdin       io.Reader,
	stdout      io.Writer,
	title       string,
	dir         string,
	catalogue   []T,
	nameOf      func(T) string,
	isInstalled func(T, string) bool,
	install     func(context.Context, *executor.Executor, io.Writer, T, string) error,
	remove      func(context.Context, *executor.Executor, io.Writer, T, string) error,
	successMsg  string,
) error {
	ui.Info(stdout, "Verificando itens instalados...")
	items := make([]ui.SelectItem, len(catalogue))
	for i, e := range catalogue {
		items[i] = ui.SelectItem{
			Label:    nameOf(e),
			ID:       nameOf(e),
			Selected: isInstalled(e, dir),
		}
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
	if len(finalItems) != len(catalogue) {
		return fmt.Errorf("inconsistência interna: UI retornou %d itens, catálogo tem %d", len(finalItems), len(catalogue))
	}

	var toInstall, toRemove []T
	for i, item := range finalItems {
		wasInstalled := items[i].Selected
		switch {
		case item.Selected && !wasInstalled:
			toInstall = append(toInstall, catalogue[i])
		case !item.Selected && wasInstalled:
			toRemove = append(toRemove, catalogue[i])
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, title)

	for _, e := range toRemove {
		ui.Info(stdout, "Removendo "+nameOf(e)+"...")
		if rErr := remove(ctx, exe, stdout, e, dir); rErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", nameOf(e), rErr))
		}
	}
	for _, e := range toInstall {
		ui.Info(stdout, "Instalando "+nameOf(e)+"...")
		if iErr := install(ctx, exe, stdout, e, dir); iErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", nameOf(e), iErr))
		}
	}

	ui.Success(stdout, successMsg)
	ui.WaitEnter(stdout)
	return nil
}

// installPackagesByFamily installs packages using the appropriate package manager
// for the running distribution family. packages maps distro family → package list.
// hint is shown when the family is not covered; the function returns nil in that case.
func installPackagesByFamily(
	ctx      context.Context,
	exe      *executor.Executor,
	stdout   io.Writer,
	packages map[string][]string,
	hint     string,
) error {
	family := distro.Detect()
	pkgs, ok := packages[family]
	if !ok {
		ui.Warning(stdout, "Distribuição não identificada para instalação automática de pacotes.")
		ui.Info(stdout, "Instale manualmente: "+hint)
		return nil
	}
	switch family {
	case distro.Debian:
		args := append([]string{
			"install", "-y",
			"-o", "Dpkg::Use-Pty=0",
			"-o", "Dpkg::Progress-Fancy=0",
			"-o", "APT::Color=0",
			"--",
		}, pkgs...)
		return exe.Run(ctx, executor.Options{
			RequiresSudo: true,
			Stdout:       stdout,
			Stderr:       stdout,
			Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
		}, "apt-get", args...)
	case distro.Fedora:
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", append([]string{"install", "-y"}, pkgs...)...,
		)
	case distro.Arch:
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"pacman", append([]string{"-S", "--noconfirm"}, pkgs...)...,
		)
	}
	return nil
}

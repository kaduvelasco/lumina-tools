package terminal

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install shows available terminals and installs selected ones.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Terminais :: Instalar")
	ui.Info(stdout, "Verificando terminais instalados...")

	installed := InstalledMap(ctx, exe)

	for i, t := range Catalogue {
		status := ""
		if installed[t.Name] {
			status = " (instalado)"
		}
		fmt.Fprintf(stdout, "  %d. %s%s\n", i+1, t.Name, status)
	}

	fmt.Fprintln(stdout, "\nDigite os números para instalar, ou Enter para cancelar:")
	fmt.Fprint(stdout, "> ")

	line := prompt.ReadLine()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	selected := prompt.ParseSelection(line, len(Catalogue))
	if len(selected) == 0 {
		ui.Warning(stdout, "Nenhuma seleção válida.")
		ui.WaitEnter(stdout)
		return nil
	}

	family := distro.Detect()

	for _, idx := range selected {
		t := Catalogue[idx]
		ui.Info(stdout, "Instalando "+t.Name+"...")
		if err := installOne(ctx, exe, stdout, t, family); err != nil {
			ui.Warning(stdout, "Falha em "+t.Name+": "+err.Error())
		} else {
			ui.Success(stdout, t.Name+" instalado.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func installOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, t Terminal, family string) error {
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	switch t.Cmd {
	case "kitty":
		script := `curl -fsSL https://sw.kovidgoyal.net/kitty/installer.sh | sh /dev/stdin`
		if err := exe.Run(ctx, opts, "bash", "-c", script); err != nil {
			return err
		}
		linkScript := `
mkdir -p "$HOME/.local/bin"
ln -sf "$HOME/.local/kitty.app/bin/kitty"  "$HOME/.local/bin/kitty"
ln -sf "$HOME/.local/kitty.app/bin/kitten" "$HOME/.local/bin/kitten"
`
		return exe.Run(ctx, opts, "bash", "-c", linkScript)

	case "alacritty":
		return installAlacritty(ctx, exe, family, sudo)

	case "blackbox-terminal":
		return exe.Run(ctx, opts, "flatpak", "install", config.FlatpakFlag(), "-y", "flathub", t.FlatID)
	}
	return fmt.Errorf("instalador desconhecido para %s", t.Name)
}

func installAlacritty(ctx context.Context, exe *executor.Executor, family string, sudo executor.Options) error {
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, sudo, "add-apt-repository", "-y", "universe"); err != nil {
			return err
		}
		if err := exe.Run(ctx, sudo, "apt-get", "update", "-q"); err != nil {
			return err
		}
		return exe.Run(ctx, sudo, "apt-get", "install", "-y", "--", "alacritty")
	case distro.Fedora:
		return exe.Run(ctx, sudo, "dnf", "install", "-y", "--", "alacritty")
	default:
		return exe.Run(ctx, sudo, "pacman", "-S", "--noconfirm", "--", "alacritty")
	}
}


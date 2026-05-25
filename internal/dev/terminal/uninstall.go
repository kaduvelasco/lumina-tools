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

// Uninstall shows installed terminals and removes selected ones.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Terminais :: Desinstalar")
	ui.Info(stdout, "Verificando terminais instalados...")

	installed := InstalledMap(ctx, exe)

	var present []Terminal
	for _, t := range Catalogue {
		if installed[t.Name] {
			present = append(present, t)
		}
	}

	if len(present) == 0 {
		ui.Info(stdout, "Nenhum terminal instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	for i, t := range present {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, t.Name)
	}

	fmt.Fprintln(stdout, "\nDigite os números para desinstalar, ou Enter para cancelar:")
	fmt.Fprint(stdout, "> ")

	line := prompt.ReadLine()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	selected := prompt.ParseSelection(line, len(present))
	if len(selected) == 0 {
		ui.Warning(stdout, "Nenhuma seleção válida.")
		ui.WaitEnter(stdout)
		return nil
	}

	family := distro.Detect()

	for _, idx := range selected {
		t := present[idx]
		ui.Info(stdout, "Desinstalando "+t.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, t, family); err != nil {
			ui.Warning(stdout, "Falha: "+err.Error())
		} else {
			ui.Success(stdout, t.Name+" removido.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func uninstallOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, t Terminal, family string) error {
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	switch t.Cmd {
	case "kitty":
		script := `rm -rf "$HOME/.local/kitty.app" "$HOME/.local/bin/kitty" "$HOME/.local/bin/kitten" 2>/dev/null; true`
		return exe.Run(ctx, opts, "bash", "-c", script)
	case "alacritty":
		switch family {
		case distro.Fedora:
			return exe.Run(ctx, sudo, "dnf", "remove", "-y", "alacritty")
		default:
			return exe.Run(ctx, sudo, "apt-get", "purge", "-y", "--", "alacritty")
		}
	case "blackbox-terminal":
		return exe.Run(ctx, opts, "flatpak", "uninstall", config.FlatpakFlag(), "-y", t.FlatID)
	}
	return fmt.Errorf("desinstalador desconhecido para %s", t.Name)
}

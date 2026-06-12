package ide

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Uninstall shows installed IDEs and removes selected ones.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "IDEs :: Desinstalar")
	ui.Info(stdout, "Verificando IDEs instalados...")

	installed := InstalledMap(ctx, exe)

	var present []IDE
	for _, e := range Catalogue {
		if installed[e.Name] {
			present = append(present, e)
		}
	}

	if len(present) == 0 {
		ui.Info(stdout, "Nenhuma IDE instalada.")
		ui.WaitEnter(stdout)
		return nil
	}

	for i, e := range present {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, e.Name)
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
		e := present[idx]
		ui.Info(stdout, "Desinstalando "+e.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, e, family); err != nil {
			ui.Warning(stdout, "Falha: "+err.Error())
		} else {
			ui.Success(stdout, e.Name+" removido.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func uninstallOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, e IDE, family string) error {
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch e.Cmd {
	case "zed":
		script := `rm -rf "$HOME/.local/zed.app" "$HOME/.local/bin/zed" 2>/dev/null; true`
		return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
	case "windsurf":
		return aptDnfRemove(ctx, exe, "windsurf", family, opts)
	case "code":
		return aptDnfRemove(ctx, exe, "code", family, opts)
	case "codium":
		return aptDnfRemove(ctx, exe, "codium", family, opts)
	}
	return fmt.Errorf("desinstalador desconhecido para %s", e.Name)
}

func aptDnfRemove(ctx context.Context, exe *executor.Executor, pkg, family string, opts executor.Options) error {
	switch family {
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", "remove", "-y", pkg)
	default:
		return exe.Run(ctx, opts, "apt-get", "purge", "-y", "--", pkg)
	}
}

package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/dev/localbin"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install shows available LLMs and installs selected ones.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "LLMs :: Instalar")
	ui.Info(stdout, "Verificando LLMs instalados...")

	installed := InstalledMap(ctx, exe)

	for i, l := range Catalogue {
		status := ""
		if installed[l.Name] {
			status = " (instalado)"
		}
		fmt.Fprintf(stdout, "  %d. %s%s\n", i+1, l.Name, status)
	}

	fmt.Fprintln(stdout, "\nDigite os números para instalar (ex: 1 3), ou Enter para cancelar:")
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

	for _, idx := range selected {
		l := Catalogue[idx]
		if installed[l.Name] {
			ui.Info(stdout, l.Name+" já instalado. Atualizando...")
		} else {
			ui.Info(stdout, "Instalando "+l.Name+"...")
		}
		if err := installOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha em "+l.Name+": "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" instalado.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func installOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, l LLM) error {
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	switch l.Cmd {
	case "claude":
		if err := exe.Run(ctx, opts, "bash", "-c", `curl -fsSL https://claude.ai/install.sh | bash`); err != nil {
			return err
		}
		localbin.EnsureInPath(stdout)
		return nil
	case "agy":
		if err := exe.Run(ctx, opts, "bash", "-c", `curl -fsSL https://antigravity.google/cli/install.sh | bash`); err != nil {
			return err
		}
		localbin.EnsureInPath(stdout)
		return nil
	case "codex":
		return localbin.RunNPMGlobal(ctx, exe, stdout, "install", "@openai/codex")
	case "opencode":
		return localbin.RunNPMGlobal(ctx, exe, stdout, "install", "opencode-ai@latest")
	}
	return fmt.Errorf("instalador desconhecido para %s", l.Name)
}

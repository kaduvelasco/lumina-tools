package llm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Uninstall shows installed LLMs and removes selected ones.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "LLMs :: Desinstalar")
	ui.Info(stdout, "Verificando LLMs instalados...")

	installed := InstalledMap(ctx, exe)

	var present []LLM
	for _, l := range Catalogue {
		if installed[l.Name] {
			present = append(present, l)
		}
	}

	if len(present) == 0 {
		ui.Info(stdout, "Nenhum LLM instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	for i, l := range present {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, l.Name)
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

	for _, idx := range selected {
		l := present[idx]
		ui.Info(stdout, "Desinstalando "+l.Name+"...")
		if err := uninstallOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha: "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" removido.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func uninstallOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, l LLM) error {
	switch l.Cmd {
	case "claude":
		if claudePath, err := exe.Output(ctx, executor.Options{}, "which", "claude"); err == nil {
			if p := strings.TrimSpace(claudePath); p != "" {
				_ = exe.Run(ctx,
					executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
					"rm", "-f", "--", p,
				)
			}
		}
		return nil
	case "agy":
		if agyPath, err := exe.Output(ctx, executor.Options{}, "which", "agy"); err == nil {
			if p := strings.TrimSpace(agyPath); p != "" {
				return exe.Run(ctx,
					executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
					"rm", "-f", "--", p,
				)
			}
		}
		return nil
	case "codex":
		return npmUninstall(ctx, exe, stdout, "@openai/codex")
	case "opencode":
		return npmUninstall(ctx, exe, stdout, "opencode-ai")
	}
	return fmt.Errorf("desinstalador desconhecido para %s", l.Name)
}

func npmUninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkg string) error {
	return exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"env", "PATH="+os.Getenv("PATH"), "npm", "uninstall", "-g", pkg,
	)
}

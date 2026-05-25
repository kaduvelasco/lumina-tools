package repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Init runs git init and applies local identity in the current directory.
func Init(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Repositórios :: Iniciar Repositório")
	ui.Info(stdout, "Diretório: "+cwd())

	if _, err := os.Stat(".git"); err == nil {
		ui.Warning(stdout, "Este diretório já é um repositório Git.")
		fmt.Fprint(stdout, "Reaplicar configurações? (s/N): ")
		if confirm := strings.TrimSpace(prompt.ReadLine()); confirm != "s" && confirm != "S" {
			ui.WaitEnter(stdout)
			return nil
		}
	}

	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	if err := exe.Run(ctx, opts, "git", "init", "-b", "main"); err != nil {
		ui.Err(stdout, "Falha no git init: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("git init: %w", err)
	}

	if err := applyLocalIdentity(ctx, exe, stdout); err != nil {
		if errors.Is(err, errCancelled) {
			ui.Info(stdout, "Repositório inicializado, mas identidade não configurada.")
		} else {
			ui.Err(stdout, err.Error())
		}
		ui.WaitEnter(stdout)
		return nil
	}
	ui.WaitEnter(stdout)
	return nil
}

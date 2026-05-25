package repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Clone clones a repository and applies local identity to it.
func Clone(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Repositórios :: Clonar Repositório")

	fmt.Fprint(stdout, "URL do repositório (Enter para cancelar): ")
	url := strings.TrimSpace(prompt.ReadLine())
	if url == "" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	fmt.Fprint(stdout, "Nome da pasta (Enter para padrão): ")
	dir := strings.TrimSpace(prompt.ReadLine())

	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	args := []string{"clone", url}
	if dir != "" {
		args = append(args, dir)
	}

	ui.Info(stdout, "Clonando repositório...")
	if err := exe.Run(ctx, opts, "git", args...); err != nil {
		ui.Err(stdout, "Falha no git clone: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("git clone: %w", err)
	}

	target := dir
	if target == "" {
		base := filepath.Base(url)
		target = strings.TrimSuffix(base, ".git")
	}

	ui.Info(stdout, "Aplicando identidade em: "+target)

	if err := applyLocalIdentityAt(ctx, exe, stdout, target); err != nil {
		if errors.Is(err, errCancelled) {
			ui.Info(stdout, "Operação cancelada.")
		} else {
			ui.Err(stdout, err.Error())
		}
		ui.WaitEnter(stdout)
		return nil
	}
	ui.WaitEnter(stdout)
	return nil
}

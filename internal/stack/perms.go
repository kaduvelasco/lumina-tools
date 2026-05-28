package stack

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// FixPerms corrects ownership and permissions on the workspace directories.
func FixPerms(ctx context.Context, exe *executor.Executor, stdout io.Writer, workspace string) error {
	ui.PrintHeader(stdout, "Corrigir Permissões")

	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		ui.Err(stdout, "Workspace não encontrado: "+workspace)
		ui.WaitEnter(stdout)
		return fmt.Errorf("workspace nao encontrado: %s", workspace)
	}

	user := executor.CurrentUser()
	if user == "" {
		ui.Err(stdout, "Não foi possível detectar o usuário atual.")
		ui.WaitEnter(stdout)
		return fmt.Errorf("nao foi possivel detectar o usuario atual")
	}

	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	ui.Info(stdout, "Ajustando propriedade de "+workspace+"/www para "+user+"...")
	if err := exe.Run(ctx, opts, "chown", "-R", user+":"+user, workspace+"/www"); err != nil {
		ui.Err(stdout, "Falha no chown: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("chown www: %w", err)
	}

	ui.Info(stdout, "Ajustando permissões www (755)...")
	if err := exe.Run(ctx, opts, "chmod", "-R", "755", workspace+"/www"); err != nil {
		ui.Err(stdout, "Falha no chmod: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("chmod www: %w", err)
	}

	mariadbDir := workspace + "/docker/mariadb"
	if _, err := os.Stat(mariadbDir); err == nil {
		ui.Info(stdout, "Ajustando permissões mariadb (775)...")
		if err := exe.Run(ctx, opts, "chmod", "-R", "775", mariadbDir); err != nil {
			ui.Err(stdout, "Falha no chmod mariadb: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("chmod mariadb: %w", err)
		}
	}

	ui.Success(stdout, "Permissões corrigidas com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

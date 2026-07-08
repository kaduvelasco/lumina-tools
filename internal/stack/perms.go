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

	// databases/ and logs/ hold container-managed data (MariaDB's own files,
	// container log output) — only the top-level directory itself is chowned,
	// non-recursively, so Go code (os.MkdirAll, running unprivileged) can keep
	// creating new subdirectories there (e.g. logs/phpNN for a newly added PHP
	// version) without needing sudo. Docker auto-creates these as root the
	// first time a container bind-mounts a path that doesn't exist yet on the
	// host, which is what leaves them root-owned instead of user-owned.
	for _, d := range []string{workspace + "/databases", workspace + "/logs"} {
		if _, statErr := os.Stat(d); statErr != nil {
			continue
		}
		ui.Info(stdout, "Ajustando propriedade de "+d+" para "+user+"...")
		if err := exe.Run(ctx, opts, "chown", user+":"+user, d); err != nil {
			ui.Err(stdout, "Falha no chown "+d+": "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("chown %s: %w", d, err)
		}
	}

	ui.Success(stdout, "Permissões corrigidas com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

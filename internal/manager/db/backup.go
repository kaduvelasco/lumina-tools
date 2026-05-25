package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Backup dumps all databases from the MariaDB container.
func Backup(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Banco de Dados :: Backup")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}
	container := "mariadb"
	backupDir := filepath.Join(cfg.WorkspacePath, "backups")

	if err := ensureDirExists(backupDir); err != nil {
		ui.Err(stdout, "Falha ao criar diretório de backup: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("criar diretorio de backup: %w", err)
	}

	if err := requireContainer(ctx, exe, container); err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	dbUser, dbPass, err := promptCredentials(stdout)
	if errors.Is(err, errCancelled) {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ts := time.Now().Format("20060102-1504")
	dest := fmt.Sprintf("%s/backup_full_%s.sql", backupDir, ts)

	ui.Info(stdout, "Executando dump para: "+dest)

	// Write password to a temp file readable only by the current user to avoid
	// exposing it in /proc/<pid>/environ of the bash process.
	pwdPath, cleanupPwd, err := writeTempSecret(dbPass, "lumina-db-*.cred")
	if err != nil {
		ui.Err(stdout, "Falha ao criar credencial temporária: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("credencial: %w", err)
	}
	defer cleanupPwd()

	script := fmt.Sprintf(
		`MYSQL_PWD=$(cat %s) docker exec -e MYSQL_PWD %s mariadb-dump -u %s --all-databases > %s`,
		shellQuote(pwdPath), shellQuote(container), shellQuote(dbUser), shellQuote(dest),
	)
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script,
	); err != nil {
		ui.Err(stdout, "Falha no dump: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("dump: %w", err)
	}

	ui.Success(stdout, "Backup concluído: "+dest)
	ui.WaitEnter(stdout)
	return nil
}

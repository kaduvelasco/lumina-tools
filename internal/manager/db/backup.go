package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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
		return failWith(stdout, "Falha ao carregar config", err)
	}
	container := defaultContainer
	backupDir := filepath.Join(cfg.WorkspacePath, "backups")

	if err := ensureDirExists(backupDir); err != nil {
		return failWith(stdout, "Falha ao criar diretório de backup", err)
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

	envPath, cleanupEnv, err := writeTempSecret("MYSQL_PWD="+dbPass+"\n", "lumina-db-*.env")
	if err != nil {
		return failWith(stdout, "Falha ao criar credencial temporária", err)
	}
	defer cleanupEnv()

	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return failWith(stdout, "Falha ao criar arquivo de backup", err)
	}

	if err := exe.Run(ctx,
		executor.Options{Stdout: f, Stderr: stdout},
		"docker", "exec", "--env-file", envPath, container,
		"mariadb-dump", "-u", dbUser, "--all-databases",
	); err != nil {
		f.Close()
		os.Remove(dest)
		return failWith(stdout, "Falha no dump", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(dest)
		return failWith(stdout, "Falha ao fechar arquivo de backup", err)
	}

	ui.Success(stdout, "Backup concluído: "+dest)
	ui.WaitEnter(stdout)
	return nil
}

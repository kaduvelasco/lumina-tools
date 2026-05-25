package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Optimize runs mariadb-check --optimize on all databases.
func Optimize(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Banco de Dados :: Verificar / Otimizar")

	container := "mariadb"
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

	ui.Info(stdout, "Verificando e otimizando todas as tabelas...")

	// Write credentials to a temp env-file (chmod 600) to avoid exposing
	// the password in /proc/<pid>/environ of the docker process.
	envPath, cleanupEnv, err := writeTempSecret("MYSQL_PWD="+dbPass+"\n", "lumina-db-*.env")
	if err != nil {
		ui.Err(stdout, "Falha ao criar credencial temporária: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("credencial: %w", err)
	}
	defer cleanupEnv()

	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"docker", "exec", "-i", "--env-file", envPath, container,
		"mariadb-check", "-u", dbUser, "--all-databases", "--optimize",
	); err != nil {
		ui.Err(stdout, "Falha na otimização: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Otimização concluída com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// OptimizeMoodle writes an optimized MariaDB config for Moodle and restarts the container.
func OptimizeMoodle(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Banco de Dados :: Otimizar para Moodle")

	container := "mariadb"
	if err := requireContainer(ctx, exe, container); err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ramScript := `free -m 2>/dev/null | awk '/^Mem:/{print $2}'`
	ramOut, _ := exe.Output(ctx, executor.Options{}, "bash", "-c", ramScript)
	ramMB := 0
	fmt.Sscan(strings.TrimSpace(ramOut), &ramMB)

	if ramMB > 0 {
		ui.Info(stdout, fmt.Sprintf("RAM detectada: %dMB (~%dGB)", ramMB, ramMB/1024))
	} else {
		fmt.Fprint(stdout, "Informe a RAM em GB (Enter para cancelar): ")
		line := strings.TrimSpace(prompt.ReadLine())
		if line == "" {
			ui.Info(stdout, "Operação cancelada.")
			ui.WaitEnter(stdout)
			return nil
		}
		var gb int
		fmt.Sscan(line, &gb)
		if gb <= 0 {
			ui.Err(stdout, "Valor de RAM inválido. Informe um número inteiro positivo.")
			ui.WaitEnter(stdout)
			return fmt.Errorf("ram invalida")
		}
		ramMB = gb * 1024
	}

	fmt.Fprintln(stdout, "\nProporção do Buffer Pool:")
	fmt.Fprintf(stdout, "  1. 1/2 da RAM (%dMB) — DB dedicado\n", ramMB/2)
	fmt.Fprintf(stdout, "  2. 1/3 da RAM (%dMB) — Equilibrado (recomendado)\n", ramMB/3)
	fmt.Fprintf(stdout, "  3. 1/4 da RAM (%dMB) — Econômico\n", ramMB/4)
	fmt.Fprint(stdout, "\nOpção [1-3]: ")

	var choice int
	fmt.Sscan(prompt.ReadLine(), &choice)

	bufferMB := ramMB / 3
	switch choice {
	case 1:
		bufferMB = ramMB / 2
	case 3:
		bufferMB = ramMB / 4
	}

	ui.Info(stdout, fmt.Sprintf("Configurando innodb_buffer_pool_size para %dMB...", bufferMB))

	cfg, cfgErr := config.Load()
	wsPath := "/srv/workspace"
	if cfgErr != nil {
		ui.Warning(stdout, "Falha ao carregar config, usando caminho padrão: "+cfgErr.Error())
	} else if cfg.WorkspacePath != "" {
		wsPath = cfg.WorkspacePath
	}
	confDir := filepath.Join(wsPath, "docker", "mariadb", "conf.d")
	confFile := filepath.Join(confDir, "moodle-performance.cnf")

	content := fmt.Sprintf(`[mariadb]
max_allowed_packet = 64M
innodb_buffer_pool_size = %dM
innodb_log_file_size = 256M
innodb_file_per_table = 1
innodb_flush_log_at_trx_commit = 2
binlog_format = ROW
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
`, bufferMB)

	mkdirScript := fmt.Sprintf(`mkdir -p %s && printf '%%s' %s > %s`,
		shellQuote(confDir), shellQuote(content), shellQuote(confFile))
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", mkdirScript,
	); err != nil {
		ui.Err(stdout, "Falha ao escrever config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("escrever config: %w", err)
	}

	ui.Info(stdout, "Reiniciando container MariaDB...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"docker", "restart", container,
	); err != nil {
		ui.Err(stdout, "Falha ao reiniciar container: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "MariaDB otimizado para Moodle com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

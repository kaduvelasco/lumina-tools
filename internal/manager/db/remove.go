package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Remove lists user databases and drops selected ones after confirmation.
func Remove(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Banco de Dados :: Remover")
	ui.Warning(stdout, "Atenção: esta operação é irreversível!")

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

	envPath, cleanupEnv, err := writeTempSecret("MYSQL_PWD="+dbPass+"\n", "lumina-db-*.env")
	if err != nil {
		ui.Err(stdout, "Falha ao criar credencial temporária: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("credencial: %w", err)
	}
	defer cleanupEnv()

	out, err := exe.Output(ctx, executor.Options{},
		"docker", "exec", "-i", "--env-file", envPath, container,
		"mariadb", "-u", dbUser, "-e", "SHOW DATABASES;")
	if err != nil {
		ui.Err(stdout, "Falha ao listar bancos: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("listar bancos: %w", err)
	}

	var userDBs []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Database" {
			continue
		}
		switch line {
		case "mysql", "information_schema", "performance_schema", "sys":
			continue
		}
		userDBs = append(userDBs, line)
	}

	if len(userDBs) == 0 {
		ui.Info(stdout, "Nenhum banco de dados personalizado encontrado.")
		ui.WaitEnter(stdout)
		return nil
	}

	for _, db := range userDBs {
		fmt.Fprintf(stdout, "  Remover '%s'? (s/N): ", db)
		confirm := strings.TrimSpace(prompt.ReadLine())
		if confirm != "s" && confirm != "S" {
			continue
		}
		query := "DROP DATABASE `" + strings.ReplaceAll(db, "`", "``") + "`;"
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"docker", "exec", "-i", "--env-file", envPath, container,
			"mariadb", "-u", dbUser, "-e", query,
		); err != nil {
			ui.Warning(stdout, "Falha ao remover '"+db+"': "+err.Error())
		} else {
			ui.Success(stdout, "'"+db+"' removido.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

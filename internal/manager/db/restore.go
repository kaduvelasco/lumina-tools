package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Restore lists available backups and imports the selected one.
func Restore(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Banco de Dados :: Restaurar")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}
	container := "mariadb"
	backupDir := filepath.Join(cfg.WorkspacePath, "backups")

	if err := requireContainer(ctx, exe, container); err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	files, err := listSQLFiles(backupDir)
	if err != nil {
		ui.Err(stdout, "Erro ao listar backups em "+backupDir+": "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("listar backups: %w", err)
	}
	if len(files) == 0 {
		ui.Warning(stdout, "Nenhum backup encontrado em: "+backupDir)
		ui.WaitEnter(stdout)
		return nil
	}

	for i, f := range files {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, filepath.Base(f))
	}

	fmt.Fprint(stdout, "\nSelecione o arquivo [1-N] (Enter para cancelar): ")
	choice := strings.TrimSpace(prompt.ReadLine())
	if choice == "" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}
	var idx int
	if _, err := fmt.Sscan(choice, &idx); err != nil || idx < 1 || idx > len(files) {
		ui.Err(stdout, "Opção inválida.")
		ui.WaitEnter(stdout)
		return fmt.Errorf("opcao invalida")
	}
	file := files[idx-1]

	ui.Info(stdout, "Arquivo: "+filepath.Base(file))
	ui.Warning(stdout, "Use o usuário root ou um superusuário do MariaDB.")

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

	ui.Info(stdout, "Restaurando... Isso pode levar alguns minutos.")

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
		`MYSQL_PWD=$(cat %s) docker exec -i -e MYSQL_PWD %s mariadb -u %s < %s`,
		shellQuote(pwdPath), shellQuote(container), shellQuote(dbUser), shellQuote(file),
	)
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script,
	); err != nil {
		ui.Err(stdout, "Falha no restore: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("restore: %w", err)
	}

	ui.Success(stdout, "Restore concluído com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

func listSQLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, nil
}

package stack

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// These constants mirror the fixed values in the generated docker-compose.yml.
const (
	dbDefaultHost   = "127.0.0.1"
	dbDefaultPort   = "3306"
	dbDefaultName   = "dev_db"
	dbContainerName = "mariadb"
)

// DBInfo prints the MariaDB connection details stored in config.
func DBInfo(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Dados do Banco de Dados")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}

	if cfg.Stack.DBUser == "" {
		ui.Warning(stdout, "Stack não configurada. Execute 'Configurar > Criar Stack' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout,
		"Host     : "+dbDefaultHost+
			"\nPorta    : "+dbDefaultPort+
			"\nBanco    : "+dbDefaultName+
			"\nUsuário  : "+cfg.Stack.DBUser+
			"\nSenha    : "+cfg.Stack.DBPass+
			"\nRoot pw  : "+cfg.Stack.DBRootPass+
			"\nContainer: "+dbContainerName+
			"\n\nArquivo .env: "+cfg.DockerComposeDir+"/.env")
	ui.WaitEnter(stdout)
	return nil
}

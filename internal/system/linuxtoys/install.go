package linuxtoys

import (
	"context"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install runs the official linux.toys installer script.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Linux Toys — Instalação")
	ui.Info(stdout, "Executando instalador do Linux Toys...")
	if err := exe.Run(ctx,
		executor.Options{
			Stdin:  os.Stdin,
			Stdout: stdout,
			Stderr: stdout,
		},
		"bash", "-c", "curl -fsSL https://linux.toys/install.sh | bash",
	); err != nil {
		ui.Err(stdout, "Falha ao instalar Linux Toys: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Linux Toys instalado com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

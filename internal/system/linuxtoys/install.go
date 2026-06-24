package linuxtoys

import (
	"context"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install runs the official linux.toys installer script. Re-running the same
// script is also how Linux Toys updates an existing installation.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Linux Toys — Instalação")

	if _, err := exe.Output(ctx, executor.Options{}, "which", "linuxtoys"); err == nil {
		if !prompt.Confirm(os.Stdin, stdout, "Linux Toys já instalado. Deseja atualizar?", false) {
			ui.Info(stdout, "Operação cancelada.")
			ui.WaitEnter(stdout)
			return nil
		}
	} else if !prompt.Confirm(os.Stdin, stdout, "Deseja continuar com a instalação do Linux Toys?", true) {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

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

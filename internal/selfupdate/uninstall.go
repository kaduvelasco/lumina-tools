package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// ErrUninstalled is returned by Uninstall after the binary is successfully removed.
// Callers use this sentinel to trigger a clean process exit (e.g., the TUI quits).
var ErrUninstalled = errors.New("lumina tools desinstalado")

// Uninstall removes the lumina binary and optionally its config directory.
// Returns ErrUninstalled on success so the caller can exit cleanly.
// Returns nil if the user cancels; returns a wrapped error on failure.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Desinstalar Lumina Tools")
	ui.Warning(stdout, "Esta operação removerá o binário do Lumina Tools do sistema.\nO processo atual continuará em execução até você sair.")

	currentExe, err := os.Executable()
	if err != nil {
		ui.Err(stdout, "Não foi possível localizar o binário: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("localizar binario: %w", err)
	}

	ui.Info(stdout, "Binário: "+currentExe)

	fmt.Fprint(stdout, "\nConfirmar remoção? (s/N): ")
	if c := strings.TrimSpace(prompt.ReadLine()); c != "s" && c != "S" {
		ui.Info(stdout, "Desinstalação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Removendo binário...")
	if err := os.Remove(currentExe); err != nil {
		ui.Info(stdout, "Permissão negada. Tentando com sudo...")
		if sudoErr := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"rm", "-f", "--", currentExe,
		); sudoErr != nil {
			ui.Err(stdout, "Falha ao remover o binário: "+sudoErr.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("remover binario: %w", sudoErr)
		}
	}

	fmt.Fprint(stdout, "\nRemover configurações (~/.lumina)? (s/N): ")
	if c := strings.TrimSpace(prompt.ReadLine()); c == "s" || c == "S" {
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			cfgDir := filepath.Join(home, ".lumina")
			if rmErr := os.RemoveAll(cfgDir); rmErr != nil {
				ui.Warning(stdout, "Falha ao remover configurações: "+rmErr.Error())
			} else {
				ui.Info(stdout, "Configurações removidas.")
			}
		}
	}

	ui.Success(stdout, "Lumina Tools desinstalado com sucesso.")
	fmt.Fprintln(stdout, "\nObrigado por usar o Lumina! Até a próxima.")
	ui.WaitEnter(stdout)
	return ErrUninstalled
}

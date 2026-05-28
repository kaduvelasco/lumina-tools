package stack

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Start brings the stack up in detached mode.
func Start(ctx context.Context, exe *executor.Executor, stdout io.Writer, composeDir string) error {
	ui.PrintHeader(stdout, "Iniciar Stack")
	if composeDir == "" {
		ui.Warning(stdout, "Stack não configurada. Execute 'Configurar > Criar Stack' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}
	cmd, base := composeCmd()
	args := append(append([]string(nil), base...), "-f", filepath.Join(composeDir, "docker-compose.yml"), "up", "-d", "--remove-orphans")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, cmd, args...); err != nil {
		ui.Err(stdout, "Falha ao iniciar stack: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Stack iniciada com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// Stop brings the stack down.
func Stop(ctx context.Context, exe *executor.Executor, stdout io.Writer, composeDir string) error {
	ui.PrintHeader(stdout, "Finalizar Stack")
	if composeDir == "" {
		ui.Warning(stdout, "Stack não configurada. Execute 'Configurar > Criar Stack' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}
	cmd, base := composeCmd()
	args := append(append([]string(nil), base...), "-f", filepath.Join(composeDir, "docker-compose.yml"), "down")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, cmd, args...); err != nil {
		ui.Err(stdout, "Falha ao finalizar stack: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Stack finalizada com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// Logs streams the last 200 lines from all containers.
// Exits cleanly when the user presses Ctrl+C (signal kill).
// Shows an error panel for any other failure (daemon down, file not found, etc.).
func Logs(ctx context.Context, exe *executor.Executor, stdout io.Writer, composeDir string) error {
	ui.PrintHeader(stdout, "Visualizar Logs  —  Ctrl+C para sair")

	if composeDir == "" {
		ui.Warning(stdout, "Stack não configurada. Execute 'Configurar > Criar Stack' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}

	cmd, base := composeCmd()
	composeFile := filepath.Join(composeDir, "docker-compose.yml")
	args := append(append([]string(nil), base...), "-f", composeFile, "logs", "-f", "--tail=200")

	err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, cmd, args...)
	if err == nil || ctx.Err() != nil {
		// clean exit or context cancelled — no panel needed
		return err
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() < 0 {
		// killed by signal (Ctrl+C) — expected, treat as clean exit
		return nil
	}
	ui.Err(stdout, "Falha ao visualizar logs: "+err.Error())
	ui.WaitEnter(stdout)
	return err
}

// Stats shows live resource usage for running containers.
// Exits cleanly when the user presses Ctrl+C.
func Stats(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Status e Recursos  —  Ctrl+C para sair")

	err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout},
		"docker", "stats", "--format",
		"table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}",
	)
	if err == nil || ctx.Err() != nil {
		return err
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() < 0 {
		return nil
	}
	ui.Err(stdout, "Falha ao obter status: "+err.Error())
	ui.WaitEnter(stdout)
	return err
}

var (
	composeOnce    sync.Once
	composeCmdName string
	composeCmdBase []string
)

// composeCmd returns the compose command and its base args, preferring
// the modern 'docker compose' plugin over the legacy 'docker-compose' binary.
// The result is detected once and cached for the lifetime of the process.
func composeCmd() (string, []string) {
	composeOnce.Do(func() {
		if err := exec.Command("docker", "compose", "version").Run(); err == nil {
			composeCmdName = "docker"
			composeCmdBase = []string{"compose"}
		} else {
			composeCmdName = "docker-compose"
		}
	})
	return composeCmdName, append([]string(nil), composeCmdBase...)
}

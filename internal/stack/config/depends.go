package stackconfig

import (
	"context"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Depends installs the prerequisites required by the dev stack.
func Depends(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar Pré-requisitos da Stack")

	family := distro.Detect()
	ui.Info(stdout, "Distribuição detectada: "+family)

	pkgs := []string{"curl", "git", "openssl", "lsof"}
	ui.Info(stdout, "Instalando: "+strings.Join(pkgs, ", "))

	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	var runErr error
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			runErr = err
		} else {
			args := append([]string{"install", "-y", "--"}, pkgs...)
			runErr = exe.Run(ctx, opts, "apt-get", args...)
		}
	case distro.Fedora:
		args := append([]string{"install", "-y", "--"}, pkgs...)
		runErr = exe.Run(ctx, opts, "dnf", args...)
	default:
		args := append([]string{"-S", "--noconfirm", "--"}, pkgs...)
		runErr = exe.Run(ctx, opts, "pacman", args...)
	}

	if runErr != nil {
		ui.Err(stdout, "Falha ao instalar pré-requisitos: "+runErr.Error())
		ui.WaitEnter(stdout)
		return runErr
	}

	ui.Success(stdout, "Pré-requisitos instalados com sucesso!")
	ui.WaitEnter(stdout)
	return nil
}

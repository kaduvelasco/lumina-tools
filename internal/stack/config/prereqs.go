package stackconfig

import (
	"context"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// SetupPrereqs installs base packages (curl, git, openssl, lsof) and Docker Engine
// via the system package manager. Intended as a single first-time setup step.
func SetupPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar Pré-requisitos da Stack")

	family := distro.Detect()
	ui.Info(stdout, "Distribuição detectada: "+family)

	pkgs := []string{"curl", "git", "openssl", "lsof"}
	ui.Info(stdout, "Instalando pacotes base: "+strings.Join(pkgs, ", "))

	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			ui.Err(stdout, "Falha ao atualizar pacotes: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
		args := append([]string{"install", "-y", "--"}, pkgs...)
		if err := exe.Run(ctx, opts, "apt-get", args...); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes base: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	case distro.Fedora:
		args := append([]string{"install", "-y", "--"}, pkgs...)
		if err := exe.Run(ctx, opts, "dnf", args...); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes base: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	default:
		args := append([]string{"-S", "--noconfirm", "--"}, pkgs...)
		if err := exe.Run(ctx, opts, "pacman", args...); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes base: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}
	ui.Success(stdout, "Pacotes base instalados.")

	if _, err := exe.Output(ctx, executor.Options{}, "which", "docker"); err == nil {
		ui.Success(stdout, "Docker já está instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Instalando Docker via gerenciador de pacotes...")
	if err := installDockerPkg(ctx, exe, stdout, family, opts); err != nil {
		ui.Err(stdout, "Falha ao instalar Docker: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	_ = exe.Run(ctx, opts, "systemctl", "enable", "--now", "docker")

	if err := addUserToDockerGroup(ctx, exe, stdout, opts); err != nil {
		ui.Warning(stdout, err.Error())
	}

	if family == distro.Fedora {
		ui.Warning(stdout, "Fedora: se volumes não funcionarem, execute:\n  sudo setsebool -P container_manage_cgroup on")
	}

	ui.Success(stdout, "Pré-requisitos da stack instalados com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

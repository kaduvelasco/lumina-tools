package stackconfig

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Docker installs Docker Engine and configures the current user's group.
func Docker(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar Docker")

	if _, err := exe.Output(ctx, executor.Options{}, "which", "docker"); err == nil {
		ui.Success(stdout, "Docker já está instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	family := distro.Detect()
	ui.Info(stdout, "Distribuição: "+family)

	fmt.Fprintln(stdout, "\nMétodo de instalação:")
	fmt.Fprintln(stdout, "  1. Via gerenciador de pacotes (recomendado)")
	fmt.Fprintln(stdout, "  2. Via script oficial (get.docker.com)")
	fmt.Fprint(stdout, "\nOpção [1]: ")

	method := strings.TrimSpace(prompt.ReadLine())
	if method == "" {
		method = "1"
	}

	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	if method == "1" {
		if err := installDockerPkg(ctx, exe, stdout, family, opts); err != nil {
			ui.Err(stdout, "Falha na instalação: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	} else {
		if err := installDockerScript(ctx, exe, stdout, opts); err != nil {
			ui.Err(stdout, "Falha na instalação: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}

	ui.Info(stdout, "Habilitando e iniciando serviço Docker...")
	_ = exe.Run(ctx, opts, "systemctl", "enable", "--now", "docker")

	if err := addUserToDockerGroup(ctx, exe, stdout, opts); err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	if family == distro.Fedora {
		ui.Warning(stdout, "Fedora: se volumes não funcionarem, execute:\n  sudo setsebool -P container_manage_cgroup on")
	}

	ui.Success(stdout, "Docker instalado com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

func installDockerPkg(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, opts executor.Options) error {
	ui.Info(stdout, "Instalando Docker via gerenciador de pacotes...")
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			return err
		}
		return exe.Run(ctx, opts, "apt-get", "install", "-y", "--", "docker.io", "docker-compose-v2")
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", "install", "-y", "--", "docker", "docker-compose")
	default:
		return exe.Run(ctx, opts, "pacman", "-S", "--noconfirm", "--", "docker", "docker-compose")
	}
}

func installDockerScript(ctx context.Context, exe *executor.Executor, stdout io.Writer, opts executor.Options) error {
	ui.Info(stdout, "Baixando e executando script oficial Docker...")
	script := `
set -e
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -fsSL https://get.docker.com -o "$TMP"
sh "$TMP"
`
	return exe.Run(ctx, opts, "bash", "-c", script)
}

// currentUser returns the real user, preferring SUDO_USER when running under sudo.
func currentUser() string {
	if u := os.Getenv("SUDO_USER"); u != "" {
		return u
	}
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return os.Getenv("LOGNAME")
}

func addUserToDockerGroup(ctx context.Context, exe *executor.Executor, stdout io.Writer, opts executor.Options) error {
	user := currentUser()
	if user == "" {
		return nil
	}

	out, _ := exe.Output(ctx, executor.Options{}, "groups", user)
	if strings.Contains(out, "docker") {
		return nil
	}

	ui.Info(stdout, "Adicionando "+user+" ao grupo docker...")
	if err := exe.Run(ctx, opts, "usermod", "-aG", "docker", user); err != nil {
		return fmt.Errorf("usermod: %w", err)
	}
	ui.Warning(stdout, "Reinicie a sessão para aplicar as permissões do grupo docker.")
	return nil
}

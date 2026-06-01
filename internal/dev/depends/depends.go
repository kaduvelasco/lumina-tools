package depends

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/dev/localbin"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install installs and configures git and libsecret.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Instalar Pré-requisitos")

	family := distro.Detect()
	ui.Info(stdout, "Distribuição detectada: "+family)

	var pkgs []string
	switch family {
	case distro.Debian:
		pkgs = []string{"git", "libsecret-1-0", "libsecret-tools", "gnome-keyring"}
	case distro.Fedora:
		pkgs = []string{"git", "libsecret", "libsecret-devel", "gnome-keyring"}
	default:
		pkgs = []string{"git", "libsecret", "gnome-keyring"}
	}

	ui.Info(stdout, "Pacotes a instalar:\n  "+strings.Join(pkgs, "\n  "))

	fmt.Fprint(stdout, "\nDeseja prosseguir? (s/N): ")
	if c := strings.TrimSpace(prompt.ReadLine()); c != "s" && c != "S" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

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

	freshNVM, nodeErr := installNodeViaNVM(ctx, exe, stdout)
	if nodeErr != nil {
		ui.Warning(stdout, "Falha ao instalar Node.js: "+nodeErr.Error())
	}

	localbin.EnsureInPath(stdout)

	ui.Success(stdout, "Pré-requisitos instalados com sucesso!")
	if freshNVM && nodeErr == nil {
		ui.Warning(stdout, "Reinicie o terminal para ativar o nvm e o Node.js.")
	} else {
		ui.Warning(stdout, "Reinicie o terminal para aplicar todas as alterações de PATH.")
	}
	ui.WaitEnter(stdout)
	return nil
}

func installNodeViaNVM(ctx context.Context, exe *executor.Executor, stdout io.Writer) (freshNVM bool, err error) {
	// Check if node is already available (accounting for nvm in PATH).
	checkScript := `export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"; command -v node`
	if _, e := exe.Output(ctx, executor.Options{}, "bash", "-c", checkScript); e == nil {
		ui.Info(stdout, "Node.js já disponível.")
		return false, nil
	}

	home, e := os.UserHomeDir()
	if e != nil {
		return false, fmt.Errorf("obter diretório home: %w", e)
	}
	_, statErr := os.Stat(filepath.Join(home, ".nvm", "nvm.sh"))
	freshNVM = statErr != nil // nvm not yet installed

	script := `set -e
export NVM_DIR="$HOME/.nvm"
if [ ! -s "$NVM_DIR/nvm.sh" ]; then
    curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
    export NVM_DIR="$HOME/.nvm"
fi
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
nvm install --lts
nvm use --lts
`
	ui.Info(stdout, "Instalando Node.js LTS via nvm...")
	if runErr := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script); runErr != nil {
		return freshNVM, fmt.Errorf("instalar Node.js: %w", runErr)
	}
	return freshNVM, nil
}

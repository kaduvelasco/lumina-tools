package gnome

import (
	"context"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// InstallXFCEPrereqs installs XFCE customization prerequisites for the current distro.
// Installs the murrine GTK engine and sassc (theme compilation, used by the
// install.sh-based entries in the XFCE theme catalogue), git and curl (cloning
// theme repos and downloading the ADW-GTK3 release tarball).
func InstallXFCEPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar XFCE — Pré-requisitos")

	if !prompt.Confirm(os.Stdin, stdout, "Deseja continuar com a instalação dos pré-requisitos de customização?", true) {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Instalando pacotes necessários...")
	if err := installPackagesByFamily(ctx, exe, stdout, map[string][]string{
		distro.Debian: {"gtk2-engines-murrine", "sassc", "git", "curl"},
		distro.Fedora: {"gtk-murrine-engine", "sassc", "git", "curl"},
		distro.Arch:   {"gtk-engine-murrine", "sassc", "git", "curl"},
	}, "gtk2-engines-murrine, sassc, git, curl"); err != nil {
		ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Pré-requisitos instalados!")
	ui.WaitEnter(stdout)
	return nil
}

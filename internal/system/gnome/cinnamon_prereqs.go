package gnome

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// InstallCinnamonPrereqs installs Cinnamon customization prerequisites for the current distro.
// Installs the murrine GTK engine, sassc (for theme compilation) and git.
func InstallCinnamonPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar Cinnamon — Pré-requisitos")
	ui.Info(stdout, "Instalando pacotes necessários...")
	if err := installPackagesByFamily(ctx, exe, stdout, map[string][]string{
		distro.Debian: {"gtk2-engines-murrine", "sassc", "git"},
		distro.Fedora: {"gtk-murrine-engine", "sassc", "git"},
		distro.Arch:   {"gtk-engine-murrine", "sassc", "git"},
	}, "gtk2-engines-murrine, sassc, git"); err != nil {
		ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Pré-requisitos instalados!")
	ui.WaitEnter(stdout)
	return nil
}

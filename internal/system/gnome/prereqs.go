package gnome

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// InstallPrereqs installs GNOME customization prerequisites for the current distro.
func InstallPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar GNOME — Pré-requisitos")

	if !isGnome() {
		ui.Err(stdout, ErrNotGnome.Error())
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Instalando pacotes necessários...")
	if err := installPackagesByFamily(ctx, exe, stdout, map[string][]string{
		distro.Debian: {"gnome-tweaks", "gnome-themes-extra", "gtk2-engines-murrine", "sassc", "git"},
		distro.Fedora: {"gnome-tweaks", "gnome-themes-extra", "gtk-murrine-engine", "sassc", "git"},
		distro.Arch:   {"gnome-tweaks", "gnome-themes-extra", "gtk-engine-murrine", "sassc", "git"},
	}, "gnome-tweaks, gnome-themes-extra, murrine-engine, sassc, git"); err != nil {
		ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	flatpakFlag := config.FlatpakFlag()
	requiresSudo := flatpakFlag == "--system"

	ui.Info(stdout, "Instalando extensões via Flatpak...")
	for _, app := range []string{"org.gnome.Extensions", "com.mattjakeman.ExtensionManager"} {
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: requiresSudo, Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}},
			"flatpak", "install", "--noninteractive", flatpakFlag, "-y", "flathub", app,
		); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+app+": "+err.Error())
		}
	}

	ui.Success(stdout, "Pré-requisitos instalados!")
	ui.WaitEnter(stdout)
	return nil
}

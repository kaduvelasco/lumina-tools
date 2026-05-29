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

	family := distro.Detect()

	ui.Info(stdout, "Instalando pacotes necessários...")
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"apt-get", "install", "-y", "--",
			"gnome-tweaks", "gnome-themes-extra", "gtk2-engines-murrine", "sassc", "git",
			"inkscape", "x11-apps",
		); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	case distro.Fedora:
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", "install", "-y",
			"gnome-tweaks", "gnome-themes-extra", "gtk-murrine-engine", "sassc", "git",
		); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	case distro.Arch:
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"pacman", "-S", "--noconfirm",
			"gnome-tweaks", "gnome-themes-extra", "gtk-engine-murrine", "sassc", "git",
		); err != nil {
			ui.Err(stdout, "Falha ao instalar pacotes: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	default:
		ui.Warning(stdout, "Distribuição não identificada para instalação automática de pacotes.")
		ui.Info(stdout, "Instale manualmente: gnome-tweaks, gnome-themes-extra, murrine-engine, sassc, git")
	}

	flatpakFlag := config.FlatpakFlag()
	requiresSudo := flatpakFlag == "--system"

	ui.Info(stdout, "Instalando extensões via Flatpak...")
	for _, app := range []string{"org.gnome.Extensions", "com.mattjakeman.ExtensionManager"} {
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: requiresSudo, Stdout: stdout, Stderr: stdout},
			"flatpak", "install", flatpakFlag, "-y", "flathub", app,
		); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+app+": "+err.Error())
		}
	}

	ui.Success(stdout, "Pré-requisitos instalados!")
	ui.WaitEnter(stdout)
	return nil
}

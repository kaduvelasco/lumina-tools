package gnome

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// InstallCinnamonPrereqs installs Cinnamon customization prerequisites for the current distro.
// Installs the murrine GTK engine, sassc (for theme compilation) and git.
func InstallCinnamonPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	return installPrereqsCommon(ctx, exe, stdout,
		"Customizar Cinnamon — Pré-requisitos",
		map[string][]string{
			distro.Debian: {"gtk2-engines-murrine", "sassc", "git"},
			distro.Fedora: {"gtk-murrine-engine", "sassc", "git"},
			distro.Arch:   {"gtk-engine-murrine", "sassc", "git"},
		},
		"gtk2-engines-murrine, sassc, git",
	)
}

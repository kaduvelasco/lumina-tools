package postinstall

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// step prints a step panel and runs a sudo command.
func step(ctx context.Context, exe *executor.Executor, stdout io.Writer, msg, name string, args ...string) error {
	ui.Info(stdout, msg)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, name, args...)
}

// aptInstall installs one or more apt packages with sudo.
func aptInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkgs ...string) error {
	args := append([]string{"install", "-y", "-o", "Dpkg::Use-Pty=0", "-o", "Dpkg::Progress-Fancy=0", "-o", "APT::Color=0", "--"}, pkgs...)
	return exe.Run(ctx, executor.Options{
		RequiresSudo: true,
		Stdout:       stdout,
		Stderr:       stdout,
		Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
	}, "apt-get", args...)
}

// dnfInstall installs one or more dnf packages with sudo.
func dnfInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkgs ...string) error {
	args := append([]string{"install", "-y", "--"}, pkgs...)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "dnf", args...)
}

// ensureFlatpakReady checks if flatpak is present and adds the Flathub remote if needed.
func ensureFlatpakReady(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "flatpak"); err != nil {
		ui.Info(stdout, "Instalando Flatpak...")
		if err := aptInstall(ctx, exe, stdout, "flatpak"); err != nil {
			return fmt.Errorf("instalar flatpak: %w", err)
		}
	}
	ui.Info(stdout, "Configurando repositório Flathub...")
	scope := config.FlatpakFlag()
	return exe.Run(ctx,
		executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout},
		"flatpak", "remote-add", scope, "--if-not-exists", "flathub",
		"https://dl.flathub.org/repo/flathub.flatpakrepo",
	)
}

// flatpakInstall installs Flatpak apps from Flathub using the configured scope.
func flatpakInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, appIDs ...string) error {
	args := append([]string{"install", config.FlatpakFlag(), "-y", "flathub"}, appIDs...)
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}}, "flatpak", args...)
}

// configureSysctl sets swappiness, inotify and applies sysctl.
func configureSysctl(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.Info(stdout, "Aplicando configurações de kernel (sysctl)...")
	conf := "vm.swappiness=10\nfs.inotify.max_user_watches=524288\n"
	path := "/etc/sysctl.d/99-lumina.conf"
	cmd := fmt.Sprintf("printf '%%s' %q > %s && sysctl -p %s", conf, path, path)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"bash", "-c", cmd,
	)
}

// acceptMsttFontsEula pre-accepts the EULA for ttf-mscorefonts-installer.
func acceptMsttFontsEula(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	sel := "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true"
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"bash", "-c", fmt.Sprintf("printf '%%s\\n' %q | debconf-set-selections", sel),
	)
}

// failWith shows err as a UI error panel, waits for Enter, and returns the error.
func failWith(stdout io.Writer, err error) error {
	ui.Err(stdout, "Falha: "+err.Error())
	ui.WaitEnter(stdout)
	return err
}

// installVAAPI prompts for hardware video acceleration (Intel / AMD / skip)
// and installs the appropriate packages via the provided installer function.
// Failures are non-fatal: shown as warnings and execution continues.
func installVAAPI(
	ctx context.Context,
	exe *executor.Executor,
	stdout io.Writer,
	intel, amd []string,
	install func(context.Context, *executor.Executor, io.Writer, ...string) error,
) {
	fmt.Fprintf(stdout, "\n  Aceleração de vídeo por hardware:\n  1. Intel\n  2. AMD\n  3. Não instalar\n  Escolha (1/2/3): ")
	choice := strings.TrimSpace(prompt.ReadLine())
	switch choice {
	case "1":
		ui.Info(stdout, "Instalando drivers VA-API para Intel...")
		if err := install(ctx, exe, stdout, intel...); err != nil {
			ui.Warning(stdout, "Falha ao instalar VA-API Intel: "+err.Error())
		}
	case "2":
		ui.Info(stdout, "Instalando drivers VA-API para AMD...")
		if err := install(ctx, exe, stdout, amd...); err != nil {
			ui.Warning(stdout, "Falha ao instalar VA-API AMD: "+err.Error())
		}
	default:
		ui.Info(stdout, "Aceleração de vídeo por hardware ignorada.")
	}
}

func stripNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

package postinstall

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var zorinPackages = []string{
	"libavcodec-extra",
	"ffmpeg",
	"gstreamer1.0-plugins-bad",
	"gstreamer1.0-plugins-ugly",
	"gstreamer1.0-libav",
	"gnome-shell-extension-manager",
	"build-essential",
	"git",
	"curl",
	"wget",
	"htop",
	"fastfetch",
	"gdebi",
	"libfuse2t64",
	"unrar",
	"unzip",
	"ntfs-3g",
	"p7zip-full",
	"tree",
	"jq",
	"plocate",
	"net-tools",
	"software-properties-common",
}

// Zorin runs the post-install routine for ZorinOS 18.1.
func Zorin(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Pós Instalação — ZorinOS 18.1")

	ui.Info(stdout, "Habilitando repositórios universe e multiverse...")
	for _, repo := range []string{"universe", "multiverse"} {
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"add-apt-repository", "-y", repo,
		); err != nil {
			return failWith(stdout, fmt.Errorf("add-apt-repository %s: %w", repo, err))
		}
	}

	if err := step(ctx, exe, stdout, "Atualizando lista de pacotes...", "apt-get", "update", "-y"); err != nil {
		return failWith(stdout, err)
	}
	if err := step(ctx, exe, stdout, "Atualizando pacotes instalados...", "apt-get", "full-upgrade", "-y"); err != nil {
		return failWith(stdout, err)
	}

	ui.Info(stdout, "Instalando pacotes essenciais...")
	if err := aptInstall(ctx, exe, stdout, zorinPackages...); err != nil {
		return failWith(stdout, fmt.Errorf("instalar pacotes: %w", err))
	}

	ui.Info(stdout, "Instalando mídias restritas e fontes Microsoft...")
	if err := acceptMsttFontsEula(ctx, exe, stdout); err != nil {
		return failWith(stdout, err)
	}
	if err := aptInstall(ctx, exe, stdout, "zorin-os-restricted-extras"); err != nil {
		return failWith(stdout, err)
	}

	if err := configureSysctl(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao aplicar sysctl: "+err.Error())
	}

	if err := step(ctx, exe, stdout, "Ativando TRIM para SSDs...",
		"systemctl", "enable", "--now", "fstrim.timer"); err != nil {
		ui.Warning(stdout, "Falha ao ativar TRIM: "+err.Error())
	}

	installVAAPI(ctx, exe, stdout,
		[]string{"intel-media-va-driver"},
		[]string{"mesa-va-drivers"},
		aptInstall,
	)

	ui.Success(stdout, "Pós-instalação do ZorinOS concluída.")
	ui.Warning(stdout, "Reinicie o sistema para aplicar todas as mudanças.")
	ui.WaitEnter(stdout)
	return nil
}

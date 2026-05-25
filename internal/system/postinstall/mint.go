package postinstall

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var mintPackages = []string{
	"mint-meta-codecs",
	"ubuntu-drivers-common",
	"libavcodec-extra",
	"ffmpeg",
	"build-essential",
	"gparted",
	"gdebi",
	"libfuse2t64",
	"unrar",
	"unzip",
	"ntfs-3g",
	"p7zip-full",
	"curl",
	"wget",
	"git",
	"htop",
	"make",
	"tree",
	"jq",
	"plocate",
	"net-tools",
	"python3-pip",
	"fastfetch",
	"software-properties-common",
	"timeshift",
}

// Mint runs the post-install routine for Linux Mint 22.3.
func Mint(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Pós Instalação — Linux Mint 22.3")

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
	if err := aptInstall(ctx, exe, stdout, mintPackages...); err != nil {
		return failWith(stdout, fmt.Errorf("instalar pacotes: %w", err))
	}

	ui.Info(stdout, "Instalando fontes Microsoft...")
	if err := acceptMsttFontsEula(ctx, exe, stdout); err != nil {
		return failWith(stdout, err)
	}
	if err := aptInstall(ctx, exe, stdout, "ttf-mscorefonts-installer"); err != nil {
		return failWith(stdout, err)
	}

	if err := ensureFlatpakReady(ctx, exe, stdout); err != nil {
		return failWith(stdout, err)
	}
	ui.Info(stdout, "Instalando Flatpaks essenciais...")
	if err := flatpakInstall(ctx, exe, stdout, "org.videolan.VLC", "net.codelogistics.webapps"); err != nil {
		ui.Warning(stdout, "Falha ao instalar Flatpaks: "+err.Error())
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

	_ = step(ctx, exe, stdout, "Detectando drivers adicionais...", "ubuntu-drivers", "autoinstall")

	ui.Success(stdout, "Pós-instalação do Linux Mint concluída.")
	ui.Warning(stdout, "Reinicie o sistema para aplicar todas as mudanças.")
	ui.WaitEnter(stdout)
	return nil
}

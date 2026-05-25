package postinstall

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var fedoraPackages = []string{
	"git",
	"curl",
	"wget",
	"htop",
	"fastfetch",
	"make",
	"gcc",
	"gcc-c++",
	"tree",
	"jq",
	"p7zip",
	"p7zip-plugins",
	"unzip",
	"unrar",
	"net-tools",
	"plocate",
	"python3-pip",
}

// Fedora runs the post-install routine for Fedora 44.
func Fedora(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Pós Instalação — Fedora 44")

	if err := step(ctx, exe, stdout, "Atualizando sistema base...", "dnf", "upgrade", "--refresh", "-y"); err != nil {
		return failWith(stdout, err)
	}

	if err := enableRPMFusion(ctx, exe, stdout); err != nil {
		return failWith(stdout, err)
	}

	ui.Info(stdout, "Instalando codecs multimídia...")
	if err := exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"dnf", "group", "install", "-y", "multimedia",
	); err != nil {
		return failWith(stdout, fmt.Errorf("codecs multimedia: %w", err))
	}

	ui.Info(stdout, "Instalando pacotes essenciais...")
	if err := dnfInstall(ctx, exe, stdout, fedoraPackages...); err != nil {
		return failWith(stdout, fmt.Errorf("instalar pacotes: %w", err))
	}

	// ntfs-3g lives in RPM Fusion free but may be unavailable if superseded by the
	// kernel ntfs3 driver in newer Fedora releases — treat as best-effort.
	ui.Info(stdout, "Instalando suporte NTFS (ntfs-3g)...")
	if err := dnfInstall(ctx, exe, stdout, "ntfs-3g"); err != nil {
		ui.Warning(stdout, "ntfs-3g não disponível — o driver ntfs3 do kernel pode já estar ativo.")
	}

	if err := configureSysctl(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao aplicar sysctl: "+err.Error())
	}

	if err := step(ctx, exe, stdout, "Ativando TRIM para SSDs...",
		"systemctl", "enable", "--now", "fstrim.timer"); err != nil {
		ui.Warning(stdout, "Falha ao ativar TRIM: "+err.Error())
	}

	installVAAPI(ctx, exe, stdout,
		[]string{"libva-intel-driver"},
		[]string{"mesa-va-drivers"},
		dnfInstall,
	)

	if err := ensureFlatpakReady(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao configurar Flatpak: "+err.Error())
	}

	ui.Info(stdout, "Limpando pacotes desnecessários...")
	_ = exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "dnf", "autoremove", "-y")
	_ = exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "dnf", "clean", "all")

	ui.Success(stdout, "Pós-instalação do Fedora concluída.")
	ui.Warning(stdout, "Reinicie o sistema para aplicar todas as mudanças.")
	ui.WaitEnter(stdout)
	return nil
}

func enableRPMFusion(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.Info(stdout, "Habilitando repositórios RPM Fusion...")

	ver, err := exe.Output(ctx, executor.Options{}, "rpm", "-E", "%fedora")
	if err != nil {
		return fmt.Errorf("detectar versão fedora: %w", err)
	}
	v := stripNewline(ver)

	freeURL := fmt.Sprintf(
		"https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-%s.noarch.rpm", v,
	)
	nonfreeURL := fmt.Sprintf(
		"https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-%s.noarch.rpm", v,
	)
	return exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"dnf", "install", "-y", freeURL, nonfreeURL,
	)
}

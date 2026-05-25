package update

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Run updates the system: packages, snap and flatpak (when available).
func Run(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Atualizar Sistema")

	family := distro.Detect()
	if family == distro.Unknown {
		ui.Err(stdout, "Distribuição não suportada. Gerenciador de pacotes não reconhecido.")
		ui.WaitEnter(stdout)
		return fmt.Errorf("nenhum gerenciador de pacotes suportado encontrado (apt/dnf/pacman)")
	}
	ui.Info(stdout, "Distribuição detectada: "+family)

	if err := updatePackages(ctx, exe, stdout, family); err != nil {
		ui.Err(stdout, "Falha ao atualizar pacotes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	if err := updateSnap(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao atualizar snaps: "+err.Error())
	}
	if err := updateFlatpak(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao atualizar flatpaks: "+err.Error())
	}
	if err := cleanJournal(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao limpar journal: "+err.Error())
	}

	ui.Success(stdout, "Atualizacao finalizada com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// aptOpts returns executor options for apt-get with interactive output suppressed.
func aptOpts(stdout io.Writer) executor.Options {
	return executor.Options{
		RequiresSudo: true,
		Stdout:       stdout,
		Stderr:       stdout,
		Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
	}
}

func updatePackages(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string) error {
	switch family {
	case distro.Debian:
		ui.Info(stdout, "Atualizando pacotes APT...")
		steps := []struct {
			msg  string
			args []string
		}{
			{"Sincronizando lista de pacotes...", []string{"update", "-o", "APT::Color=0", "-o", "Dpkg::Progress-Fancy=0"}},
			{"Instalando atualizacoes...", []string{"upgrade", "-y", "-o", "APT::Color=0", "-o", "Dpkg::Progress-Fancy=0"}},
			{"Atualizacao completa do sistema...", []string{"full-upgrade", "-y", "-o", "APT::Color=0", "-o", "Dpkg::Progress-Fancy=0"}},
			{"Removendo pacotes desnecessarios...", []string{"autoremove", "-y", "-o", "APT::Color=0"}},
			{"Limpando cache APT...", []string{"autoclean", "-y"}},
		}
		opts := aptOpts(stdout)
		for _, s := range steps {
			ui.Info(stdout, s.msg)
			if err := exe.Run(ctx, opts, "apt-get", s.args...); err != nil {
				return fmt.Errorf("apt: %w", err)
			}
		}
	case distro.Fedora:
		ui.Info(stdout, "Atualizando pacotes DNF...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", "upgrade", "--refresh", "-y",
		); err != nil {
			return fmt.Errorf("dnf upgrade: %w", err)
		}
		if err := exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "dnf", "autoremove", "-y"); err != nil {
			ui.Warning(stdout, "Falha no dnf autoremove: "+err.Error())
		}
	case distro.Arch:
		ui.Info(stdout, "Atualizando pacotes Pacman...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"pacman", "-Syu", "--noconfirm",
		); err != nil {
			return fmt.Errorf("pacman: %w", err)
		}
	}
	return nil
}

func updateSnap(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "snap"); err != nil {
		ui.Info(stdout, "Snap nao instalado. Etapa ignorada.")
		return nil
	}
	ui.Info(stdout, "Atualizando Snaps...")
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "snap", "refresh")
}

func updateFlatpak(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "flatpak"); err != nil {
		ui.Info(stdout, "Flatpak nao instalado. Etapa ignorada.")
		return nil
	}
	ui.Info(stdout, "Atualizando Flatpaks...")
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	if err := exe.Run(ctx, opts, "flatpak", "update", config.FlatpakFlag(), "-y"); err != nil {
		return fmt.Errorf("flatpak update: %w", err)
	}
	if err := exe.Run(ctx, opts, "flatpak", "uninstall", config.FlatpakFlag(), "--unused", "-y"); err != nil {
		ui.Warning(stdout, "Falha ao remover flatpaks nao utilizados: "+err.Error())
	}
	return nil
}

func cleanJournal(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "journalctl"); err != nil {
		return nil
	}
	ui.Info(stdout, "Limpando logs antigos do journal...")
	return exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"journalctl", "--vacuum-time=7d",
	)
}

package ulauncher

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install installs Ulauncher and clones the two official themes.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar Ulauncher")

	family := distro.Detect()

	switch family {
	case distro.Debian, distro.Fedora:
		// supported
	default:
		ui.Warning(stdout, "Distribuição não suportada pelo instalador automático (família: "+family+").\nInstale o Ulauncher manualmente: https://ulauncher.io")
		ui.WaitEnter(stdout)
		return fmt.Errorf("distro nao suportada: %s", family)
	}

	fmt.Fprint(stdout, "  Deseja instalar o Ulauncher? (s/N): ")
	if c := strings.TrimSpace(prompt.ReadLine()); c != "s" && c != "S" {
		ui.Warning(stdout, "Instalação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Instalando Ulauncher...")
	if err := installPackage(ctx, exe, stdout, family); err != nil {
		ui.Err(stdout, "Falha ao instalar Ulauncher: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	ui.Success(stdout, "Ulauncher instalado com sucesso!")

	ui.Info(stdout, "Instalando temas...")
	if err := installThemes(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao instalar temas: "+err.Error())
	} else {
		ui.Success(stdout, "Temas instalados com sucesso!")
	}

	ui.Success(stdout, "Instalação finalizada com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

func installPackage(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string) error {
	if family == distro.Fedora {
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", "install", "-y", "ulauncher",
		)
	}

	steps := []struct {
		msg  string
		name string
		args []string
	}{
		{"Adicionando repositório universe...", "add-apt-repository", []string{"-y", "universe"}},
		{"Adicionando PPA do Ulauncher...", "add-apt-repository", []string{"-y", "ppa:agornostal/ulauncher"}},
		{"Atualizando lista de pacotes...", "apt-get", []string{"update", "-y", "-o", "APT::Color=0"}},
		{"Instalando pacote ulauncher...", "apt-get", []string{"install", "-y", "--", "ulauncher"}},
	}
	for _, s := range steps {
		ui.Info(stdout, s.msg)
		opts := executor.Options{
			RequiresSudo: true,
			Stdout:       stdout,
			Stderr:       stdout,
			Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
		}
		if err := exe.Run(ctx, opts, s.name, s.args...); err != nil {
			return fmt.Errorf("%s: %w", s.msg, err)
		}
	}
	return nil
}

func installThemes(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	themesDir := filepath.Join(home, ".config", "ulauncher", "user-themes")

	ui.Info(stdout, "Criando diretório de temas...")
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		return fmt.Errorf("criar diretorio de temas: %w", err)
	}

	themes := []struct {
		name string
		url  string
	}{
		{"libadwaita-dark", "https://github.com/kareemkasem/ulauncher-theme-libadwaita-dark"},
		{"libadwaita", "https://github.com/leodr/ulauncher-theme-libadwaita.git"},
	}

	for _, t := range themes {
		dest := filepath.Join(themesDir, t.name)
		if _, err := os.Stat(dest); err == nil {
			ui.Info(stdout, fmt.Sprintf("Tema %q já existe. Pulando.", t.name))
			continue
		}
		ui.Info(stdout, fmt.Sprintf("Clonando tema %q...", t.name))
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"git", "clone", t.url, dest,
		); err != nil {
			return fmt.Errorf("clonar tema %s: %w", t.name, err)
		}
	}
	return nil
}

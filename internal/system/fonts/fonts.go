package fonts

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/sets"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Font describes an installable font.
type Font struct {
	Name   string
	Check  string // substring searched in fc-list output
	AptPkg string // empty when installed via download
}

// Catalogue lists all fonts managed by lumina.
var Catalogue = []Font{
	{Name: "JetBrains Mono", Check: "JetBrains Mono", AptPkg: ""},
	{Name: "Carlito", Check: "Carlito", AptPkg: "fonts-crosextra-carlito"},
	{Name: "Caladea", Check: "Caladea", AptPkg: "fonts-crosextra-caladea"},
	{Name: "Noto", Check: "Noto Sans", AptPkg: "fonts-noto"},
	{Name: "Noto CJK", Check: "Noto Sans CJK", AptPkg: "fonts-noto-cjk"},
	{Name: "Noto Color Emoji", Check: "Noto Color Emoji", AptPkg: "fonts-noto-color-emoji"},
}

// InstalledNames returns a set of font names that are currently installed.
func InstalledNames(ctx context.Context, exe *executor.Executor) map[string]bool {
	out, err := exe.Output(ctx, executor.Options{}, "fc-list")
	if err != nil {
		return map[string]bool{}
	}
	result := make(map[string]bool, len(Catalogue))
	for _, line := range strings.Split(out, "\n") {
		for _, f := range Catalogue {
			if !result[f.Name] && strings.Contains(line, f.Check) {
				result[f.Name] = true
			}
		}
	}
	return result
}

// Select shows an interactive multi-select for fonts and applies the diff.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar / Desinstalar Fontes")
	ui.Info(stdout, "Verificando fontes instaladas...")

	installed := InstalledNames(ctx, exe)
	items := make([]ui.SelectItem, len(Catalogue))
	for i, f := range Catalogue {
		items[i] = ui.SelectItem{Label: f.Name, ID: f.Name, Selected: installed[f.Name]}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operacao cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var toInstall, toRemove []string
	for _, item := range finalItems {
		switch {
		case item.Selected && !installed[item.ID]:
			toInstall = append(toInstall, item.ID)
		case !item.Selected && installed[item.ID]:
			toRemove = append(toRemove, item.ID)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteracao necessaria.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "Instalar / Desinstalar Fontes")
	if err := Apply(ctx, exe, stdout, toInstall, toRemove); err != nil {
		ui.Err(stdout, "Erro ao aplicar alteracoes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Fontes atualizadas com sucesso!")
	ui.WaitEnter(stdout)
	return nil
}

// Apply installs fonts listed in toInstall and removes those in toRemove.
func Apply(ctx context.Context, exe *executor.Executor, stdout io.Writer, toInstall, toRemove []string) error {
	installSet := sets.Of(toInstall)
	removeSet := sets.Of(toRemove)
	family := distro.Detect()

	for _, f := range Catalogue {
		switch {
		case installSet[f.Name]:
			if err := install(ctx, exe, stdout, f, family); err != nil {
				ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", f.Name, err))
			}
		case removeSet[f.Name]:
			if err := remove(ctx, exe, stdout, f, family); err != nil {
				ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", f.Name, err))
			}
		}
	}

	ui.Info(stdout, "Atualizando cache de fontes...")
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "fc-cache", "-f")
}

func install(ctx context.Context, exe *executor.Executor, stdout io.Writer, f Font, family string) error {
	ui.Info(stdout, "Instalando "+f.Name+"...")
	if f.AptPkg != "" {
		if family != distro.Debian {
			ui.Warning(stdout, f.Name+" requer apt-get e só está disponível em distribuições Debian/Ubuntu. Instale manualmente.")
			return nil
		}
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"apt-get", "install", "-y", "--", f.AptPkg,
		)
	}
	return installJetBrainsMono(ctx, exe, stdout)
}

func remove(ctx context.Context, exe *executor.Executor, stdout io.Writer, f Font, family string) error {
	ui.Info(stdout, "Removendo "+f.Name+"...")
	if f.AptPkg != "" {
		if family != distro.Debian {
			ui.Warning(stdout, f.Name+" requer apt-get e só está disponível em distribuições Debian/Ubuntu.")
			return nil
		}
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"apt-get", "purge", "-y", "--", f.AptPkg,
		)
	}
	return removeJetBrainsMono(ctx, exe, stdout)
}

// installJetBrainsMono downloads and installs JetBrains Mono from GitHub.
func installJetBrainsMono(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	const version = "2.304"
	url := fmt.Sprintf(
		"https://github.com/JetBrains/JetBrainsMono/releases/download/v%s/JetBrainsMono-%s.zip",
		version, version,
	)
	fontDir := "$HOME/.local/share/fonts"
	script := fmt.Sprintf(`
set -e
mkdir -p "%s"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -fsSL %q -o "$TMP/fonts.zip"
unzip -q "$TMP/fonts.zip" -d "$TMP"
find "$TMP" -maxdepth 3 -name "*.ttf" -exec cp -- {} "%s/" \;
`, fontDir, url, fontDir)

	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script,
	)
}

func removeJetBrainsMono(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	script := `find "$HOME/.local/share/fonts" -name "JetBrainsMono-*.ttf" -delete 2>/dev/null; true`
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
}

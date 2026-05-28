package gnome

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type themeEntry struct {
	Name        string
	DirPattern  string   // glob under ~/.themes/ for detection and removal
	RepoURL     string
	CloneTarget string   // non-empty: clone directly to ~/.themes/<CloneTarget>
	InstallArgs []string // args for ./install.sh (only when CloneTarget is empty)
	AskIcon     bool     // true: prompt the user for the -i <icon> flag (WhiteSur)
	FlatpakName string   // theme name used for GTK_THEME flatpak override
}

var themeCatalogue = []themeEntry{
	{
		Name:        "Orchis",
		DirPattern:  "Orchis*",
		RepoURL:     "https://github.com/vinceliuice/Orchis-theme.git",
		InstallArgs: []string{"-t", "all"},
		FlatpakName: "Orchis",
	},
	{
		Name:        "WhiteSur",
		DirPattern:  "WhiteSur*",
		RepoURL:     "https://github.com/vinceliuice/WhiteSur-gtk-theme.git",
		InstallArgs: []string{"-t", "all"},
		AskIcon:     true,
		FlatpakName: "WhiteSur",
	},
	{
		Name:        "Nordic",
		DirPattern:  "Nordic",
		RepoURL:     "https://github.com/EliverLara/Nordic.git",
		CloneTarget: "Nordic",
		FlatpakName: "Nordic",
	},
	{
		Name:        "Colloid",
		DirPattern:  "Colloid*",
		RepoURL:     "https://github.com/vinceliuice/Colloid-gtk-theme.git",
		InstallArgs: []string{"-t", "all"},
		FlatpakName: "Colloid",
	},
	{
		Name:        "Fluent",
		DirPattern:  "Fluent*",
		RepoURL:     "https://github.com/vinceliuice/Fluent-gtk-theme.git",
		InstallArgs: []string{"-t", "all"},
		FlatpakName: "Fluent",
	},
	{
		Name:        "Tokyonight",
		DirPattern:  "Tokyonight",
		RepoURL:     "https://github.com/Fausto-Korpsvart/Tokyonight-GTK-Theme.git",
		CloneTarget: "Tokyonight",
		FlatpakName: "Tokyonight",
	},
	{
		Name:        "Everforest",
		DirPattern:  "Everforest",
		RepoURL:     "https://github.com/Fausto-Korpsvart/Everforest-GTK-Theme.git",
		CloneTarget: "Everforest",
		FlatpakName: "Everforest",
	},
	{
		Name:        "Rose Pine",
		DirPattern:  "RosePine",
		RepoURL:     "https://github.com/rose-pine/gtk.git",
		CloneTarget: "RosePine",
		FlatpakName: "RosePine",
	},
	{
		Name:        "Gruvbox",
		DirPattern:  "Gruvbox",
		RepoURL:     "https://github.com/Fausto-Korpsvart/Gruvbox-GTK-Theme.git",
		CloneTarget: "Gruvbox",
		FlatpakName: "Gruvbox",
	},
}

// whiteSurIconOptions lists valid values for WhiteSur's -i (titlebar icon) flag.
var whiteSurIconOptions = []ui.SelectItem{
	{Label: "gnome (neutro)", ID: "gnome"},
	{Label: "apple", ID: "apple"},
	{Label: "simple", ID: "simple"},
	{Label: "ubuntu", ID: "ubuntu"},
	{Label: "tux (Linux)", ID: "tux"},
	{Label: "arch", ID: "arch"},
	{Label: "fedora", ID: "fedora"},
	{Label: "debian", ID: "debian"},
	{Label: "zorin", ID: "zorin"},
	{Label: "opensuse", ID: "opensuse"},
	{Label: "popos", ID: "popos"},
	{Label: "mxlinux", ID: "mxlinux"},
	{Label: "budgie", ID: "budgie"},
	{Label: "gentoo", ID: "gentoo"},
}

func isThemeInstalled(t themeEntry, td string) bool {
	return globExists(filepath.Join(td, t.DirPattern))
}

// ManageThemes shows a multi-select for GTK themes and applies the diff.
func ManageThemes(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar GNOME — Temas GTK")

	if !isGnome() {
		ui.Err(stdout, ErrNotGnome.Error())
		ui.WaitEnter(stdout)
		return nil
	}

	td, err := themesDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Info(stdout, "Verificando temas instalados...")
	items := make([]ui.SelectItem, len(themeCatalogue))
	for i, t := range themeCatalogue {
		items[i] = ui.SelectItem{Label: t.Name, ID: t.Name, Selected: isThemeInstalled(t, td)}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var toInstall, toRemove []themeEntry
	for i, item := range finalItems {
		t := themeCatalogue[i]
		wasInstalled := items[i].Selected
		switch {
		case item.Selected && !wasInstalled:
			toInstall = append(toInstall, t)
		case !item.Selected && wasInstalled:
			toRemove = append(toRemove, t)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	// Collect WhiteSur icon choice before starting long operations
	whiteSurIcon := "gnome"
	for _, t := range toInstall {
		if t.AskIcon {
			fmt.Fprintf(stdout, "\nEscolha o ícone da barra de título para WhiteSur:\n")
			idx, ok, ssErr := ui.RunSingleSelect(ctx, stdin, stdout, whiteSurIconOptions)
			if ssErr != nil {
				return ssErr
			}
			if ok && idx >= 0 {
				whiteSurIcon = whiteSurIconOptions[idx].ID
			}
			break
		}
	}

	ui.PrintHeader(stdout, "Customizar GNOME — Temas GTK")

	for _, t := range toRemove {
		ui.Info(stdout, "Removendo "+t.Name+"...")
		if rErr := removeTheme(ctx, exe, stdout, t, td); rErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", t.Name, rErr))
		}
	}

	for _, t := range toInstall {
		ui.Info(stdout, "Instalando "+t.Name+"...")
		icon := ""
		if t.AskIcon {
			icon = whiteSurIcon
		}
		if iErr := installTheme(ctx, exe, stdout, t, td, icon); iErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", t.Name, iErr))
		}
	}

	offerFlatpak(ctx, exe, stdin, stdout, td)

	ui.Success(stdout, "Temas atualizados!")
	ui.WaitEnter(stdout)
	return nil
}

func installTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, t themeEntry, td, icon string) error {
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", "mkdir -p -- \"$1\"", "--", td,
	); err != nil {
		return err
	}

	if t.CloneTarget != "" {
		target := filepath.Join(td, t.CloneTarget)
		script := `
set -e
rm -rf -- "$2"
git clone --depth=1 "$1" "$2"
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", t.RepoURL, target,
		)
	}

	// Build install invocation: ./install.sh <args...> [-i <icon>]
	installCmd := "./install.sh"
	for _, a := range t.InstallArgs {
		installCmd += " " + shellQuote(a)
	}
	if icon != "" {
		installCmd += " -i " + shellQuote(icon)
	}

	script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
git clone --depth=1 "$1" "$TMP/theme"
cd "$TMP/theme"
bash ` + installCmd + `
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", t.RepoURL,
	)
}

func removeTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, t themeEntry, td string) error {
	// $1 = themes dir, $2 = glob pattern; nullglob prevents a no-match from being a literal arg
	script := `
set -e
shopt -s nullglob
for d in "$1"/$2; do
    rm -rf -- "$d"
done
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", td, t.DirPattern,
	)
}

// offerFlatpak prompts the user to apply a GTK theme override to all Flatpak apps.
func offerFlatpak(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer, td string) {
	var installed []ui.SelectItem
	for _, t := range themeCatalogue {
		if isThemeInstalled(t, td) {
			installed = append(installed, ui.SelectItem{Label: t.Name, ID: t.FlatpakName})
		}
	}
	if len(installed) == 0 {
		return
	}
	installed = append(installed, ui.SelectItem{Label: "Não aplicar", ID: ""})

	fmt.Fprintln(stdout)
	ui.Info(stdout, "Aplicar tema GTK ao Flatpak?")
	ui.Info(stdout, "Isso configura todos os apps Flatpak para usar o tema escolhido.")
	fmt.Fprintln(stdout)

	idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, installed)
	if err != nil || !ok || idx < 0 {
		return
	}
	chosen := installed[idx].ID
	if chosen == "" {
		return
	}

	h, err := os.UserHomeDir()
	if err != nil {
		ui.Warning(stdout, "Erro ao obter diretório home: "+err.Error())
		return
	}

	ui.Info(stdout, "Configurando Flatpak para o tema "+chosen+"...")
	_ = exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"flatpak", "override", "--user", "--filesystem="+filepath.Join(h, ".themes"),
	)
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"flatpak", "override", "--user", "--env=GTK_THEME="+chosen,
	); err != nil {
		ui.Warning(stdout, "Falha ao configurar Flatpak: "+err.Error())
		return
	}
	ui.Success(stdout, "Flatpak configurado com o tema "+chosen+".")
}

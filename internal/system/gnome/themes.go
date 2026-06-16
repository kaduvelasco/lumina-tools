package gnome

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type themeEntry struct {
	Name          string   `yaml:"name"`
	DirPattern    string   `yaml:"dir_pattern"`
	RepoURL       string   `yaml:"repo_url"`
	CloneTarget   string   `yaml:"clone_target,omitempty"`
	CopySubDir    string   `yaml:"copy_sub_dir,omitempty"`
	InstallDir    string   `yaml:"install_dir,omitempty"`
	InstallArgs   []string `yaml:"install_args,omitempty"`
	AskIcon       bool     `yaml:"ask_icon,omitempty"`
	FlatpakName   string   `yaml:"flatpak_name,omitempty"`
	CustomScript  string   `yaml:"custom_script,omitempty"`
	PurgePackages []string `yaml:"purge_packages,omitempty"`
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
	if filepath.IsAbs(t.DirPattern) {
		return globExists(t.DirPattern)
	}
	return globExists(filepath.Join(td, t.DirPattern))
}

// ManageThemes shows a multi-select for GNOME GTK themes and applies the diff.
func ManageThemes(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	catalogue, err := loadThemeCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	return manageThemesFrom(ctx, exe, stdin, stdout, catalogue, "Customizar GNOME — Temas GTK")
}

// manageThemesFrom is the shared implementation for GNOME and Cinnamon theme management.
func manageThemesFrom(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer, catalogue []themeEntry, title string) error {
	ui.PrintHeader(stdout, title)

	td, err := themesDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Info(stdout, "Verificando temas instalados...")
	items := make([]ui.SelectItem, len(catalogue))
	for i, t := range catalogue {
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
	if len(finalItems) != len(catalogue) {
		return fmt.Errorf("inconsistência interna: UI retornou %d itens, catálogo tem %d", len(finalItems), len(catalogue))
	}

	var toInstall, toRemove []themeEntry
	for i, item := range finalItems {
		t := catalogue[i]
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
			ui.Info(stdout, "Escolha o ícone da barra de título para WhiteSur:")
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

	ui.PrintHeader(stdout, title)

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

	offerFlatpak(ctx, exe, stdin, stdout, td, catalogue)

	ui.Success(stdout, "Temas atualizados!")
	ui.WaitEnter(stdout)
	return nil
}

func installTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, t themeEntry, td, icon string) error {
	if t.CustomScript != "" {
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"bash", "-c", t.CustomScript,
		)
	}

	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", "mkdir -p -- \"$1\"", "--", td,
	); err != nil {
		return err
	}

	if t.CloneTarget != "" {
		// Clone entire repo as the theme directory (e.g. Nordic, Dracula).
		target := filepath.Join(td, t.CloneTarget)
		script := `
set -e
rm -rf -- "$2"
git clone --depth=1 -- "$1" "$2"
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", t.RepoURL, target,
		)
	}

	if t.CopySubDir != "" {
		// Clone to tempdir and copy each pre-built theme subdir to ~/.themes/ (e.g. Rose Pine).
		// Hidden directories (.git, .github, etc.) are explicitly skipped.
		script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 -- "$1" "$TMP/repo"
for d in "$TMP/repo/$2"/*/; do
    [ -d "$d" ] || continue
    name="$(basename -- "$d")"
    case "$name" in .*) continue ;; esac
    rm -rf -- "$3/$name"
    cp -r -- "$d" "$3/"
done
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", t.RepoURL, t.CopySubDir, td,
		)
	}

	// Run install.sh from InstallDir (empty = repo root; e.g. "themes" for Fausto-Korpsvart repos).
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
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 -- "$1" "$TMP/repo"
cd "$TMP/repo/$2"
bash ` + installCmd + `
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", t.RepoURL, t.InstallDir,
	)
}

func removeTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, t themeEntry, td string) error {
	if len(t.PurgePackages) > 0 {
		args := append([]string{
			"purge", "-y",
			"-o", "Dpkg::Use-Pty=0",
			"-o", "Dpkg::Progress-Fancy=0",
			"-o", "APT::Color=0",
			"--",
		}, t.PurgePackages...)
		return exe.Run(ctx,
			executor.Options{
				RequiresSudo: true,
				Stdout:       stdout,
				Stderr:       stdout,
				Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
			},
			"apt-get", args...,
		)
	}

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

// applyFlatpakTheme configures Flatpak to use the given GTK theme for all apps.
func applyFlatpakTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, chosen, homeDir string) error {
	themesPath := filepath.Join(homeDir, ".themes")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"flatpak", "override", "--user", "--filesystem="+themesPath,
	); err != nil {
		return fmt.Errorf("configurar acesso ao diretório de temas: %w", err)
	}
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"flatpak", "override", "--user", "--env=GTK_THEME="+chosen,
	); err != nil {
		return fmt.Errorf("configurar GTK_THEME: %w", err)
	}
	return nil
}

// selectFlatpakTheme scans catalogue for installed themes and runs a single-select.
// Returns the chosen FlatpakName and ok=true when the user confirms a choice.
// Returns ("", false, nil) when no themes are installed or the user cancels (ESC).
// Returns ("", true, nil) when the user explicitly picks "Não aplicar".
func selectFlatpakTheme(ctx context.Context, stdin io.Reader, stdout io.Writer, td string, catalogue []themeEntry) (string, bool, error) {
	var items []ui.SelectItem
	for _, t := range catalogue {
		if isThemeInstalled(t, td) {
			items = append(items, ui.SelectItem{Label: t.Name, ID: t.FlatpakName})
		}
	}
	if len(items) == 0 {
		return "", false, nil
	}
	items = append(items, ui.SelectItem{Label: "Não aplicar", ID: ""})

	idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, items)
	if err != nil {
		return "", false, err
	}
	if !ok || idx < 0 {
		return "", false, nil
	}
	return items[idx].ID, true, nil
}

// offerFlatpak prompts the user to apply a GTK theme override to all Flatpak apps.
func offerFlatpak(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer, td string, catalogue []themeEntry) {
	// Pre-check: avoid printing the prompt when no themes are installed.
	anyInstalled := false
	for _, t := range catalogue {
		if isThemeInstalled(t, td) {
			anyInstalled = true
			break
		}
	}
	if !anyInstalled {
		return
	}

	fmt.Fprintln(stdout)
	ui.Info(stdout, "Aplicar tema GTK ao Flatpak?")
	ui.Info(stdout, "Isso configura todos os apps Flatpak para usar o tema escolhido.")
	fmt.Fprintln(stdout)

	chosen, _, err := selectFlatpakTheme(ctx, stdin, stdout, td, catalogue)
	if err != nil || chosen == "" {
		return
	}

	h, err := os.UserHomeDir()
	if err != nil {
		ui.Warning(stdout, "Erro ao obter diretório home: "+err.Error())
		return
	}

	ui.Info(stdout, "Configurando Flatpak para o tema "+chosen+"...")
	if err := applyFlatpakTheme(ctx, exe, stdout, chosen, h); err != nil {
		ui.Warning(stdout, "Falha ao configurar Flatpak: "+err.Error())
		return
	}
	ui.Success(stdout, "Flatpak configurado com o tema "+chosen+".")
}

// ApplyFlatpakTheme lets the user pick an installed GTK theme and apply it as a
// GTK_THEME override for all Flatpak apps. Works for both GNOME and Cinnamon —
// shows the appropriate catalogue based on the DE saved in config.
func ApplyFlatpakTheme(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar — Aplicar Tema no Flatpak")

	td, err := themesDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	catalogue, err := loadThemeCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		ui.Warning(stdout, "Não foi possível carregar configuração: "+cfgErr.Error())
	} else if cfg.DE == "cinnamon" {
		cin, cinErr := loadCinnamonThemeCatalogue()
		if cinErr != nil {
			ui.Warning(stdout, "Erro ao carregar catálogo Cinnamon: "+cinErr.Error())
		} else {
			catalogue = cin
		}
	}

	// Pre-check: show an explicit warning when no themes are installed yet.
	hasInstalled := false
	for _, t := range catalogue {
		if isThemeInstalled(t, td) {
			hasInstalled = true
			break
		}
	}
	if !hasInstalled {
		ui.Warning(stdout, "Nenhum tema compatível encontrado em ~/.themes/")
		ui.Info(stdout, "Instale pelo menos um tema GTK antes de usar esta opção.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Selecione o tema GTK para aplicar em todos os apps Flatpak:")
	fmt.Fprintln(stdout)

	chosen, ok, err := selectFlatpakTheme(ctx, stdin, stdout, td, catalogue)
	if err != nil {
		return err
	}
	// hasInstalled=true means selectFlatpakTheme had items; !ok here means user cancelled.
	if !ok {
		ui.Warning(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}
	if chosen == "" {
		ui.Info(stdout, "Nenhuma alteração aplicada.")
		ui.WaitEnter(stdout)
		return nil
	}

	h, err := os.UserHomeDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório home: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Info(stdout, "Configurando Flatpak para o tema "+chosen+"...")
	if err := applyFlatpakTheme(ctx, exe, stdout, chosen, h); err != nil {
		ui.Warning(stdout, "Falha ao configurar Flatpak: "+err.Error())
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Success(stdout, "Flatpak configurado com o tema "+chosen+".")
	ui.WaitEnter(stdout)
	return nil
}

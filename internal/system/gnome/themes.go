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
	Name          string     `yaml:"name"`
	DirPattern    string     `yaml:"dir_pattern"`
	RepoURL       string     `yaml:"repo_url,omitempty"`
	CloneTarget   string     `yaml:"clone_target,omitempty"`
	CopySubDir    string     `yaml:"copy_sub_dir,omitempty"`
	InstallDir    string     `yaml:"install_dir,omitempty"`
	InstallArgs   []string   `yaml:"install_args,omitempty"`
	TweakVariants [][]string `yaml:"tweak_variants,omitempty"`
	FixedTweaks   []string   `yaml:"fixed_tweaks,omitempty"`
	AskBorder     bool       `yaml:"ask_border,omitempty"`
	BorderTweak   string     `yaml:"border_tweak,omitempty"`
	AskButtons    bool       `yaml:"ask_buttons,omitempty"`
	ButtonsTweak  string     `yaml:"buttons_tweak,omitempty"`
	FlatpakName   string     `yaml:"flatpak_name,omitempty"`
	CustomScript  string     `yaml:"custom_script,omitempty"`
	UserScript    string     `yaml:"user_script,omitempty"`
	PurgePackages []string   `yaml:"purge_packages,omitempty"`

	// ExtraDirPatterns lists additional glob patterns (besides DirPattern) to
	// remove on uninstall — for entries that install several unrelated-looking
	// folders at once (e.g. a multi-theme collection) that a single glob can't cover.
	ExtraDirPatterns []string `yaml:"extra_dir_patterns,omitempty"`
}

// borderOptions returns the SelectItem list for a given border tweak value.
func borderOptions(tweakVal string) []ui.SelectItem {
	switch tweakVal {
	case "rimless":
		return []ui.SelectItem{
			{Label: "Com borda (padrão)", ID: ""},
			{Label: "Rimless (sem borda)", ID: "rimless"},
		}
	case "outline":
		return []ui.SelectItem{
			{Label: "Sem borda (padrão)", ID: ""},
			{Label: "Com borda (2px outline)", ID: "outline"},
		}
	default:
		return []ui.SelectItem{
			{Label: "Sem borda (padrão)", ID: ""},
			{Label: "Com borda", ID: tweakVal},
		}
	}
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

// ManageXFCEThemes shows a multi-select for XFCE GTK/XFWM4 themes and applies the diff.
func ManageXFCEThemes(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	catalogue, err := loadXFCEThemeCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de temas: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	return manageThemesFrom(ctx, exe, stdin, stdout, catalogue, "Customizar XFCE — Temas")
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

	// Collect border choices (once per distinct border_tweak value).
	borderChoices := make(map[string]string)
	for _, t := range toInstall {
		if !t.AskBorder || t.BorderTweak == "" {
			continue
		}
		if _, asked := borderChoices[t.BorderTweak]; asked {
			continue
		}
		ui.Info(stdout, "Estilo de janela ("+t.BorderTweak+"):")
		opts := borderOptions(t.BorderTweak)
		idx, ok, ssErr := ui.RunSingleSelect(ctx, stdin, stdout, opts)
		if ssErr != nil {
			return ssErr
		}
		chosen := ""
		if ok && idx >= 0 {
			chosen = opts[idx].ID
		}
		borderChoices[t.BorderTweak] = chosen
	}

	// Collect button style choice (once, shared across all themes that ask).
	buttonsChoice := ""
	for _, t := range toInstall {
		if !t.AskButtons || t.ButtonsTweak == "" {
			continue
		}
		ui.Info(stdout, "Estilo dos botões de janela:")
		opts := []ui.SelectItem{
			{Label: "Legacy (padrão)", ID: ""},
			{Label: "macOS", ID: t.ButtonsTweak},
		}
		idx, ok, ssErr := ui.RunSingleSelect(ctx, stdin, stdout, opts)
		if ssErr != nil {
			return ssErr
		}
		if ok && idx >= 0 {
			buttonsChoice = opts[idx].ID
		}
		break
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
		border := borderChoices[t.BorderTweak]
		if iErr := installTheme(ctx, exe, stdout, t, td, border, buttonsChoice); iErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", t.Name, iErr))
		}
	}

	offerFlatpak(ctx, exe, stdin, stdout, td, catalogue)

	ui.Success(stdout, "Temas atualizados!")
	ui.WaitEnter(stdout)
	return nil
}

func installTheme(ctx context.Context, exe *executor.Executor, stdout io.Writer, t themeEntry, td, border, buttons string) error {
	if t.CustomScript != "" {
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"bash", "-c", t.CustomScript,
		)
	}

	// UserScript runs as the current user (unlike CustomScript) — for themes
	// that write to the user's own ~/.themes via a multi-step process (e.g.
	// downloading a release tarball and merging in a companion repo) that a
	// plain clone/install.sh flow can't express.
	if t.UserScript != "" {
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", "mkdir -p -- \"$1\"", "--", td,
		); err != nil {
			return err
		}
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", t.UserScript, "--", td,
		)
	}

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
git clone --depth=1 -- "$1" "$2"
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", t.RepoURL, target,
		)
	}

	if t.CopySubDir != "" {
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

	// Run install.sh once per TweakVariant (or once with no extra tweaks if empty).
	runs := t.TweakVariants
	if len(runs) == 0 {
		runs = [][]string{{}}
	}

	for _, varTweaks := range runs {
		// Combine: fixed + variant + user border + user buttons
		var tweaks []string
		tweaks = append(tweaks, t.FixedTweaks...)
		tweaks = append(tweaks, varTweaks...)
		if border != "" {
			tweaks = append(tweaks, border)
		}
		if buttons != "" {
			tweaks = append(tweaks, buttons)
		}

		installCmd := "./install.sh"
		for _, a := range t.InstallArgs {
			installCmd += " " + shellQuote(a)
		}
		if len(tweaks) > 0 {
			installCmd += " --tweaks"
			for _, tw := range tweaks {
				installCmd += " " + shellQuote(tw)
			}
		}

		script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 -- "$1" "$TMP/repo"
cd "$TMP/repo/$2"
bash ` + installCmd + `
`
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", t.RepoURL, t.InstallDir,
		); err != nil {
			return err
		}
	}
	return nil
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

	// $1 = themes dir, $2 = glob pattern; nullglob prevents a no-match from being a literal arg.
	// Run once per pattern so each glob stays a single, properly quoted argument.
	script := `
set -e
shopt -s nullglob
for d in "$1"/$2; do
    rm -rf -- "$d"
done
`
	for _, pattern := range append([]string{t.DirPattern}, t.ExtraDirPatterns...) {
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", td, pattern,
		); err != nil {
			return err
		}
	}
	return nil
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

// selectFlatpakTheme lists all directories in ~/.themes as flatpak theme options.
// Returns the chosen theme name and ok=true when the user confirms a choice.
// Returns ("", false, nil) when ~/.themes is empty or the user cancels (ESC).
// Returns ("", true, nil) when the user explicitly picks "Não aplicar".
func selectFlatpakTheme(ctx context.Context, stdin io.Reader, stdout io.Writer, td string, _ []themeEntry) (string, bool, error) {
	entries, err := os.ReadDir(td)
	if err != nil || len(entries) == 0 {
		return "", false, nil
	}
	var items []ui.SelectItem
	for _, e := range entries {
		if e.IsDir() {
			items = append(items, ui.SelectItem{Label: e.Name(), ID: e.Name()})
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
	entries, _ := os.ReadDir(td)
	hasDirs := false
	for _, e := range entries {
		if e.IsDir() {
			hasDirs = true
			break
		}
	}
	if !hasDirs {
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

	entries, _ := os.ReadDir(td)
	hasDirs := false
	for _, e := range entries {
		if e.IsDir() {
			hasDirs = true
			break
		}
	}
	if !hasDirs {
		ui.Warning(stdout, "Nenhum tema encontrado em ~/.themes/")
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

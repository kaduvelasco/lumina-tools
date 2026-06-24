package gnome

import (
	"context"
	"io"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type iconEntry struct {
	Name       string `yaml:"name"`
	DirPattern string `yaml:"dir_pattern"`
	RepoURL    string `yaml:"repo_url"`
	CloneAs    string `yaml:"clone_as,omitempty"`
	CopyGlob   string `yaml:"copy_glob,omitempty"`
	// UserScript runs after cloning the repo to a temp dir (cwd = the clone),
	// for icon packs whose theme folders must be built by their own install.sh
	// rather than copied straight from the repo (e.g. Tela). $1 = repo URL,
	// $2 = icons dir — same positional params as the rest of this script.
	UserScript string `yaml:"user_script,omitempty"`
	// ExtraDirPatterns lists additional exact names/globs (besides DirPattern)
	// removed on uninstall, for packs that install several folders at once.
	ExtraDirPatterns []string `yaml:"extra_dir_patterns,omitempty"`
}

func isIconInstalled(ic iconEntry, id string) bool {
	return globExists(filepath.Join(id, ic.DirPattern))
}

// ManageIcons shows a multi-select for icon packs and applies the diff.
func ManageIcons(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar — Ícones")

	id, err := iconsDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de ícones: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	catalogue, err := loadIconCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de ícones: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	return manageCatalogue(ctx, exe, stdin, stdout,
		"Customizar — Ícones", id, catalogue,
		func(ic iconEntry) string { return ic.Name },
		isIconInstalled,
		installIcon,
		removeIcon,
		"Ícones atualizados!",
	)
}

func installIcon(ctx context.Context, exe *executor.Executor, stdout io.Writer, ic iconEntry, id string) error {
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", "mkdir -p -- \"$1\"", "--", id,
	); err != nil {
		return err
	}

	if ic.CloneAs != "" {
		target := filepath.Join(id, ic.CloneAs)
		script := `
set -e
rm -rf -- "$2"
git clone --depth=1 "$1" "$2"
gtk-update-icon-cache -f -t "$2" 2>/dev/null || true
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", ic.RepoURL, target,
		)
	}

	if ic.UserScript != "" {
		script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 "$1" "$TMP/pack"
cd "$TMP/pack"
` + ic.UserScript + `
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", ic.RepoURL, id,
		)
	}

	// Clone to temp, copy matching icon theme subdirs into the icons directory.
	// $1 = repo URL, $2 = glob pattern for subdirs, $3 = icons dir
	script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 "$1" "$TMP/pack"
shopt -s nullglob
for d in "$TMP/pack"/$2; do
    [ -d "$d" ] || continue
    cp -r "$d" "$3/"
    gtk-update-icon-cache -f -t "$3/$(basename "$d")" 2>/dev/null || true
done
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", ic.RepoURL, ic.CopyGlob, id,
	)
}

func removeIcon(ctx context.Context, exe *executor.Executor, stdout io.Writer, ic iconEntry, id string) error {
	// $1 = icons dir, $2 = glob pattern; nullglob prevents a no-match literal arg.
	// Run once per pattern so each glob stays a single, properly quoted argument.
	script := `
set -e
shopt -s nullglob
for d in "$1"/$2; do
    rm -rf -- "$d"
done
`
	for _, pattern := range append([]string{ic.DirPattern}, ic.ExtraDirPatterns...) {
		if err := exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", id, pattern,
		); err != nil {
			return err
		}
	}
	return nil
}

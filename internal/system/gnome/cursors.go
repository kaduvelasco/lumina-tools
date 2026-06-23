package gnome

import (
	"context"
	"io"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type cursorEntry struct {
	Name          string `yaml:"name"`
	DirPattern    string `yaml:"dir_pattern"`
	RepoURL       string `yaml:"repo_url"`
	Branch        string `yaml:"branch,omitempty"`
	HasInstall    bool   `yaml:"has_install,omitempty"`
	InstallSubDir string `yaml:"install_sub_dir,omitempty"`
	BuildScript   string `yaml:"build_script,omitempty"`
	CopyFrom      string `yaml:"copy_from,omitempty"`
	CopyGlob      string `yaml:"copy_glob,omitempty"`
}

func isCursorInstalled(ce cursorEntry, id string) bool {
	return globExists(filepath.Join(id, ce.DirPattern))
}

// ManageCursors shows a multi-select for cursor themes and applies the diff.
func ManageCursors(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar — Cursores")

	id, err := iconsDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de ícones: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	catalogue, err := loadCursorCatalogue()
	if err != nil {
		ui.Err(stdout, "Erro ao carregar catálogo de cursores: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	return manageCatalogue(ctx, exe, stdin, stdout,
		"Customizar — Cursores", id, catalogue,
		func(ce cursorEntry) string { return ce.Name },
		isCursorInstalled,
		installCursor,
		removeCursor,
		"Cursores atualizados!",
	)
}

func installCursor(ctx context.Context, exe *executor.Executor, stdout io.Writer, ce cursorEntry, id string) error {
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", "mkdir -p -- \"$1\"", "--", id,
	); err != nil {
		return err
	}

	branchFlag := ""
	if ce.Branch != "" {
		branchFlag = "--branch " + shellQuote(ce.Branch) + " "
	}

	if ce.HasInstall {
		subdir := "."
		if ce.InstallSubDir != "" {
			subdir = ce.InstallSubDir
		}
		script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 ` + branchFlag + `"$1" "$TMP/repo"
cd "$TMP/repo/` + subdir + `"
bash ./install.sh
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", ce.RepoURL,
		)
	}

	// Clone to temp, optionally build, then copy matching subdirs into the icons directory.
	// $1 = repo URL, $2 = glob pattern, $3 = icons dir
	copyFrom := ce.CopyFrom
	if copyFrom == "" {
		copyFrom = "."
	}
	buildStep := ""
	if ce.BuildScript != "" {
		buildStep = "\ncd \"$TMP/repo\"\n" + ce.BuildScript + "\n"
	}
	script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf -- "$TMP"' EXIT
git clone --depth=1 ` + branchFlag + `-- "$1" "$TMP/repo"` + buildStep + `
shopt -s nullglob
for d in "$TMP/repo/` + copyFrom + `"/$2; do
    [ -d "$d" ] || continue
    rm -rf -- "$3/$(basename -- "$d")"
    cp -r -- "$d" "$3/"
done
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", ce.RepoURL, ce.CopyGlob, id,
	)
}

func removeCursor(ctx context.Context, exe *executor.Executor, stdout io.Writer, ce cursorEntry, id string) error {
	// $1 = icons dir, $2 = glob pattern
	script := `
set -e
shopt -s nullglob
for d in "$1"/$2; do
    rm -rf -- "$d"
done
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", id, ce.DirPattern,
	)
}

package gnome

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type cursorEntry struct {
	Name          string
	DirPattern    string // glob under ~/.local/share/icons/ for detection and removal
	RepoURL       string
	Branch        string // non-empty: clone this branch
	HasInstall    bool   // true: run ./install.sh from InstallSubDir (or repo root)
	InstallSubDir string // subdir to cd into before running install.sh
	CopyFrom      string // non-empty: copy matching dirs from this subdir
	CopyGlob      string // non-empty: glob within CopyFrom (or repo root) to copy
}

var cursorCatalogue = []cursorEntry{
	{
		Name:       "Layan",
		DirPattern: "Layan-cursors",
		RepoURL:    "https://github.com/vinceliuice/Layan-cursors.git",
		HasInstall: true,
	},
	{
		Name:     "Oreo",
		DirPattern: "oreo*",
		RepoURL:  "https://github.com/varlesh/oreo-cursors.git",
		CopyFrom: "dist",
		CopyGlob: "oreo*",
	},
	{
		Name:       "Sweet",
		DirPattern: "Sweet-cursors*",
		RepoURL:    "https://github.com/EliverLara/Sweet.git",
		Branch:     "nova",
		CopyFrom:   "cursors",
		CopyGlob:   "*",
	},
	{
		Name:          "Colloid",
		DirPattern:    "Colloid-cursors*",
		RepoURL:       "https://github.com/vinceliuice/Colloid-icon-theme.git",
		HasInstall:    true,
		InstallSubDir: "cursors",
	},
	{
		Name:       "Future",
		DirPattern: "Future-cursors*",
		RepoURL:    "https://github.com/yeyushengfan258/Future-cursors.git",
		HasInstall: true,
	},
}

func isCursorInstalled(ce cursorEntry, id string) bool {
	return globExists(filepath.Join(id, ce.DirPattern))
}

// ManageCursors shows a multi-select for cursor themes and applies the diff.
func ManageCursors(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar GNOME — Cursores")

	if !isGnome() {
		ui.Err(stdout, ErrNotGnome.Error())
		ui.WaitEnter(stdout)
		return nil
	}

	id, err := iconsDir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretório de ícones: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Info(stdout, "Verificando cursores instalados...")
	items := make([]ui.SelectItem, len(cursorCatalogue))
	for i, ce := range cursorCatalogue {
		items[i] = ui.SelectItem{Label: ce.Name, ID: ce.Name, Selected: isCursorInstalled(ce, id)}
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

	var toInstall, toRemove []cursorEntry
	for i, item := range finalItems {
		ce := cursorCatalogue[i]
		wasInstalled := items[i].Selected
		switch {
		case item.Selected && !wasInstalled:
			toInstall = append(toInstall, ce)
		case !item.Selected && wasInstalled:
			toRemove = append(toRemove, ce)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "Customizar GNOME — Cursores")

	for _, ce := range toRemove {
		ui.Info(stdout, "Removendo "+ce.Name+"...")
		if rErr := removeCursor(ctx, exe, stdout, ce, id); rErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", ce.Name, rErr))
		}
	}

	for _, ce := range toInstall {
		ui.Info(stdout, "Instalando "+ce.Name+"...")
		if iErr := installCursor(ctx, exe, stdout, ce, id); iErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", ce.Name, iErr))
		}
	}

	ui.Success(stdout, "Cursores atualizados!")
	ui.WaitEnter(stdout)
	return nil
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
trap 'rm -rf "$TMP"' EXIT
git clone --depth=1 ` + branchFlag + `"$1" "$TMP/repo"
cd "$TMP/repo/` + subdir + `"
bash ./install.sh
`
		return exe.Run(ctx,
			executor.Options{Stdout: stdout, Stderr: stdout},
			"bash", "-c", script, "--", ce.RepoURL,
		)
	}

	// Clone to temp, copy matching subdirs into the icons directory.
	// $1 = repo URL, $2 = glob pattern, $3 = icons dir
	copyFrom := ce.CopyFrom
	if copyFrom == "" {
		copyFrom = "."
	}
	script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
git clone --depth=1 ` + branchFlag + `"$1" "$TMP/repo"
shopt -s nullglob
for d in "$TMP/repo/` + copyFrom + `"/$2; do
    [ -d "$d" ] || continue
    cp -r "$d" "$3/"
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

package gnome

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type iconEntry struct {
	Name       string
	DirPattern string // glob under ~/.local/share/icons/ for detection and removal
	RepoURL    string
	CloneAs    string // non-empty: clone directly as ~/.local/share/icons/<CloneAs>
	CopyGlob   string // non-empty: clone to tmp, copy matching subdirs into icons dir
}

var iconCatalogue = []iconEntry{
	{
		Name:       "Gruvbox Plus",
		DirPattern: "Gruvbox-Plus-*",
		RepoURL:    "https://github.com/SylEleuth/gruvbox-plus-icon-pack.git",
		CopyGlob:   "Gruvbox-Plus-*",
	},
	{
		Name:       "Kora",
		DirPattern: "kora",
		RepoURL:    "https://github.com/bikass/kora.git",
		CopyGlob:   "kora*",
	},
	{
		Name:       "Candy Icons",
		DirPattern: "candy-icons",
		RepoURL:    "https://github.com/EliverLara/candy-icons.git",
		CloneAs:    "candy-icons",
	},
	{
		Name:       "Flatery",
		DirPattern: "Flatery",
		RepoURL:    "https://github.com/cbrnix/Flatery.git",
		CopyGlob:   "Flatery*",
	},
	{
		Name:       "Newaita",
		DirPattern: "Newaita*",
		RepoURL:    "https://github.com/cbrnix/Newaita-reborn.git",
		CopyGlob:   "Newaita*",
	},
}

func isIconInstalled(ic iconEntry, id string) bool {
	return globExists(filepath.Join(id, ic.DirPattern))
}

// ManageIcons shows a multi-select for icon packs and applies the diff.
func ManageIcons(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar GNOME — Ícones")

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

	ui.Info(stdout, "Verificando pacotes de ícones instalados...")
	items := make([]ui.SelectItem, len(iconCatalogue))
	for i, ic := range iconCatalogue {
		items[i] = ui.SelectItem{Label: ic.Name, ID: ic.Name, Selected: isIconInstalled(ic, id)}
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

	var toInstall, toRemove []iconEntry
	for i, item := range finalItems {
		ic := iconCatalogue[i]
		wasInstalled := items[i].Selected
		switch {
		case item.Selected && !wasInstalled:
			toInstall = append(toInstall, ic)
		case !item.Selected && wasInstalled:
			toRemove = append(toRemove, ic)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "Customizar GNOME — Ícones")

	for _, ic := range toRemove {
		ui.Info(stdout, "Removendo "+ic.Name+"...")
		if rErr := removeIcon(ctx, exe, stdout, ic, id); rErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", ic.Name, rErr))
		}
	}

	for _, ic := range toInstall {
		ui.Info(stdout, "Instalando "+ic.Name+"...")
		if iErr := installIcon(ctx, exe, stdout, ic, id); iErr != nil {
			ui.Warning(stdout, fmt.Sprintf("Falha ao instalar %s: %v", ic.Name, iErr))
		}
	}

	ui.Success(stdout, "Ícones atualizados!")
	ui.WaitEnter(stdout)
	return nil
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

	// Clone to temp, copy matching icon theme subdirs into the icons directory.
	// $1 = repo URL, $2 = glob pattern for subdirs, $3 = icons dir
	script := `
set -e
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
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
	// $1 = icons dir, $2 = glob pattern; nullglob prevents a no-match literal arg
	script := `
set -e
shopt -s nullglob
for d in "$1"/$2; do
    rm -rf -- "$d"
done
`
	return exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"bash", "-c", script, "--", id, ic.DirPattern,
	)
}

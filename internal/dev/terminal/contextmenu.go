package terminal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// addContextMenuEntries creates "Open here" entries for Nautilus, Nemo and Dolphin.
func addContextMenuEntries(t Terminal, stdout io.Writer) {
	if t.DirFlag == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Warning(stdout, "Menu de contexto: falha ao obter diretório home: "+err.Error())
		return
	}
	ui.Info(stdout, "Integrando "+t.Name+" ao menu de contexto dos gerenciadores de arquivos...")
	writeContextMenuFile(nautilusScriptPath(t, home), nautilusScript(t), 0o755, "Nautilus", stdout)
	writeContextMenuFile(nemoScriptPath(t, home), nemoScript(t), 0o755, "Nemo", stdout)
	writeContextMenuFile(dolphinMenuPath(t, home), dolphinMenu(t), 0o644, "Dolphin", stdout)
}

// removeContextMenuEntries removes "Open here" entries for all file managers.
func removeContextMenuEntries(t Terminal, stdout io.Writer) {
	if t.DirFlag == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	for _, p := range []string{
		nautilusScriptPath(t, home),
		nemoScriptPath(t, home),
		dolphinMenuPath(t, home),
	} {
		_ = os.Remove(p)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeContextMenuFile(path, content string, perm os.FileMode, label string, stdout io.Writer) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		ui.Warning(stdout, label+": falha ao criar diretório: "+err.Error())
		return
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		ui.Warning(stdout, label+": falha ao escrever arquivo: "+err.Error())
	}
}

func execCmd(t Terminal) string {
	if t.MenuExec != "" {
		return t.MenuExec
	}
	return t.Cmd
}

func menuSlug(t Terminal) string {
	return strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
}

// shellescapeWords single-quotes each whitespace-separated word in s.
// Handles multi-word commands like "flatpak run com.example.App" and
// prevents shell injection if any word ever contains special characters.
func shellescapeWords(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = "'" + strings.ReplaceAll(w, "'", `'\''`) + "'"
	}
	return strings.Join(words, " ")
}

// ── Nautilus ─────────────────────────────────────────────────────────────────

func nautilusScriptPath(t Terminal, home string) string {
	return filepath.Join(home, ".local", "share", "nautilus", "scripts", "Abrir no "+t.Name)
}

// nautilusScript returns a bash script for the Nautilus Scripts menu.
// NAUTILUS_SCRIPT_CURRENT_URI holds the current folder as a file:// URI.
func nautilusScript(t Terminal) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
URI="${NAUTILUS_SCRIPT_CURRENT_URI:-${NEMO_SCRIPT_CURRENT_URI:-}}"
DIR=$(python3 -c "import sys,urllib.parse; u=sys.argv[1]; print(urllib.parse.unquote(u[7:] if u.startswith('file://') else u))" "$URI" 2>/dev/null)
[[ -z "$DIR" ]] && DIR="$HOME"
%s %s "$DIR"
`, shellescapeWords(execCmd(t)), shellescapeWords(t.DirFlag))
}

// ── Nemo ─────────────────────────────────────────────────────────────────────

func nemoScriptPath(t Terminal, home string) string {
	return filepath.Join(home, ".local", "share", "nemo", "scripts", "Abrir no "+t.Name)
}

// nemoScript returns a bash script for the Nemo Scripts menu.
// NEMO_SCRIPT_CURRENT_URI holds the current folder as a file:// URI.
func nemoScript(t Terminal) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
URI="${NEMO_SCRIPT_CURRENT_URI:-${NAUTILUS_SCRIPT_CURRENT_URI:-}}"
DIR=$(python3 -c "import sys,urllib.parse; u=sys.argv[1]; print(urllib.parse.unquote(u[7:] if u.startswith('file://') else u))" "$URI" 2>/dev/null)
[[ -z "$DIR" ]] && DIR="$HOME"
%s %s "$DIR"
`, shellescapeWords(execCmd(t)), shellescapeWords(t.DirFlag))
}

// ── Dolphin ───────────────────────────────────────────────────────────────────

func dolphinMenuPath(t Terminal, home string) string {
	return filepath.Join(home, ".local", "share", "kio", "servicemenus", "lumina-"+menuSlug(t)+".desktop")
}

// dolphinMenu returns a KIO service menu .desktop file.
// MimeType=inode/directory shows the entry when right-clicking any directory
// (including the background of the current folder in Dolphin).
func dolphinMenu(t Terminal) string {
	// Action ID must be unique within the file; prefix avoids collisions with other service menus.
	actionID := "lumina-open-" + menuSlug(t)
	return fmt.Sprintf(`[Desktop Entry]
Type=Service
ServiceTypes=KonqPopupMenu/Plugin
MimeType=inode/directory;
Actions=%s;

[Desktop Action %s]
Name=Abrir no %s
Icon=utilities-terminal
Exec=%s %s %%f
`, actionID, actionID, t.Name, execCmd(t), t.DirFlag)
}

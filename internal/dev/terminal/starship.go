package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var starshipPresets = []string{
	"gruvbox-rainbow",
	"tokyo-night",
	"pastel-powerline",
	"pure-preset",
}

// installStarship downloads Starship to ~/.local/bin, optionally applies a preset,
// and registers the init hook in every shell config file found on disk.
// When ~/.config/starship.toml already exists (update path), the preset prompt is skipped.
func installStarship(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	script := `
set -e
mkdir -p "$HOME/.local/bin"
curl -fsSL https://starship.rs/install.sh | sh -s -- --yes --bin-dir "$HOME/.local/bin"
`
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(filepath.Join(home, ".config", "starship.toml")); os.IsNotExist(statErr) {
		fmt.Fprintln(stdout, "\nEscolha um preset do Starship:")
		for i, p := range starshipPresets {
			fmt.Fprintf(stdout, "  %d. %s\n", i+1, p)
		}
		fmt.Fprintf(stdout, "\nEnter = %s: ", starshipPresets[0])

		choice := strings.TrimSpace(prompt.ReadLine())
		preset := starshipPresets[0]
		if sel := prompt.ParseSelection(choice, len(starshipPresets)); len(sel) > 0 {
			preset = starshipPresets[sel[0]]
		}

		presetScript := `"$HOME/.local/bin/starship" preset ` + preset + ` -o "$HOME/.config/starship.toml"`
		if pErr := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", presetScript); pErr != nil {
			ui.Warning(stdout, "Falha ao aplicar preset: "+pErr.Error())
		}
	}

	configureStarshipShells(stdout)
	return nil
}

// uninstallStarship removes the Starship binary, config file, and shell init hooks.
func uninstallStarship(stdout io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	os.Remove(filepath.Join(home, ".local", "bin", "starship"))
	os.Remove(filepath.Join(home, ".config", "starship.toml"))

	stripStarshipLine(filepath.Join(home, ".bashrc"), `eval "$(starship init bash)"`)
	stripStarshipLine(filepath.Join(home, ".zshrc"), `eval "$(starship init zsh)"`)
	stripStarshipLine(filepath.Join(home, ".config", "fish", "config.fish"), `starship init fish | source`)

	ui.Info(stdout, "Configurações do Starship removidas dos shells.")
	return nil
}

// configureStarshipShells appends the init hook to each shell config that exists.
func configureStarshipShells(stdout io.Writer) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	type entry struct {
		rc   string
		line string
	}
	shells := []entry{
		{filepath.Join(home, ".bashrc"), `eval "$(starship init bash)"`},
		{filepath.Join(home, ".zshrc"), `eval "$(starship init zsh)"`},
	}
	for _, s := range shells {
		if _, statErr := os.Stat(s.rc); statErr == nil {
			if appendLineIfMissing(s.rc, s.line) {
				ui.Info(stdout, "Starship registrado em "+s.rc)
			}
		}
	}

	fishConfig := filepath.Join(home, ".config", "fish", "config.fish")
	if _, statErr := os.Stat(fishConfig); statErr == nil {
		if appendLineIfMissing(fishConfig, `starship init fish | source`) {
			ui.Info(stdout, "Starship registrado em "+fishConfig)
		}
	}
}

// appendLineIfMissing appends line to path when not already present.
// Returns true when the line was actually written.
func appendLineIfMissing(path, line string) bool {
	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), line) {
		return false
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	defer f.Close()
	fmt.Fprintf(f, "\n%s\n", line)
	return true
}

// stripStarshipLine removes exact-match lines from a shell config file.
func stripStarshipLine(path, line string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	rawLines := strings.Split(string(data), "\n")
	kept := rawLines[:0]
	for _, l := range rawLines {
		if strings.TrimSpace(l) != strings.TrimSpace(line) {
			kept = append(kept, l)
		}
	}
	os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644)
}

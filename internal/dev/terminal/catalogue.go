package terminal

import (
	"context"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// Terminal describes a terminal emulator managed by lumina.
type Terminal struct {
	Name     string
	Cmd      string
	FlatID   string // non-empty when installed via Flatpak
	DirFlag  string // flag to open in a directory (empty = no context menu, e.g. Starship)
	MenuExec string // exec command for context menus; defaults to Cmd when empty
}

// Catalogue lists all terminals and shell tools managed by lumina.
var Catalogue = []Terminal{
	{Name: "Kitty", Cmd: "kitty", DirFlag: "--directory"},
	{Name: "Alacritty", Cmd: "alacritty", DirFlag: "--working-directory"},
	{Name: "Black Box", Cmd: "blackbox-terminal", FlatID: "com.raggesilver.BlackBox", DirFlag: "--working-directory", MenuExec: "flatpak run com.raggesilver.BlackBox"},
	{Name: "GNOME Console", Cmd: "kgx", DirFlag: "--working-directory"},
	{Name: "Starship", Cmd: "starship"},
}

// InstalledMap returns which terminals are currently installed (by Name).
func InstalledMap(ctx context.Context, exe *executor.Executor) map[string]bool {
	result := make(map[string]bool, len(Catalogue))
	flatpakOut, _ := exe.Output(ctx, executor.Options{},
		"flatpak", "list", config.FlatpakFlag(), "--app", "--columns=application")
	for _, t := range Catalogue {
		if t.FlatID != "" && contains(flatpakOut, t.FlatID) {
			result[t.Name] = true
			continue
		}
		if _, err := exe.Output(ctx, executor.Options{}, "which", t.Cmd); err == nil {
			result[t.Name] = true
		}
	}
	return result
}

func contains(haystack, needle string) bool {
	for _, line := range strings.Split(haystack, "\n") {
		if strings.TrimSpace(line) == needle {
			return true
		}
	}
	return false
}

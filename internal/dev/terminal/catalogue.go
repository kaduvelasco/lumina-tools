package terminal

import (
	"context"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// Terminal describes a terminal emulator managed by lumina.
type Terminal struct {
	Name   string
	Cmd    string
	FlatID string // non-empty when installed via Flatpak
}

// Catalogue lists all terminals managed by lumina.
var Catalogue = []Terminal{
	{Name: "Kitty", Cmd: "kitty", FlatID: ""},
	{Name: "Alacritty", Cmd: "alacritty", FlatID: ""},
	{Name: "Black Box", Cmd: "blackbox-terminal", FlatID: "com.raggesilver.BlackBox"},
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

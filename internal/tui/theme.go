package tui

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines a color palette for the TUI.
type Theme struct {
	Name    string
	Primary lipgloss.Color // dividers, borders, breadcrumb
	Accent  lipgloss.Color // active item bar and text
	Text    lipgloss.Color // inactive item text
	Muted   lipgloss.Color // footer hints
	Badge   lipgloss.Color // key badge background in the footer
	Success lipgloss.Color
	Err     lipgloss.Color
	Warning lipgloss.Color
}

// TUIStyles holds pre-built lipgloss styles derived from a Theme.
// The Model stores one instance and rebuilds it whenever the theme changes.
type TUIStyles struct {
	ActiveBar  lipgloss.Style
	ActiveText lipgloss.Style
	Inactive   lipgloss.Style
	Breadcrumb lipgloss.Style
	Footer     lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	HeaderBox  lipgloss.Style // rounded-border container for the persistent header
	KeyBadge   lipgloss.Style // keyboard shortcut badge in the footer
	FooterBox  lipgloss.Style // footer container with rounded border (mirrors HeaderBox)
}

// availableThemes is the ordered list of themes presented to the user.
var availableThemes = []Theme{
	{
		Name:    "Lumina",
		Primary: "#9966FF",
		Accent:  "#FF99FF",
		Text:    "#FFFFFF",
		Muted:   "#666666",
		Badge:   "#2a2a3a",
		Success: "#00FF88",
		Err:     "#FF4466",
		Warning: "#FFAA00",
	},
	{
		Name:    "Claro",
		Primary: "#5500CC",
		Accent:  "#7700EE",
		Text:    "#111111",
		Muted:   "#777777",
		Badge:   "#e0e0e0",
		Success: "#006633",
		Err:     "#BB0000",
		Warning: "#AA5500",
	},
	{
		Name:    "Dracula",
		Primary: "#BD93F9",
		Accent:  "#FF79C6",
		Text:    "#F8F8F2",
		Muted:   "#6272A4",
		Badge:   "#282a36",
		Success: "#50FA7B",
		Err:     "#FF5555",
		Warning: "#FFB86C",
	},
	{
		Name:    "Nord",
		Primary: "#81A1C1",
		Accent:  "#88C0D0",
		Text:    "#ECEFF4",
		Muted:   "#4C566A",
		Badge:   "#2e3440",
		Success: "#A3BE8C",
		Err:     "#BF616A",
		Warning: "#EBCB8B",
	},
	{
		Name:    "Tokyo Night",
		Primary: "#7AA2F7",
		Accent:  "#BB9AF7",
		Text:    "#C0CAF5",
		Muted:   "#565F89",
		Badge:   "#1f2335",
		Success: "#9ECE6A",
		Err:     "#F7768E",
		Warning: "#E0AF68",
	},
	{
		Name:    "Gruvbox",
		Primary: "#D79921",
		Accent:  "#FABD2F",
		Text:    "#EBDBB2",
		Muted:   "#928374",
		Badge:   "#282828",
		Success: "#B8BB26",
		Err:     "#FB4934",
		Warning: "#FE8019",
	},
}

// buildStyles creates a TUIStyles from the given Theme.
func buildStyles(t Theme) TUIStyles {
	return TUIStyles{
		ActiveBar:  lipgloss.NewStyle().Foreground(t.Accent),
		ActiveText: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Inactive:   lipgloss.NewStyle().Foreground(t.Text),
		Breadcrumb: lipgloss.NewStyle().Foreground(t.Primary),
		Footer:     lipgloss.NewStyle().Foreground(t.Muted),
		Success:    lipgloss.NewStyle().Foreground(t.Success),
		Error:      lipgloss.NewStyle().Foreground(t.Err),
		Warning:    lipgloss.NewStyle().Foreground(t.Warning),
		HeaderBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 2),
		KeyBadge: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Badge).
			Padding(0, 1),
		FooterBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 2),
	}
}

// detectDefaultTheme picks an initial theme based on terminal background signals.
// COLORFGBG is set by many terminals: last segment >= 7 indicates a light background.
func detectDefaultTheme() Theme {
	if v := os.Getenv("COLORFGBG"); v != "" {
		parts := strings.Split(v, ";")
		if len(parts) >= 2 {
			bg, err := strconv.Atoi(parts[len(parts)-1])
			if err == nil && bg >= 7 {
				return themeByName("Claro")
			}
		}
	}
	return availableThemes[0]
}

// themeByName looks up a theme by name; falls back to the first theme.
func themeByName(name string) Theme {
	for _, t := range availableThemes {
		if t.Name == name {
			return t
		}
	}
	return availableThemes[0]
}

package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// ── logo ─────────────────────────────────────────────────────────────────────

var logoLines = []string{
	"    \x1b[38;2;245;155;255m░\x1b[0m\x1b[38;2;242;155;255m█\x1b[0m\x1b[38;2;240;156;255m█\x1b[0m                                                  \x1b[38;2;116;183;255m░\x1b[0m\x1b[38;2;114;184;255m█\x1b[0m\x1b[38;2;111;184;255m█\x1b[0m    ",
	"  \x1b[38;2;250;154;255m░\x1b[0m\x1b[38;2;247;154;255m█\x1b[0m\x1b[38;2;245;155;255m█\x1b[0m  \x1b[38;2;238;156;255m░\x1b[0m\x1b[38;2;235;157;255m█\x1b[0m\x1b[38;2;233;157;255m█\x1b[0m      \x1b[38;2;216;161;255m░\x1b[0m\x1b[38;2;213;162;255m█\x1b[0m\x1b[38;2;211;162;255m█\x1b[0m   \x1b[38;2;201;164;255m░\x1b[0m\x1b[38;2;199;165;255m█\x1b[0m\x1b[38;2;196;165;255m█\x1b[0m\x1b[38;2;194;166;255m░\x1b[0m\x1b[38;2;191;167;255m█\x1b[0m\x1b[38;2;189;167;255m█\x1b[0m     \x1b[38;2;174;170;255m░\x1b[0m\x1b[38;2;172;171;255m█\x1b[0m\x1b[38;2;170;171;255m█\x1b[0m\x1b[38;2;167;172;255m░\x1b[0m\x1b[38;2;165;172;255m█\x1b[0m\x1b[38;2;162;173;255m█\x1b[0m\x1b[38;2;160;174;255m░\x1b[0m\x1b[38;2;157;174;255m█\x1b[0m\x1b[38;2;155;175;255m█\x1b[0m   \x1b[38;2;145;177;255m░\x1b[0m\x1b[38;2;143;177;255m█\x1b[0m\x1b[38;2;140;178;255m█\x1b[0m  \x1b[38;2;133;179;255m░\x1b[0m\x1b[38;2;131;180;255m█\x1b[0m\x1b[38;2;128;181;255m█\x1b[0m\x1b[38;2;126;181;255m█\x1b[0m\x1b[38;2;123;182;255m█\x1b[0m    \x1b[38;2;111;184;255m░\x1b[0m\x1b[38;2;109;185;255m█\x1b[0m\x1b[38;2;106;185;255m█\x1b[0m  ",
	"    \x1b[38;2;245;155;255m░\x1b[0m\x1b[38;2;242;155;255m█\x1b[0m\x1b[38;2;240;156;255m█\x1b[0m\x1b[38;2;238;156;255m░\x1b[0m\x1b[38;2;235;157;255m█\x1b[0m\x1b[38;2;233;157;255m█\x1b[0m      \x1b[38;2;216;161;255m░\x1b[0m\x1b[38;2;213;162;255m█\x1b[0m\x1b[38;2;211;162;255m█\x1b[0m   \x1b[38;2;201;164;255m░\x1b[0m\x1b[38;2;199;165;255m█\x1b[0m\x1b[38;2;196;165;255m█\x1b[0m\x1b[38;2;194;166;255m░\x1b[0m\x1b[38;2;191;167;255m█\x1b[0m\x1b[38;2;189;167;255m█\x1b[0m\x1b[38;2;187;168;255m█\x1b[0m\x1b[38;2;184;168;255m█\x1b[0m \x1b[38;2;179;169;255m░\x1b[0m\x1b[38;2;177;170;255m█\x1b[0m\x1b[38;2;174;170;255m█\x1b[0m\x1b[38;2;172;171;255m█\x1b[0m\x1b[38;2;170;171;255m█\x1b[0m\x1b[38;2;167;172;255m░\x1b[0m\x1b[38;2;165;172;255m█\x1b[0m\x1b[38;2;162;173;255m█\x1b[0m\x1b[38;2;160;174;255m░\x1b[0m\x1b[38;2;157;174;255m█\x1b[0m\x1b[38;2;155;175;255m█\x1b[0m\x1b[38;2;153;175;255m█\x1b[0m\x1b[38;2;150;176;255m█\x1b[0m \x1b[38;2;145;177;255m░\x1b[0m\x1b[38;2;143;177;255m█\x1b[0m\x1b[38;2;140;178;255m█\x1b[0m\x1b[38;2;138;178;255m░\x1b[0m\x1b[38;2;136;179;255m█\x1b[0m\x1b[38;2;133;179;255m█\x1b[0m   \x1b[38;2;123;182;255m░\x1b[0m\x1b[38;2;121;182;255m█\x1b[0m\x1b[38;2;119;183;255m█\x1b[0m\x1b[38;2;116;183;255m░\x1b[0m\x1b[38;2;114;184;255m█\x1b[0m\x1b[38;2;111;184;255m█\x1b[0m    ",
	"\x1b[38;2;255;153;255m░\x1b[0m\x1b[38;2;252;153;255m█\x1b[0m\x1b[38;2;250;154;255m█\x1b[0m\x1b[38;2;247;154;255m█\x1b[0m\x1b[38;2;245;155;255m█\x1b[0m  \x1b[38;2;238;156;255m░\x1b[0m\x1b[38;2;235;157;255m█\x1b[0m\x1b[38;2;233;157;255m█\x1b[0m      \x1b[38;2;216;161;255m░\x1b[0m\x1b[38;2;213;162;255m█\x1b[0m\x1b[38;2;211;162;255m█\x1b[0m   \x1b[38;2;201;164;255m░\x1b[0m\x1b[38;2;199;165;255m█\x1b[0m\x1b[38;2;196;165;255m█\x1b[0m\x1b[38;2;194;166;255m░\x1b[0m\x1b[38;2;191;167;255m█\x1b[0m\x1b[38;2;189;167;255m█\x1b[0m \x1b[38;2;184;168;255m░\x1b[0m\x1b[38;2;182;169;255m█\x1b[0m\x1b[38;2;179;169;255m█\x1b[0m \x1b[38;2;174;170;255m░\x1b[0m\x1b[38;2;172;171;255m█\x1b[0m\x1b[38;2;170;171;255m█\x1b[0m\x1b[38;2;167;172;255m░\x1b[0m\x1b[38;2;165;172;255m█\x1b[0m\x1b[38;2;162;173;255m█\x1b[0m\x1b[38;2;160;174;255m░\x1b[0m\x1b[38;2;157;174;255m█\x1b[0m\x1b[38;2;155;175;255m█\x1b[0m \x1b[38;2;150;176;255m░\x1b[0m\x1b[38;2;148;176;255m█\x1b[0m\x1b[38;2;145;177;255m█\x1b[0m\x1b[38;2;143;177;255m█\x1b[0m\x1b[38;2;140;178;255m█\x1b[0m\x1b[38;2;138;178;255m░\x1b[0m\x1b[38;2;136;179;255m█\x1b[0m\x1b[38;2;133;179;255m█\x1b[0m\x1b[38;2;131;180;255m█\x1b[0m\x1b[38;2;128;181;255m█\x1b[0m\x1b[38;2;126;181;255m█\x1b[0m\x1b[38;2;123;182;255m█\x1b[0m\x1b[38;2;121;182;255m█\x1b[0m\x1b[38;2;119;183;255m█\x1b[0m  \x1b[38;2;111;184;255m░\x1b[0m\x1b[38;2;109;185;255m█\x1b[0m\x1b[38;2;106;185;255m█\x1b[0m\x1b[38;2;104;186;255m█\x1b[0m\x1b[38;2;102;187;255m█\x1b[0m",
	"    \x1b[38;2;245;155;255m░\x1b[0m\x1b[38;2;242;155;255m█\x1b[0m\x1b[38;2;240;156;255m█\x1b[0m\x1b[38;2;238;156;255m░\x1b[0m\x1b[38;2;235;157;255m█\x1b[0m\x1b[38;2;233;157;255m█\x1b[0m      \x1b[38;2;216;161;255m░\x1b[0m\x1b[38;2;213;162;255m█\x1b[0m\x1b[38;2;211;162;255m█\x1b[0m   \x1b[38;2;201;164;255m░\x1b[0m\x1b[38;2;199;165;255m█\x1b[0m\x1b[38;2;196;165;255m█\x1b[0m\x1b[38;2;194;166;255m░\x1b[0m\x1b[38;2;191;167;255m█\x1b[0m\x1b[38;2;189;167;255m█\x1b[0m     \x1b[38;2;174;170;255m░\x1b[0m\x1b[38;2;172;171;255m█\x1b[0m\x1b[38;2;170;171;255m█\x1b[0m\x1b[38;2;167;172;255m░\x1b[0m\x1b[38;2;165;172;255m█\x1b[0m\x1b[38;2;162;173;255m█\x1b[0m\x1b[38;2;160;174;255m░\x1b[0m\x1b[38;2;157;174;255m█\x1b[0m\x1b[38;2;155;175;255m█\x1b[0m   \x1b[38;2;145;177;255m░\x1b[0m\x1b[38;2;143;177;255m█\x1b[0m\x1b[38;2;140;178;255m█\x1b[0m\x1b[38;2;138;178;255m░\x1b[0m\x1b[38;2;136;179;255m█\x1b[0m\x1b[38;2;133;179;255m█\x1b[0m   \x1b[38;2;123;182;255m░\x1b[0m\x1b[38;2;121;182;255m█\x1b[0m\x1b[38;2;119;183;255m█\x1b[0m\x1b[38;2;116;183;255m░\x1b[0m\x1b[38;2;114;184;255m█\x1b[0m\x1b[38;2;111;184;255m█\x1b[0m    ",
	"  \x1b[38;2;250;154;255m░\x1b[0m\x1b[38;2;247;154;255m█\x1b[0m\x1b[38;2;245;155;255m█\x1b[0m  \x1b[38;2;238;156;255m░\x1b[0m\x1b[38;2;235;157;255m█\x1b[0m\x1b[38;2;233;157;255m█\x1b[0m\x1b[38;2;230;158;255m█\x1b[0m\x1b[38;2;228;158;255m█\x1b[0m\x1b[38;2;225;159;255m█\x1b[0m\x1b[38;2;223;160;255m█\x1b[0m\x1b[38;2;221;160;255m█\x1b[0m\x1b[38;2;218;161;255m█\x1b[0m  \x1b[38;2;211;162;255m░\x1b[0m\x1b[38;2;208;163;255m█\x1b[0m\x1b[38;2;206;163;255m█\x1b[0m\x1b[38;2;204;164;255m█\x1b[0m\x1b[38;2;201;164;255m█\x1b[0m  \x1b[38;2;194;166;255m░\x1b[0m\x1b[38;2;191;167;255m█\x1b[0m\x1b[38;2;189;167;255m█\x1b[0m     \x1b[38;2;174;170;255m░\x1b[0m\x1b[38;2;172;171;255m█\x1b[0m\x1b[38;2;170;171;255m█\x1b[0m\x1b[38;2;167;172;255m░\x1b[0m\x1b[38;2;165;172;255m█\x1b[0m\x1b[38;2;162;173;255m█\x1b[0m\x1b[38;2;160;174;255m░\x1b[0m\x1b[38;2;157;174;255m█\x1b[0m\x1b[38;2;155;175;255m█\x1b[0m   \x1b[38;2;145;177;255m░\x1b[0m\x1b[38;2;143;177;255m█\x1b[0m\x1b[38;2;140;178;255m█\x1b[0m\x1b[38;2;138;178;255m░\x1b[0m\x1b[38;2;136;179;255m█\x1b[0m\x1b[38;2;133;179;255m█\x1b[0m   \x1b[38;2;123;182;255m░\x1b[0m\x1b[38;2;121;182;255m█\x1b[0m\x1b[38;2;119;183;255m█\x1b[0m  \x1b[38;2;111;184;255m░\x1b[0m\x1b[38;2;109;185;255m█\x1b[0m\x1b[38;2;106;185;255m█\x1b[0m  ",
	"    \x1b[38;2;245;155;255m░\x1b[0m\x1b[38;2;242;155;255m█\x1b[0m\x1b[38;2;240;156;255m█\x1b[0m                                                  \x1b[38;2;116;183;255m░\x1b[0m\x1b[38;2;114;184;255m█\x1b[0m\x1b[38;2;111;184;255m█\x1b[0m    ",
}

var cachedHeader = strings.Join(logoLines, "\n") + "\n"

// RenderHeader returns the pre-rendered gradient logo string.
func RenderHeader() string { return cachedHeader }

// ── terminal width ────────────────────────────────────────────────────────────

func termWidth(w io.Writer) int {
	width, _, err := term.GetSize(writerFD(w))
	if err != nil || width < 40 {
		return 80
	}
	if width > 120 {
		return 120
	}
	return width
}

func writerFD(w io.Writer) int {
	type fder interface{ Fd() uintptr }
	if f, ok := w.(fder); ok {
		return int(f.Fd())
	}
	return int(os.Stdout.Fd())
}

// ── theme ─────────────────────────────────────────────────────────────────────

// ScriptTheme is the color palette used by ui.* helpers and by any other
// package that renders text outside the main TUI render loop (via tea.Exec).
// Exported so those packages can stay in sync with the user's selected theme
// instead of hardcoding their own colors.
type ScriptTheme struct {
	Primary lipgloss.Color
	Accent  lipgloss.Color
	Muted   lipgloss.Color
	Success lipgloss.Color
	Err     lipgloss.Color
	Warning lipgloss.Color
}

// scriptThemes mirrors the color palette in internal/tui/theme.go.
var scriptThemes = map[string]ScriptTheme{
	"Lumina":      {"#9966FF", "#FF99FF", "#666666", "#00FF88", "#FF4466", "#FFAA00"},
	"Claro":       {"#5500CC", "#7700EE", "#777777", "#006633", "#BB0000", "#AA5500"},
	"Dracula":     {"#BD93F9", "#FF79C6", "#6272A4", "#50FA7B", "#FF5555", "#FFB86C"},
	"Nord":        {"#81A1C1", "#88C0D0", "#4C566A", "#A3BE8C", "#BF616A", "#EBCB8B"},
	"Tokyo Night": {"#7AA2F7", "#BB9AF7", "#565F89", "#9ECE6A", "#F7768E", "#E0AF68"},
	"Gruvbox":     {"#D79921", "#FABD2F", "#928374", "#B8BB26", "#FB4934", "#FE8019"},
}

// LoadTheme returns the color palette for the user's currently configured
// theme (~/.lumina/config.yaml), falling back to "Lumina" when unset or
// unrecognized. Reads the config fresh on every call, so a theme change saved
// by the TUI is picked up immediately by the next script that runs.
func LoadTheme() ScriptTheme {
	cfg, err := config.Load()
	if err != nil || cfg.Theme == "" {
		return scriptThemes["Lumina"]
	}
	t, ok := scriptThemes[cfg.Theme]
	if !ok {
		return scriptThemes["Lumina"]
	}
	return t
}

// ── PrintHeader ───────────────────────────────────────────────────────────────

// PrintHeader clears the terminal and renders the chrome-style header bar:
// left = brand identity, right = action title (mirrors the TUI v2 header).
// Colors follow the theme saved in the user's config.
func PrintHeader(w io.Writer, title string) {
	fmt.Fprint(w, "\033[2J\033[H")
	t := LoadTheme()

	sAccent := lipgloss.NewStyle().Foreground(t.Accent)
	sMuted := lipgloss.NewStyle().Foreground(t.Muted)
	sPrimary := lipgloss.NewStyle().Foreground(t.Primary)
	sBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(0, 2)

	width := termWidth(w)
	left := lipgloss.JoinHorizontal(lipgloss.Center,
		sAccent.Render("◈ "),
		sMuted.Render("lumina"),
		sPrimary.Bold(true).Render(".tools"),
		sMuted.Render("  │  "),
		sMuted.Render(version.Version),
	)
	right := sMuted.Render(title)

	// HeaderBox: Padding(0,2)=4 + Border=2 → content = width-6.
	contentW := width - 6
	if contentW < 20 {
		contentW = 20
	}
	spacerW := contentW - lipgloss.Width(left) - lipgloss.Width(right)
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := lipgloss.NewStyle().Width(spacerW).Render("")
	line := lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)

	fmt.Fprintln(w, sBox.Render(line))
	fmt.Fprintln(w)
}

// ── panels ────────────────────────────────────────────────────────────────────

func panel(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)
}

func printPanel(w io.Writer, style lipgloss.Style, text string) {
	// Width(n) in lipgloss v1 includes padding; border adds 2.
	// Panel has Padding(0,1)=2. Width(width-2) → visual = (width-2)+2 = width.
	width := termWidth(w)
	contentW := width - 2
	if contentW < 10 {
		contentW = 10
	}
	fmt.Fprintln(w, style.Width(contentW).Render(text))
}

// Info prints a primary-colored bordered panel for general messages.
func Info(w io.Writer, text string) {
	t := LoadTheme()
	printPanel(w, panel(t.Primary), text)
}

// Err prints an error-colored bordered panel.
func Err(w io.Writer, text string) {
	t := LoadTheme()
	printPanel(w, panel(t.Err), text)
}

// Warning prints a warning-colored bordered panel.
func Warning(w io.Writer, text string) {
	t := LoadTheme()
	printPanel(w, panel(t.Warning), text)
}

// Success prints a success-colored bordered panel.
func Success(w io.Writer, text string) {
	t := LoadTheme()
	printPanel(w, panel(t.Success), text)
}

// PrintBox renders content inside a primary-color bordered panel at full terminal width.
func PrintBox(w io.Writer, content string) {
	t := LoadTheme()
	printPanel(w, panel(t.Primary), content)
}

// ── WaitEnter ────────────────────────────────────────────────────────────────

// WaitEnter prints a themed panel and blocks until the user presses Enter.
// Reads from os.Stdin directly — safe inside tea.Exec where the real terminal is active.
func WaitEnter(w io.Writer) {
	t := LoadTheme()
	printPanel(w, panel(t.Muted), "Pressione ENTER para continuar...")
	r := bufio.NewReader(os.Stdin)
	_, _ = r.ReadString('\n')
}

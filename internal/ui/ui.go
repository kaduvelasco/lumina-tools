package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
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

// ── shared styles ─────────────────────────────────────────────────────────────

var (
	styleDivider = lipgloss.NewStyle().Foreground(lipgloss.Color("#9966FF"))
	styleTitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF99FF")).Bold(true)

	panelInfo    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#9966FF")).Padding(0, 1)
	panelError   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF4466")).Padding(0, 1)
	panelWarning = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FFAA00")).Padding(0, 1)
	panelSuccess = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00FF88")).Padding(0, 1)
)

// ── PrintHeader ───────────────────────────────────────────────────────────────

// PrintHeader clears the terminal, prints the LUMINA logo, then a
// divider / "LUMINA TOOLS :: title" / divider section.
func PrintHeader(w io.Writer, title string) {
	fmt.Fprint(w, "\033[2J\033[H")
	fmt.Fprint(w, cachedHeader)
	width := termWidth(w)
	div := styleDivider.Render(strings.Repeat("-", width))
	fmt.Fprintln(w, div)
	fmt.Fprintln(w, styleTitle.Render("LUMINA TOOLS :: "+title))
	fmt.Fprintln(w, div)
	fmt.Fprintln(w)
}

// ── panels ────────────────────────────────────────────────────────────────────

func printPanel(w io.Writer, style lipgloss.Style, text string) {
	maxW := termWidth(w) - 4
	fmt.Fprintln(w, style.MaxWidth(maxW).Render(text))
}

// Info prints a purple-bordered panel for general messages.
func Info(w io.Writer, text string) { printPanel(w, panelInfo, text) }

// Err prints a red-bordered panel for errors.
func Err(w io.Writer, text string) { printPanel(w, panelError, text) }

// Warning prints a yellow-bordered panel for warnings.
func Warning(w io.Writer, text string) { printPanel(w, panelWarning, text) }

// Success prints a green-bordered panel for success messages.
func Success(w io.Writer, text string) { printPanel(w, panelSuccess, text) }

// ── WaitEnter ────────────────────────────────────────────────────────────────

// WaitEnter prints a divider and blocks until the user presses Enter.
// Reads from os.Stdin directly — safe inside tea.Exec where the real terminal is active.
func WaitEnter(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, styleDivider.Render(strings.Repeat("-", termWidth(w))))
	fmt.Fprint(w, "  Pressione ENTER para continuar...")
	r := bufio.NewReader(os.Stdin)
	_, _ = r.ReadString('\n')
	fmt.Fprintln(w)
}

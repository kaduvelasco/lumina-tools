package tui

import (
	"context"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// Chrome for the v2 layout (persistent header + contextual footer).
// Pure render functions: given layout/theme state they return a styled string.

// chromeBrand is the v2 header's left-aligned product identity (plan §2).
const chromeBrand = "◈ lumina.tools"

// connectivityHost is resolved to determine internet access — github.com is a
// host the project is already coupled to (self-update relies on GitHub
// Releases), so a failed lookup is a meaningful signal (plan §6.2).
const connectivityHost = "github.com"

const (
	connectivityTimeout = 2 * time.Second
	connectivityPeriod  = 45 * time.Second
)

// connectivityMsg reports the outcome of a connectivity check.
type connectivityMsg struct{ online bool }

// connectivityTickMsg requests the next periodic connectivity re-check.
type connectivityTickMsg struct{}

// checkConnectivity resolves connectivityHost with a short timeout and
// reports a binary outcome — no "degraded" state (plan §6.2).
func checkConnectivity(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		cctx, cancel := context.WithTimeout(ctx, connectivityTimeout)
		defer cancel()
		_, err := net.DefaultResolver.LookupHost(cctx, connectivityHost)
		return connectivityMsg{online: err == nil}
	}
}

// tickConnectivity schedules the next periodic connectivity re-check.
func tickConnectivity() tea.Cmd {
	return tea.Tick(connectivityPeriod, func(time.Time) tea.Msg {
		return connectivityTickMsg{}
	})
}

// renderChromeHeader renders the persistent header inside a rounded border box.
// Left side: ◈ lumina.tools  │  version  │  distro (when set).
// Right side: ● (green/red) + dim status text.
func renderChromeHeader(width int, online bool, distroLabel string, s TUIStyles) string {
	// Left: brand identity split into styled segments.
	left := lipgloss.JoinHorizontal(lipgloss.Center,
		s.ActiveBar.Render("◈ "),
		s.Inactive.Render("lumina"),
		s.Breadcrumb.Bold(true).Render(".tools"),
		s.Footer.Render("  │  "),
		s.Footer.Render(version.Version),
	)
	if distroLabel != "" {
		left = lipgloss.JoinHorizontal(lipgloss.Center,
			left,
			s.Footer.Render("  │  "),
			s.Breadcrumb.Render(distroLabel),
		)
	}

	// Right: colored dot + dim status text (separate styles like the reference).
	var right string
	if online {
		right = s.Success.Render("●") + s.Footer.Render(" Conexão com internet ativa")
	} else {
		right = s.Error.Render("●") + s.Footer.Render(" Conexão com internet inativa")
	}

	// Content area: total width − border(2) − padding(4) = width − 6.
	contentW := width - 6
	if contentW < 20 {
		contentW = 20
	}

	// Fixed-width spacer between left and right — avoids word-wrap issues
	// that arise when passing a pre-built string to Width().Render().
	spacerW := contentW - lipgloss.Width(left) - lipgloss.Width(right)
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := lipgloss.NewStyle().Width(spacerW).Render("")

	line := lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
	return s.HeaderBox.Render(line)
}

// ── footer: contextual, key-badge styled (plan §6.5) ──────────────────────────

// focusState identifies which panel holds input focus in the v2 layout.
type focusState int

const (
	focusSubmenu focusState = iota
	focusContent
)

// overlayKind identifies a modal overlay that takes over input handling.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayTheme
	overlayQuitConfirm
)

// hintBadge is a single "[key] description" footer entry.
type hintBadge struct {
	key  string
	desc string
}

// hintsFor returns the baseline hint set for a focus/overlay combination
// (plan §6.5 "Baseline hint sets").
func hintsFor(focus focusState, overlay overlayKind) []hintBadge {
	switch overlay {
	case overlayQuitConfirm:
		return []hintBadge{
			{"Enter/y", "Confirmar"},
			{"Esc", "Cancelar"},
		}
	case overlayTheme:
		return []hintBadge{
			{"↑↓/jk", "navegar"},
			{"enter", "confirmar"},
			{"esc", "cancelar"},
			{"q", "sair"},
		}
	}

	if focus == focusContent {
		return []hintBadge{
			{"↑/↓", "Navegar"},
			{"Enter", "Executar"},
			{"Esc/Tab", "← Seções"},
		}
	}
	return []hintBadge{
		{"↑/↓", "Seção"},
		{"Enter/Tab", "Itens"},
		{"t", "Tema"},
		{"q", "Sair"},
	}
}

// renderChromeFooter renders the contextual footer inside a rounded border box,
// mirroring the header layout. Shows only the key-hint badges for the current
// focus/overlay state — no section label on the right.
func renderChromeFooter(width int, focus focusState, overlay overlayKind, s TUIStyles) string {
	hints := hintsFor(focus, overlay)
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, s.KeyBadge.Render(h.key)+" "+s.Footer.Render(h.desc))
	}
	linha := strings.Join(parts, s.Footer.Render("   "))

	// Width(n) includes padding; only border adds 2 on top → visual = Width + 2.
	// Width(w-3) → visual = w-1, keeping one character of margin to prevent the
	// terminal from wrapping the closing ╯ when the footer is the last rendered line.
	contentW := width - 3
	if contentW < 10 {
		contentW = 10
	}
	return s.FooterBox.Width(contentW).Render(linha)
}

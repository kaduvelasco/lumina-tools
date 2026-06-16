package tui

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kaduvelasco/lumina-tools/internal/config"
)

// ── shared message types ──────────────────────────────────────────────────────

type notImplementedMsg struct{}
type actionDoneMsg struct{ err error }

type msgKind int

const (
	msgNone    msgKind = iota
	msgSuccess         // green
	msgWarning         // yellow
	msgError           // red
)

// ── TUI entry points ──────────────────────────────────────────────────────────

// Run loads config and starts the Bubble Tea TUI at the main menu.
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAtV2(ctx, stdin, stdout, sectionHome, 0)
}

// RunAtSystemPostInstall starts the TUI at Gerenciamento Linux > Pós Instalação.
func RunAtSystemPostInstall(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAtV2(ctx, stdin, stdout, sectionLinux, 0)
}

// RunAtStackConfig starts the TUI at Ambiente de Desenvolvimento > Criar Stack PHP.
func RunAtStackConfig(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAtV2(ctx, stdin, stdout, sectionDev, 2)
}

// stderr is intentionally ignored: tea.NewProgram renders exclusively to stdout (alt-screen).
func runAtV2(ctx context.Context, stdin io.Reader, stdout io.Writer, section, cursor int) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregar config: %w", err)
	}
	if err := ensureSystemInfo(ctx, stdin, stdout, cfg); err != nil {
		return fmt.Errorf("configurar sistema: %w", err)
	}
	m := NewV2(ctx, cfg)
	m.section = section
	m.cursor = cursor
	p := tea.NewProgram(
		m,
		tea.WithContext(ctx),
		tea.WithInput(stdin),
		tea.WithOutput(stdout),
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	fmt.Fprint(stdout, "\033[3J\033[2J\033[H")
	return err
}

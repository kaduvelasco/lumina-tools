package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectItem represents a toggleable item in a multi-select list.
type SelectItem struct {
	Label    string
	ID       string
	Selected bool
}

// ── key bindings ──────────────────────────────────────────────────────────────

type msKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Done   key.Binding
	Quit   key.Binding
}

var msKeys = msKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k")),
	Down:   key.NewBinding(key.WithKeys("down", "j")),
	Toggle: key.NewBinding(key.WithKeys(" ")),
	Done:   key.NewBinding(key.WithKeys("enter")),
	Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c")),
}

// ── model ─────────────────────────────────────────────────────────────────────

type msModel struct {
	items     []SelectItem
	cursor    int
	confirmed bool
	aborted   bool
}

var (
	msChecked   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))
	msUnchecked = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	msActive    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF99FF")).Bold(true)
	msHint      = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	msCount     = lipgloss.NewStyle().Foreground(lipgloss.Color("#9966FF"))
)

func (m msModel) Init() tea.Cmd { return nil }

func (m msModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, msKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.items) - 1
			}
		case key.Matches(msg, msKeys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case key.Matches(msg, msKeys.Toggle):
			items := make([]SelectItem, len(m.items))
			copy(items, m.items)
			items[m.cursor].Selected = !items[m.cursor].Selected
			m.items = items
		case key.Matches(msg, msKeys.Done):
			m.confirmed = true
			return m, tea.Quit
		case key.Matches(msg, msKeys.Quit):
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m msModel) View() string {
	var sb strings.Builder
	sb.WriteString(msHint.Render("  Setas: navegar  |  Espaco: selecionar  |  Enter: confirmar  |  q: cancelar") + "\n\n")

	for i, item := range m.items {
		checkbox := msUnchecked.Render("[ ]")
		if item.Selected {
			checkbox = msChecked.Render("[x]")
		}
		label := item.Label
		if i == m.cursor {
			label = msActive.Render(item.Label)
		}
		sb.WriteString(fmt.Sprintf("  %s  %s\n", checkbox, label))
	}

	count := 0
	for _, it := range m.items {
		if it.Selected {
			count++
		}
	}
	sb.WriteString("\n")
	sb.WriteString(msCount.Render(fmt.Sprintf("  %d selecionado(s)", count)) + "\n")
	return sb.String()
}

// ── single-select ─────────────────────────────────────────────────────────────

// ssModel is the Bubble Tea model for the single-item picker.
// The user navigates with arrows and confirms with Enter — no toggle needed.
type ssModel struct {
	items   []SelectItem
	cursor  int
	chosen  int  // -1 = nothing chosen yet
	aborted bool
}

var ssKeys = msKeyMap{
	Up:   key.NewBinding(key.WithKeys("up", "k")),
	Down: key.NewBinding(key.WithKeys("down", "j")),
	Done: key.NewBinding(key.WithKeys("enter")),
	Quit: key.NewBinding(key.WithKeys("q", "esc", "ctrl+c")),
}

func (m ssModel) Init() tea.Cmd { return nil }

func (m ssModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, ssKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.items) - 1
			}
		case key.Matches(msg, ssKeys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case key.Matches(msg, ssKeys.Done):
			m.chosen = m.cursor
			return m, tea.Quit
		case key.Matches(msg, ssKeys.Quit):
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ssModel) View() string {
	var sb strings.Builder
	sb.WriteString(msHint.Render("  ↑↓/jk navegar  |  Enter selecionar  |  q/esc cancelar") + "\n\n")
	for i, item := range m.items {
		if i == m.cursor {
			sb.WriteString(msActive.Render("  › "+item.Label) + "\n")
		} else {
			sb.WriteString(msUnchecked.Render("    "+item.Label) + "\n")
		}
	}
	return sb.String()
}

// RunSingleSelect shows a keyboard-driven single-item picker.
// Returns the index of the chosen item, whether the user confirmed, and any error.
// Returns -1, false when cancelled with q/Esc.
func RunSingleSelect(ctx context.Context, stdin io.Reader, stdout io.Writer, items []SelectItem) (int, bool, error) {
	if len(items) == 0 {
		return -1, false, nil
	}
	m := ssModel{items: items, chosen: -1}
	opts := []tea.ProgramOption{
		tea.WithOutput(stdout),
		tea.WithContext(ctx),
	}
	if stdin != nil {
		opts = append(opts, tea.WithInput(stdin))
	}
	p := tea.NewProgram(m, opts...)
	final, err := p.Run()
	if err != nil {
		return -1, false, err
	}
	fm, ok := final.(ssModel)
	if !ok {
		return -1, false, fmt.Errorf("modelo inesperado retornado pelo programa")
	}
	if fm.aborted {
		return -1, false, nil
	}
	return fm.chosen, true, nil
}

// ── RunMultiSelect ────────────────────────────────────────────────────────────

// RunMultiSelect shows an interactive multi-select list.
// Returns the updated items slice and whether the user confirmed (vs. aborted with q/Esc).
func RunMultiSelect(ctx context.Context, stdin io.Reader, stdout io.Writer, items []SelectItem) ([]SelectItem, bool, error) {
	m := msModel{items: items}
	opts := []tea.ProgramOption{
		tea.WithOutput(stdout),
		tea.WithContext(ctx),
	}
	if stdin != nil {
		opts = append(opts, tea.WithInput(stdin))
	}

	p := tea.NewProgram(m, opts...)
	final, err := p.Run()
	if err != nil {
		return items, false, err
	}
	fm, ok := final.(msModel)
	if !ok {
		return items, false, fmt.Errorf("modelo inesperado retornado pelo programa")
	}
	if fm.aborted {
		return items, false, nil
	}
	return fm.items, fm.confirmed, nil
}

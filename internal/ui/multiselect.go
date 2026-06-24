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

const msPageSize = 10

type msModel struct {
	items     []SelectItem
	cursor    int
	confirmed bool
	aborted   bool
	width     int
}

func (m msModel) Init() tea.Cmd { return nil }

func (m msModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
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
	t := LoadTheme()

	w := m.width
	if w < 20 {
		w = 80
	}
	if w > 120 {
		w = 120
	}
	// Width(w-2) + border(2) = w (same formula as printPanel).
	contentW := w - 2
	if contentW < 10 {
		contentW = 10
	}

	themedPanel := func(color lipgloss.Color, content string) string {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			Padding(0, 1).
			Width(contentW).
			Render(content)
	}

	hintStyle := lipgloss.NewStyle().Foreground(t.Muted)
	checkedStyle := lipgloss.NewStyle().Foreground(t.Success)
	unchecked := lipgloss.NewStyle().Foreground(t.Muted)
	activeStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	countStyle := lipgloss.NewStyle().Foreground(t.Accent)

	// Instructions panel.
	hint := hintStyle.Render("Setas: navegar  |  Espaço: selecionar  |  Enter: confirmar  |  q: cancelar")

	// Pagination: show msPageSize items at a time, page derived from cursor.
	totalPages := (len(m.items) + msPageSize - 1) / msPageSize
	page := m.cursor / msPageSize
	start := page * msPageSize
	end := start + msPageSize
	if end > len(m.items) {
		end = len(m.items)
	}

	// Items list panel (current page only).
	var listSb strings.Builder
	for i := start; i < end; i++ {
		item := m.items[i]
		checkbox := unchecked.Render("[ ]")
		if item.Selected {
			checkbox = checkedStyle.Render("[x]")
		}
		label := item.Label
		if i == m.cursor {
			label = activeStyle.Render(item.Label)
		}
		listSb.WriteString(fmt.Sprintf("  %s  %s\n", checkbox, label))
	}

	// Count + page indicator panel.
	count := 0
	for _, it := range m.items {
		if it.Selected {
			count++
		}
	}
	countText := fmt.Sprintf("%d selecionado(s)", count)
	if totalPages > 1 {
		pageStyle := lipgloss.NewStyle().Foreground(t.Muted)
		countText = countStyle.Render(countText) + "   " + pageStyle.Render(fmt.Sprintf("Página %d de %d", page+1, totalPages))
	} else {
		countText = countStyle.Render(countText)
	}

	var sb strings.Builder
	sb.WriteString(themedPanel(t.Muted, hint))
	sb.WriteString("\n")
	sb.WriteString(themedPanel(t.Primary, strings.TrimRight(listSb.String(), "\n")))
	sb.WriteString("\n")
	sb.WriteString(themedPanel(t.Primary, countText))
	return sb.String()
}

// ── single-select ─────────────────────────────────────────────────────────────

// ssModel is the Bubble Tea model for the single-item picker.
// The user navigates with arrows and confirms with Enter — no toggle needed.
type ssModel struct {
	items   []SelectItem
	cursor  int
	chosen  int // -1 = nothing chosen yet
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
	t := LoadTheme()
	hintStyle := lipgloss.NewStyle().Foreground(t.Muted)
	activeStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(t.Muted)

	var sb strings.Builder
	sb.WriteString(hintStyle.Render("  ↑↓/jk navegar  |  Enter selecionar  |  q/esc cancelar") + "\n\n")

	// Pagination: show msPageSize items at a time, page derived from cursor.
	totalPages := (len(m.items) + msPageSize - 1) / msPageSize
	page := m.cursor / msPageSize
	start := page * msPageSize
	end := start + msPageSize
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		if i == m.cursor {
			sb.WriteString(activeStyle.Render("  › "+item.Label) + "\n")
		} else {
			sb.WriteString(inactiveStyle.Render("    "+item.Label) + "\n")
		}
	}

	if totalPages > 1 {
		sb.WriteString("\n" + hintStyle.Render(fmt.Sprintf("  Página %d de %d", page+1, totalPages)) + "\n")
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
	m := msModel{items: items, width: 80}
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

// ── gate ──────────────────────────────────────────────────────────────────────

// gateModel is a minimal single-keypress confirmation: Enter confirms,
// q/Esc/Ctrl+C cancels. Shares msKeys and the same capped-width panel
// treatment as msModel/ssModel.
type gateModel struct {
	message   string
	width     int
	confirmed bool
}

func (m gateModel) Init() tea.Cmd { return nil }

func (m gateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, msKeys.Done):
			m.confirmed = true
			return m, tea.Quit
		case key.Matches(msg, msKeys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m gateModel) View() string {
	t := LoadTheme()

	w := m.width
	if w < 20 {
		w = 80
	}
	if w > 120 {
		w = 120
	}
	contentW := w - 2
	if contentW < 10 {
		contentW = 10
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Muted).
		Padding(0, 1).
		Width(contentW).
		Render(m.message)
}

// RunGate shows a single-keypress confirmation panel with the given message.
// Returns true when the user presses Enter, false when cancelled with q/Esc.
func RunGate(ctx context.Context, stdin io.Reader, stdout io.Writer, message string) (bool, error) {
	m := gateModel{message: message, width: 80}
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
		return false, err
	}
	fm, ok := final.(gateModel)
	if !ok {
		return false, fmt.Errorf("modelo inesperado retornado pelo programa")
	}
	return fm.confirmed, nil
}

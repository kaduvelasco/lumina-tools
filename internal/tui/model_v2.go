package tui

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	devflutter "github.com/kaduvelasco/lumina-tools/internal/dev/flutter"
	devgolang "github.com/kaduvelasco/lumina-tools/internal/dev/golang"
	"github.com/kaduvelasco/lumina-tools/internal/dev/ide"
	"github.com/kaduvelasco/lumina-tools/internal/dev/llm"
	"github.com/kaduvelasco/lumina-tools/internal/dev/mcp"
	"github.com/kaduvelasco/lumina-tools/internal/dev/prereqs"
	devterminal "github.com/kaduvelasco/lumina-tools/internal/dev/terminal"
	"github.com/kaduvelasco/lumina-tools/internal/dev/upgrade"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	managerai "github.com/kaduvelasco/lumina-tools/internal/manager/ai"
	managerdb "github.com/kaduvelasco/lumina-tools/internal/manager/db"
	managergitignore "github.com/kaduvelasco/lumina-tools/internal/manager/gitignore"
	managerrepo "github.com/kaduvelasco/lumina-tools/internal/manager/repo"
	"github.com/kaduvelasco/lumina-tools/internal/selfupdate"
	"github.com/kaduvelasco/lumina-tools/internal/stack"
	stackconfig "github.com/kaduvelasco/lumina-tools/internal/stack/config"
	"github.com/kaduvelasco/lumina-tools/internal/system/apps"
	"github.com/kaduvelasco/lumina-tools/internal/system/fonts"
	"github.com/kaduvelasco/lumina-tools/internal/system/gnome"
	"github.com/kaduvelasco/lumina-tools/internal/system/linuxtoys"
	"github.com/kaduvelasco/lumina-tools/internal/system/megasync"
	"github.com/kaduvelasco/lumina-tools/internal/system/postinstall"
	"github.com/kaduvelasco/lumina-tools/internal/system/templates"
	"github.com/kaduvelasco/lumina-tools/internal/system/update"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// openQuitConfirmMsg is sent by actQuit to open the quit-confirmation overlay.
type openQuitConfirmMsg struct{}

// ModelV2 is the v2 Bubble Tea application model.
type ModelV2 struct {
	ctx    context.Context
	cfg    *config.Config
	width  int
	height int

	theme         Theme
	styles        TUIStyles
	online        bool
	distroDisplay string

	section int
	cursor  int
	focus   focusState

	msgKind msgKind
	msg     string

	// overlay state
	themeOpen       bool
	themeCursor     int
	quitConfirmOpen bool
}

// NewV2 returns the initial v2 model with focus on the section sidebar.
func NewV2(ctx context.Context, cfg *config.Config) ModelV2 {
	var t Theme
	if cfg.Theme != "" {
		t = themeByName(cfg.Theme)
	} else {
		t = detectDefaultTheme()
	}
	return ModelV2{
		ctx:           ctx,
		cfg:           cfg,
		width:         80,
		height:        24,
		theme:         t,
		styles:        buildStyles(t),
		online:        true,
		focus:         focusSubmenu,
		distroDisplay: distroDisplayName(cfg.Distro),
	}
}

// Init fires the first connectivity check and schedules the periodic tick.
func (m ModelV2) Init() tea.Cmd {
	return tea.Batch(checkConnectivity(m.ctx), tickConnectivity())
}

func (m ModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case connectivityMsg:
		m.online = msg.online
		return m, nil

	case connectivityTickMsg:
		return m, tea.Batch(checkConnectivity(m.ctx), tickConnectivity())

	case openQuitConfirmMsg:
		m.quitConfirmOpen = true
		return m, nil

	case notImplementedMsg:
		m.msgKind = msgWarning
		m.msg = "Em desenvolvimento..."
		return m, nil

	case actionDoneMsg:
		if errors.Is(msg.err, selfupdate.ErrUninstalled) {
			return m, tea.Quit
		}
		// Reload config from disk — actions like Workspace and Compose may have updated it.
		if cfg, err := config.Load(); err == nil {
			m.cfg = cfg
		}
		if msg.err != nil {
			m.msgKind = msgError
			m.msg = msg.err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m ModelV2) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Overlays intercept all key navigation first.
	if m.themeOpen {
		return m.updateThemeOverlay(msg)
	}
	if m.quitConfirmOpen {
		return m.updateQuitOverlay(msg)
	}

	// Clear any previous notification on every keypress.
	m.msg = ""
	m.msgKind = msgNone

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch msg.String() {
	case "q":
		m.quitConfirmOpen = true
		return m, nil

	case "t":
		m.themeOpen = true
		for i, t := range availableThemes {
			if t.Name == m.theme.Name {
				m.themeCursor = i
				break
			}
		}
		m.styles = buildStyles(availableThemes[m.themeCursor])
		return m, nil

	case "tab", "shift+tab":
		if m.focus == focusSubmenu {
			m.focus = focusContent
		} else {
			m.focus = focusSubmenu
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Back):
		if m.focus == focusContent {
			m.focus = focusSubmenu
		}
		return m, nil

	case key.Matches(msg, keys.Select):
		if m.focus == focusContent {
			entry := m.visibleItems(m.section)[m.cursor]
			if entry.pending {
				m.msgKind = msgWarning
				m.msg = "Em breve"
				return m, nil
			}
			if entry.action != actNone {
				return m, m.runActionV2(entry.action)
			}
		} else {
			// Enter on section sidebar moves focus to the items list.
			m.focus = focusContent
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.focus == focusSubmenu {
			m.moveSection(-1)
		} else {
			m.moveCursor(-1)
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.focus == focusSubmenu {
			m.moveSection(1)
		} else {
			m.moveCursor(1)
		}
		return m, nil
	}
	return m, nil
}

// ── overlay handlers ──────────────────────────────────────────────────────────

func (m ModelV2) updateThemeOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "ctrl+c":
		return m, tea.Quit

	case msg.String() == "q":
		return m, tea.Quit

	case key.Matches(msg, keys.Back):
		m.themeOpen = false
		m.styles = buildStyles(m.theme)
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.themeCursor > 0 {
			m.themeCursor--
		} else {
			m.themeCursor = len(availableThemes) - 1
		}
		m.styles = buildStyles(availableThemes[m.themeCursor])
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.themeCursor < len(availableThemes)-1 {
			m.themeCursor++
		} else {
			m.themeCursor = 0
		}
		m.styles = buildStyles(availableThemes[m.themeCursor])
		return m, nil

	case key.Matches(msg, keys.Select):
		m.themeOpen = false
		m.theme = availableThemes[m.themeCursor]
		m.styles = buildStyles(m.theme)
		m.cfg.Theme = m.theme.Name
		if err := config.Save(m.cfg); err != nil {
			m.msgKind = msgError
			m.msg = "Falha ao salvar tema: " + err.Error()
		}
		return m, nil
	}
	return m, nil
}

func (m ModelV2) updateQuitOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "enter", " ", "y", "s", "Y", "S":
		return m, tea.Quit
	case "esc", "q", "n", "N":
		m.quitConfirmOpen = false
		return m, nil
	}
	return m, nil
}

// ── action dispatch ───────────────────────────────────────────────────────────

func (m ModelV2) runActionV2(a actionID) tea.Cmd {
	done := func(err error) tea.Msg { return actionDoneMsg{err: err} }

	exec := func(fn func(context.Context, *executor.Executor, io.Writer) error) tea.Cmd {
		return tea.Exec(newFuncCmd(m.ctx, fn), done)
	}

	execInteractive := func(fn func(context.Context, *executor.Executor, io.Reader, io.Writer) error) tea.Cmd {
		return tea.Exec(newInteractiveFuncCmd(m.ctx, fn), done)
	}

	switch a {
	case actQuit:
		return func() tea.Msg { return openQuitConfirmMsg{} }

	case actSystemUpdate:
		return exec(update.Run)

	case actSystemPostMint:
		return exec(postinstall.Mint)
	case actSystemPostZorin:
		return exec(postinstall.Zorin)
	case actSystemPostUbuntu:
		return execInteractive(postinstall.Ubuntu)
	case actSystemPostFedora:
		return exec(postinstall.Fedora)

	case actSystemFonts:
		return execInteractive(fonts.Select)

	case actSystemTemplates:
		return execInteractive(templates.Select)

	case actLinuxToys:
		return exec(linuxtoys.Install)
	case actMegaSync:
		return exec(megasync.Install)

	case actAppsInstall:
		return execInteractive(apps.SelectInstall)
	case actAppsUninstall:
		return execInteractive(apps.SelectUninstall)
	case actAppsWebApps:
		return exec(apps.ShowWebApps)

	case actGnomePrereqs:
		return exec(gnome.InstallPrereqs)
	case actGnomeExtensions:
		return exec(gnome.ShowExtensions)
	case actGnomeThemes:
		return execInteractive(gnome.ManageThemes)
	case actCinnamonPrereqs:
		return exec(gnome.InstallCinnamonPrereqs)
	case actCinnamonThemes:
		return execInteractive(gnome.ManageCinnamonThemes)
	case actGnomeIcons:
		return execInteractive(gnome.ManageIcons)
	case actGnomeCursors:
		return execInteractive(gnome.ManageCursors)
	case actGnomeFlatpak:
		return execInteractive(gnome.ApplyFlatpakTheme)

	case actPrereqs:
		return execInteractive(prereqs.Select)
	case actGoManage:
		return execInteractive(devgolang.Manage)
	case actFlutterManage:
		return execInteractive(devflutter.Manage)
	case actLLMManage:
		return execInteractive(llm.Select)
	case actIDEManage:
		return execInteractive(ide.Select)
	case actTermManage:
		return execInteractive(devterminal.Select)
	case actMCPManage:
		return execInteractive(mcp.Select)
	case actDevUpgrade:
		return exec(upgrade.Update)

	case actStackWorkspace:
		return exec(stackconfig.Workspace)
	case actStackCompose:
		return execInteractive(stackconfig.Compose)

	case actStackStart:
		composeDir := m.cfg.DockerComposeDir
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Start(ctx, exe, w, composeDir)
		})
	case actStackStop:
		composeDir := m.cfg.DockerComposeDir
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Stop(ctx, exe, w, composeDir)
		})
	case actStackRestart:
		composeDir := m.cfg.DockerComposeDir
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Restart(ctx, exe, w, composeDir)
		})
	case actStackLogs:
		composeDir := m.cfg.DockerComposeDir
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Logs(ctx, exe, w, composeDir)
		})
	case actStackStats:
		return exec(stack.Stats)
	case actStackDB:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.DBInfo(ctx, exe, w)
		})
	case actStackFixPerms:
		workspacePath := m.cfg.WorkspacePath
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.FixPerms(ctx, exe, w, workspacePath)
		})

	case actAIContext:
		return execInteractive(managerai.GenerateContext)
	case actAIContextRemove:
		return execInteractive(managerai.ClearContext)
	case actGitignore:
		return exec(managergitignore.Generate)

	case actDBBackup:
		return exec(managerdb.Backup)
	case actDBRestore:
		return exec(managerdb.Restore)
	case actDBRemove:
		return exec(managerdb.Remove)
	case actDBOptimize:
		return exec(managerdb.Optimize)
	case actDBMoodle:
		return exec(managerdb.OptimizeMoodle)

	case actRepoGlobal:
		return exec(managerrepo.ConfigureGlobal)
	case actRepoInit:
		return exec(managerrepo.Init)
	case actRepoClone:
		return exec(managerrepo.Clone)
	case actRepoIdent:
		return exec(managerrepo.ApplyIdent)
	case actRepoConduct:
		return exec(managerrepo.CreateConduct)

	case actLuminaConfig:
		return execInteractive(selfupdate.Configure)
	case actLuminaUpdate:
		return exec(selfupdate.Run)
	case actLuminaUninstall:
		return exec(selfupdate.Uninstall)
	case actLuminaHelp:
		return execInteractive(selfupdate.ShowHelp)

	default:
		return func() tea.Msg { return notImplementedMsg{} }
	}
}

// ── movement helpers ──────────────────────────────────────────────────────────

// visibleItems returns the items for the given section, applying filters:
//   - gnomeOnly items are hidden when cfg.DE != "gnome"
//   - cinnamonOnly items are hidden when cfg.DE != "cinnamon"
//   - distro-tagged items are hidden when cfg.Distro is set and does not match
func (m ModelV2) visibleItems(sec int) []submenuEntry {
	items := sections[sec].items
	result := make([]submenuEntry, 0, len(items))
	for _, item := range items {
		if item.gnomeOnly && m.cfg.DE != "gnome" {
			continue
		}
		if item.cinnamonOnly && m.cfg.DE != "cinnamon" {
			continue
		}
		if item.distro != "" && m.cfg.Distro != "" && item.distro != m.cfg.Distro {
			continue
		}
		result = append(result, item)
	}
	return result
}

// moveSection changes the active section and resets the item cursor.
func (m *ModelV2) moveSection(delta int) {
	n := len(sections)
	m.section = (m.section + delta + n) % n
	m.cursor = 0
}

// moveCursor shifts the item cursor in the right panel, wrapping at both ends.
func (m *ModelV2) moveCursor(delta int) {
	n := len(m.visibleItems(m.section))
	if n == 0 {
		return
	}
	m.cursor = (m.cursor + delta + n) % n
}

// ── layout helpers ────────────────────────────────────────────────────────────

// sectionPanelWidth returns the left-panel content width: ~30% of total, clamped [33, 40].
// Minimum 33 ensures "Ambiente de Desenvolvimento" (longest label with indent) fits.
func sectionPanelWidth(total int) int {
	w := total * 30 / 100
	if w < 33 {
		w = 33
	}
	if w > 40 {
		w = 40
	}
	return w
}

// bodyHeight returns the content height for the body panels (inside their borders).
func (m ModelV2) bodyHeight() int {
	// header(3) + \n + panels_visual(bodyH+2) + \n + footer(3) = bodyH+8 lines.
	// For bodyH+8 = m.height: bodyH = m.height - 8.
	h := m.height - 8
	if h < 5 {
		h = 5
	}
	return h
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m ModelV2) View() string {
	if m.themeOpen {
		return m.renderThemeOverlay()
	}
	if m.quitConfirmOpen {
		return m.renderQuitOverlay()
	}

	s := m.styles

	var sb strings.Builder
	sb.WriteString(renderChromeHeader(m.width, m.online, m.distroDisplay, s))
	sb.WriteString("\n")
	sb.WriteString(m.renderBody())
	if m.msg != "" {
		sb.WriteString("\n")
		sb.WriteString(m.renderStatusLine())
	}
	sb.WriteString("\n")
	sb.WriteString(renderChromeFooter(m.width, m.focus, overlayNone, s))
	return sb.String()
}

func (m ModelV2) renderThemeOverlay() string {
	s := m.styles

	var sb strings.Builder
	sb.WriteString(renderChromeHeader(m.width, m.online, m.distroDisplay, s))
	sb.WriteString("\n\n")

	for i, t := range availableThemes {
		label := t.Name
		if t.Name == m.theme.Name {
			label += "  (atual)"
		}
		if i == m.themeCursor {
			sb.WriteString(s.ActiveBar.Render("│ ") + s.ActiveText.Render(label))
		} else {
			sb.WriteString("   " + s.Inactive.Render(label))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(renderChromeFooter(m.width, m.focus, overlayTheme, s))
	return sb.String()
}

func (m ModelV2) renderQuitOverlay() string {
	s := m.styles

	content := s.Warning.Render("Deseja encerrar o Lumina Tools?") + "\n\n" +
		s.Inactive.Render("Pressione ") +
		s.ActiveText.Render("[Enter]") +
		s.Inactive.Render(" para confirmar ou ") +
		s.ActiveText.Render("[Esc]") +
		s.Inactive.Render(" para cancelar.")

	// Width(n) includes padding; only border adds 2 on top → visual = Width + 2.
	// Width(w-2) + Border(2) = w (full terminal width).
	contentW := m.width - 2
	if contentW < 20 {
		contentW = 20
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Warning).
		Padding(0, 2).
		Width(contentW)

	var sb strings.Builder
	sb.WriteString(renderChromeHeader(m.width, m.online, m.distroDisplay, s))
	sb.WriteString("\n\n")
	sb.WriteString(box.Render(content))
	sb.WriteString("\n")
	sb.WriteString(renderChromeFooter(m.width, m.focus, overlayQuitConfirm, s))
	return sb.String()
}

func (m ModelV2) renderStatusLine() string {
	var style lipgloss.Style
	switch m.msgKind {
	case msgSuccess:
		style = m.styles.Success
	case msgError:
		style = m.styles.Error
	default:
		style = m.styles.Warning
	}
	return style.Render("  " + m.msg)
}

// renderBody lays out the two-panel body: section sidebar (left) + item list (right).
// Each panel has a rounded border whose color changes to Primary when focused.
// Width(n) in lipgloss v1 includes padding; border adds 2. Each panel visual = n+2.
// Layout: left(lw+2) + sep(1) + right(rw+2) = m.width → rw = m.width - lw - 5.
func (m ModelV2) renderBody() string {
	bodyH := m.bodyHeight()
	lw := sectionPanelWidth(m.width)
	rw := m.width - lw - 5
	if rw < 20 {
		rw = 20
	}

	leftColor := m.theme.Muted
	if m.focus == focusSubmenu {
		leftColor = m.theme.Primary
	}
	rightColor := m.theme.Muted
	if m.focus == focusContent {
		rightColor = m.theme.Primary
	}

	leftPanel := lipgloss.NewStyle().
		Width(lw).
		Height(bodyH).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftColor).
		Render(m.renderSectionList(lw))

	// When "Sobre" (Home section, first item) is highlighted, show the About
	// panel instead of the item list — preview-on-selection UX.
	var rightContent string
	if m.section == 0 && m.cursor == 0 && len(sections[0].items) > 0 {
		rightContent = m.renderSobrePanel()
	} else {
		rightContent = m.renderItemsList(bodyH)
	}

	rightPanel := lipgloss.NewStyle().
		Width(rw).
		Height(bodyH).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightColor).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)
}

// renderSectionList renders the left-panel list of all sections.
// panelW is the content width of the panel (used to size the selection bar).
func (m ModelV2) renderSectionList(panelW int) string {
	s := m.styles
	lines := make([]string, len(sections))
	for i, sec := range sections {
		if i == m.section {
			lines[i] = lipgloss.NewStyle().
				Foreground(m.theme.Accent).
				Bold(true).
				Border(lipgloss.ThickBorder(), false, false, false, true).
				BorderForeground(m.theme.Primary).
				PaddingLeft(1).
				Width(panelW - 2).
				Render("› " + sec.label)
		} else {
			lines[i] = "    " + s.Inactive.Render(sec.label)
		}
	}
	return strings.Join(lines, "\n")
}

// renderSobrePanel renders the About info in the right panel.
// Shown when cursor is on the "Sobre" item (section 0, cursor 0).
func (m ModelV2) renderSobrePanel() string {
	s := m.styles
	lbl := func(l string) string { return s.Breadcrumb.Bold(true).Render(l) }
	val := func(v string) string { return s.Inactive.Render(v) }

	var b strings.Builder
	b.WriteString(s.ActiveText.Render("Lumina Tools"))
	b.WriteString("\n\n")
	b.WriteString(lbl("Descrição") + "\n")
	b.WriteString(val("Binário Go Unificado para linux com TUI") + "\n")
	b.WriteString(val("interativa e CLI completa") + "\n")
	b.WriteString("\n")
	b.WriteString(lbl("Versão") + "\n")
	b.WriteString(val(version.Version) + "   " + val("lumina") + "\n")
	b.WriteString("\n")
	b.WriteString(lbl("Autor") + "\n")
	b.WriteString(val("Kadu Velasco") + "\n")
	b.WriteString(val("kadu.velasco@gmail.com") + "\n")
	b.WriteString(val("github.com/kaduvelasco/lumina-tools"))
	return b.String()
}

// renderItemsList renders the right-panel flat list of submenu items with
// a scrolling viewport to handle sections with many entries (e.g. DevManager).
func (m ModelV2) renderItemsList(height int) string {
	s := m.styles
	items := m.visibleItems(m.section)
	if len(items) == 0 {
		return ""
	}

	const linesPerItem = 3 // title + cmd/desc + blank separator
	visible := height / linesPerItem
	if visible < 1 {
		visible = 1
	}

	// Scroll to keep cursor visible, centered when possible.
	start := m.cursor - visible/2
	maxStart := len(items) - visible
	if maxStart < 0 {
		maxStart = 0
	}
	if start > maxStart {
		start = maxStart
	}
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > len(items) {
		end = len(items)
	}

	var lines []string
	for i := start; i < end; i++ {
		entry := items[i]

		// Second line: cmd takes priority; desc is the fallback.
		secondLine := entry.cmd
		if secondLine == "" {
			secondLine = entry.desc
		}

		if entry.pending {
			// Pending items appear disabled — muted color, "(em breve)" suffix.
			lines = append(lines,
				"  "+s.Footer.Render(entry.title)+" "+s.Footer.Render("(em breve)"),
				"  "+s.Footer.Render(secondLine),
				"",
			)
		} else if i == m.cursor {
			// Thick left bar wrapping both lines, matching the reference dashboard.
			block := lipgloss.JoinVertical(lipgloss.Left,
				s.ActiveText.Render(entry.title),
				s.Footer.Render(secondLine),
			)
			block = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder(), false, false, false, true).
				BorderForeground(m.theme.Primary).
				PaddingLeft(1).
				Render(block)
			lines = append(lines, block, "")
		} else {
			lines = append(lines,
				"  "+s.Inactive.Render(entry.title),
				"  "+s.Footer.Render(secondLine),
				"",
			)
		}
	}

	// Trim the trailing blank separator line.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/dev/depends"
	"github.com/kaduvelasco/lumina-tools/internal/dev/ide"
	"github.com/kaduvelasco/lumina-tools/internal/dev/llm"
	"github.com/kaduvelasco/lumina-tools/internal/dev/mcp"
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
	"github.com/kaduvelasco/lumina-tools/internal/system/postinstall"
	"github.com/kaduvelasco/lumina-tools/internal/system/templates"
	"github.com/kaduvelasco/lumina-tools/internal/system/ulauncher"
	"github.com/kaduvelasco/lumina-tools/internal/system/update"
)

// chromeHeight is the number of lines occupied by fixed UI elements (header,
// breadcrumb, dividers, footer) that surround the list.
const chromeHeight = 12

// Run loads config and starts the Bubble Tea TUI program at the main menu.
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAt(ctx, stdin, stdout, stderr, []navLevel{{menu: menuMain, cursor: 0}})
}

// RunAtSystemPostInstall starts the TUI positioned at Pós Instalação.
func RunAtSystemPostInstall(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAt(ctx, stdin, stdout, stderr, []navLevel{
		{menu: menuMain, cursor: 0},
		{menu: menuSystem, cursor: 0},
		{menu: menuSystemPostInstall, cursor: 0},
	})
}

// RunAtStackConfig starts the TUI positioned at DevStuff > Criar Stack.
func RunAtStackConfig(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	return runAt(ctx, stdin, stdout, stderr, []navLevel{
		{menu: menuMain, cursor: 1},
		{menu: menuDev, cursor: 0},
		{menu: menuDevStack, cursor: 0},
	})
}

// stderr is intentionally ignored: tea.NewProgram renders exclusively to stdout (alt-screen).
func runAt(ctx context.Context, stdin io.Reader, stdout, _ io.Writer, nav []navLevel) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregar config: %w", err)
	}
	m := New(ctx, cfg)
	m.nav = nav
	last := nav[len(nav)-1]
	_ = m.list.SetItems(toListItems(itemsFor(last.menu)))
	m.list.Select(last.cursor)
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

// ── messages ──────────────────────────────────────────────────────────────────

type notImplementedMsg struct{}
type actionDoneMsg struct{ err error }

// ── navigation stack ──────────────────────────────────────────────────────────

type navLevel struct {
	menu   menuID
	cursor int
}

// ── message kind for styled notifications ────────────────────────────────────

type msgKind int

const (
	msgNone    msgKind = iota
	msgSuccess         // green
	msgWarning         // yellow
	msgError           // red
)

// ── model ─────────────────────────────────────────────────────────────────────

// Model is the Bubble Tea application model.
type Model struct {
	ctx     context.Context
	cfg     *config.Config
	nav     []navLevel
	list    list.Model
	width   int
	height  int
	msgKind msgKind
	msg     string

	// theme state
	theme       Theme
	styles      TUIStyles
	themeOpen   bool
	themeCursor int
}

// New returns the initial model starting at the main menu.
func New(ctx context.Context, cfg *config.Config) Model {
	var t Theme
	if cfg.Theme != "" {
		t = themeByName(cfg.Theme)
	} else {
		t = detectDefaultTheme()
	}

	l := list.New(toListItems(itemsFor(menuMain)), buildDelegate(t), 80, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return Model{
		ctx:    ctx,
		cfg:    cfg,
		nav:    []navLevel{{menu: menuMain, cursor: 0}},
		list:   l,
		width:  80,
		height: 24,
		theme:  t,
		styles: buildStyles(t),
	}
}

func (m Model) breadcrumb() string {
	parts := make([]string, len(m.nav))
	for i, lvl := range m.nav {
		parts[i] = menuLabels[lvl.menu]
	}
	return strings.Join(parts, "  ›  ")
}

// ── tea.Model interface ───────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listH := msg.Height - chromeHeight
		if listH < 4 {
			listH = 4
		}
		m.list.SetSize(msg.Width, listH)
		return m, nil

	case notImplementedMsg:
		m.msgKind = msgWarning
		m.msg = "Em desenvolvimento..."
		return m, nil

	case actionDoneMsg:
		if errors.Is(msg.err, selfupdate.ErrUninstalled) {
			return m, tea.Quit
		}
		if msg.err != nil {
			m.msgKind = msgError
			m.msg = msg.err.Error()
		} else {
			m.msgKind = msgSuccess
			m.msg = "Concluido com sucesso."
		}
		return m, nil

	case tea.KeyMsg:
		// Theme selector intercepts all key navigation.
		if m.themeOpen {
			return m.updateThemeMode(msg)
		}

		m.msg = ""
		m.msgKind = msgNone

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Theme):
			m.themeOpen = true
			for i, t := range availableThemes {
				if t.Name == m.theme.Name {
					m.themeCursor = i
					break
				}
			}
			m.styles = buildStyles(availableThemes[m.themeCursor])
			return m, nil

		case key.Matches(msg, keys.Back):
			return m.navigateBack()

		case key.Matches(msg, keys.Select):
			selected := m.list.SelectedItem()
			if selected == nil {
				return m, nil
			}
			item := selected.(menuItem)

			if item.action == actBack {
				return m.navigateBack()
			}

			if item.submenu != 0 {
				// Save cursor so we can restore it when navigating back.
				m.nav[len(m.nav)-1].cursor = m.list.Index()
				m.nav = append(m.nav, navLevel{menu: item.submenu, cursor: 0})
				cmd := m.list.SetItems(toListItems(itemsFor(item.submenu)))
				m.list.Select(0)
				return m, cmd
			}

			return m, m.runAction(item.action)
		}
	}

	// Delegate remaining input to the list (handles Up/Down/PgUp/PgDn/j/k).
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) navigateBack() (tea.Model, tea.Cmd) {
	if len(m.nav) <= 1 {
		return m, nil
	}
	m.nav = m.nav[:len(m.nav)-1]
	parent := m.nav[len(m.nav)-1]
	cmd := m.list.SetItems(toListItems(itemsFor(parent.menu)))
	m.list.Select(parent.cursor)
	return m, cmd
}

// updateThemeMode handles key events while the theme selector is open.
func (m Model) updateThemeMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
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
		m.list.SetDelegate(buildDelegate(m.theme))
		m.cfg.Theme = m.theme.Name
		if err := config.Save(m.cfg); err != nil {
			m.msgKind = msgError
			m.msg = "Falha ao salvar tema: " + err.Error()
		}
		return m, nil
	}
	return m, nil
}

// runAction dispatches an actionID to its domain implementation.
func (m Model) runAction(a actionID) tea.Cmd {
	done := func(err error) tea.Msg { return actionDoneMsg{err: err} }

	exec := func(fn func(context.Context, *executor.Executor, io.Writer) error) tea.Cmd {
		return tea.Exec(newFuncCmd(m.ctx, fn), done)
	}

	execInteractive := func(fn func(context.Context, *executor.Executor, io.Reader, io.Writer) error) tea.Cmd {
		return tea.Exec(newInteractiveFuncCmd(m.ctx, fn), done)
	}

	switch a {
	case actSystemUpdate:
		return exec(update.Run)

	case actSystemPostMint:
		return exec(postinstall.Mint)
	case actSystemPostZorin:
		return exec(postinstall.Zorin)
	case actSystemPostUbuntu:
		return exec(postinstall.Ubuntu)
	case actSystemPostFedora:
		return exec(postinstall.Fedora)

	case actSystemUlauncher:
		return exec(ulauncher.Install)
	case actUlauncherUninstall:
		return exec(ulauncher.Uninstall)

	case actSystemFonts:
		return execInteractive(fonts.Select)

	case actSystemTemplates:
		return execInteractive(templates.Select)

	case actAppsInstall:
		return execInteractive(apps.SelectInstall)

	case actAppsUninstall:
		return execInteractive(apps.SelectUninstall)

	case actAppsWebApps:
		return exec(apps.ShowWebApps)

	// ── Manager: AI / Gitignore / DB / Repo ─────────────────────────────────
	case actAIContext:
		return execInteractive(managerai.GenerateContext)

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

	// ── Dev: Prerequisites / LLMs / IDEs / Terminals / MCP / Upgrade ────────
	case actDevDepends:
		return exec(depends.Install)
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

	// ── Stack: Config ────────────────────────────────────────────────────────
	case actStackSetupPrereqs:
		return exec(stackconfig.SetupPrereqs)
	case actStackWorkspace:
		return exec(stackconfig.Workspace)
	case actStackCompose:
		return exec(stackconfig.Compose)

	// ── Stack: Lifecycle ─────────────────────────────────────────────────────
	case actStackStart:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Start(ctx, exe, w, m.cfg.DockerComposeDir)
		})
	case actStackStop:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Stop(ctx, exe, w, m.cfg.DockerComposeDir)
		})
	case actStackLogs:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.Logs(ctx, exe, w, m.cfg.DockerComposeDir)
		})
	case actStackStats:
		return exec(stack.Stats)
	case actStackDB:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.DBInfo(ctx, exe, w)
		})
	case actStackFixPerms:
		return exec(func(ctx context.Context, exe *executor.Executor, w io.Writer) error {
			return stack.FixPerms(ctx, exe, w, m.cfg.WorkspacePath)
		})

	// ── GNOME ────────────────────────────────────────────────────────────────
	case actGnomePrereqs:
		return exec(gnome.InstallPrereqs)
	case actGnomeExtensions:
		return exec(gnome.ShowExtensions)
	case actGnomeThemes:
		return execInteractive(gnome.ManageThemes)
	case actGnomeIcons:
		return execInteractive(gnome.ManageIcons)
	case actGnomeCursors:
		return execInteractive(gnome.ManageCursors)
	case actGnomeFlatpak:
		return execInteractive(gnome.ApplyFlatpakTheme)

	// ── Lumina ───────────────────────────────────────────────────────────────
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

// ── view ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	s := m.styles
	var sb strings.Builder
	div := s.Divider.Render(strings.Repeat("─", m.width))

	sb.WriteString(renderHeader())
	sb.WriteString(div + "\n")

	if m.themeOpen {
		sb.WriteString(s.Breadcrumb.Render("  "+menuLabels[menuMain]+"  ›  Selecionar Tema") + "\n")
		sb.WriteString(div + "\n\n")
		for i, t := range availableThemes {
			label := t.Name
			if t.Name == m.theme.Name {
				label += "  (atual)"
			}
			if i == m.themeCursor {
				sb.WriteString(s.ActiveBar.Render("|") + s.ActiveText.Render(" "+label) + "\n")
			} else {
				sb.WriteString(s.Inactive.Render("   "+label) + "\n")
			}
		}
		sb.WriteString("\n")
		sb.WriteString(renderThemeFooter(m.width, s))
		return sb.String()
	}

	sb.WriteString(s.Breadcrumb.Render("  "+m.breadcrumb()) + "\n")
	sb.WriteString(div + "\n")
	sb.WriteString(m.list.View())

	if m.msg != "" {
		sb.WriteString("\n")
		switch m.msgKind {
		case msgSuccess:
			sb.WriteString(s.Success.Render("  + "+m.msg) + "\n")
		case msgError:
			sb.WriteString(s.Error.Render("  x "+m.msg) + "\n")
		default:
			sb.WriteString(s.Warning.Render("  "+m.msg) + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(renderFooter(m.width, s))
	return sb.String()
}

// ── delegate ──────────────────────────────────────────────────────────────────

func buildDelegate(t Theme) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.Accent).
		Foreground(t.Accent).
		Bold(true).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.Accent).
		Foreground(t.Primary).
		Padding(0, 0, 0, 1)
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(t.Text).
		Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(0, 0, 0, 2)
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(0, 0, 0, 2)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(0, 0, 0, 2)
	return d
}

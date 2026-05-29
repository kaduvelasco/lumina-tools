package selfupdate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// helpChromeHeight is the number of terminal lines consumed by fixed elements:
// header(7) + divider + breadcrumb + divider + border-top + border-bottom + footer-hints = 13.
const helpChromeHeight = 13

// glamourGutter is glamour's built-in left margin (chars) subtracted from the
// render width so text doesn't overflow the viewport content area.
const glamourGutter = 3

// vpBorderStyle is the rounded border applied to the viewport.
// GetHorizontalFrameSize() returns 4 (border: 2 + margin: 2).
var vpBorderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#9966FF")).
	Margin(0, 1)

var (
	helpDivStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9966FF"))
	helpCrumbStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9966FF"))
	helpHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

// ShowHelp opens a full-screen scrollable help viewer rendered with Glamour.
// The signature matches the execInteractive function type used by the TUI.
func ShowHelp(_ context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	style := glamourStyleFromConfig()
	p := tea.NewProgram(helpModel{glamourStyle: style}, tea.WithInput(stdin), tea.WithOutput(stdout), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// glamourStyleFromConfig maps the user's Lumina theme to a glamour style name.
// Using a fixed style avoids the blocking terminal color-detection query done by
// glamour.WithAutoStyle(), which is the main cause of slow help viewer startup.
func glamourStyleFromConfig() string {
	cfg, err := config.Load()
	if err != nil {
		return "dark"
	}
	switch cfg.Theme {
	case "Claro":
		return "light"
	case "Dracula":
		return "dracula"
	case "Tokyo Night":
		return "tokyo-night"
	default:
		return "dark"
	}
}

// ── model ─────────────────────────────────────────────────────────────────────

type helpModel struct {
	viewport     viewport.Model
	glamourStyle string
	ready        bool
	width        int
	height       int
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		vpW := msg.Width - vpBorderStyle.GetHorizontalFrameSize()
		vpH := msg.Height - helpChromeHeight
		if vpH < 4 {
			vpH = 4
		}

		content := renderHelp(vpW-glamourGutter, m.glamourStyle)

		if !m.ready {
			m.viewport = viewport.New(vpW, vpH)
			m.viewport.Style = vpBorderStyle
			m.viewport.SetContent(content)
			m.ready = true
		} else {
			m.viewport.Width = vpW
			m.viewport.Height = vpH
			m.viewport.SetContent(content)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m helpModel) View() string {
	if !m.ready {
		return "Carregando..."
	}

	div := helpDivStyle.Render(strings.Repeat("─", m.width))
	crumb := helpCrumbStyle.Render("  Lumina Tools  ›  Configurações Lumina  ›  Ajuda")
	hints := helpHintStyle.Render("  ↑↓/jk navegar   PgUp/PgDn página   q/esc fechar")

	var sb strings.Builder
	sb.WriteString(ui.RenderHeader())
	sb.WriteString(div + "\n")
	sb.WriteString(crumb + "\n")
	sb.WriteString(div + "\n")
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")
	sb.WriteString(hints)
	return sb.String()
}

// ── content ───────────────────────────────────────────────────────────────────

func renderHelp(glamourWidth int, style string) string {
	if glamourWidth < 40 {
		glamourWidth = 40
	}
	if style == "" {
		style = "dark"
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(glamourWidth),
	)
	if err != nil {
		return helpMarkdown()
	}
	rendered, err := r.Render(helpMarkdown())
	if err != nil {
		return helpMarkdown()
	}
	return rendered
}

func helpMarkdown() string {
	return fmt.Sprintf(`# Lumina Tools — Ajuda

**Versão:** %s

---

## Atalhos da Interface TUI

| Tecla | Ação |
| --- | --- |
| ↑ ou k | Mover cursor para cima |
| ↓ ou j | Mover cursor para baixo |
| Enter ou Espaço | Selecionar item |
| Esc | Voltar ao menu anterior |
| t | Selecionar tema com preview ao vivo |
| q ou Ctrl+C | Sair |

---

## Gerenciamento Linux

| Comando | Descrição |
| --- | --- |
| lumina system pos mint | Pós instalação do Linux Mint 22.3 |
| lumina system pos zorin | Pós instalação do ZorinOS 18.1 |
| lumina system pos ubuntu | Pós instalação do Ubuntu 26.04 |
| lumina system pos fedora | Pós instalação do Fedora 44 |
| lumina system update | Atualizar o sistema operacional |
| lumina system fonts | Instalar fontes tipográficas |
| lumina system templates | Instalar templates de arquivo |
| lumina system apps install | Instalar aplicativos via Flatpak |
| lumina system apps uninstall | Remover aplicativos via Flatpak |
| lumina system apps webapps | Listar WebApps sugeridos |
| lumina system ulauncher | Instalar o Ulauncher |
| lumina system gnome pre | Instalar pré-requisitos GNOME |
| lumina system gnome ext | Listar extensões recomendadas |
| lumina system gnome themes | Gerenciar temas GTK |
| lumina system gnome icons | Gerenciar pacotes de ícones |
| lumina system gnome cursors | Gerenciar temas de cursor |
| lumina system gnome flatpak | Aplicar tema GTK em apps Flatpak |

---

## DevStuff — Criar Stack

| Comando | Descrição |
| --- | --- |
| lumina stack config pre | Instalar pré-requisitos e Docker Engine |
| lumina stack config workspace | Criar estrutura de diretórios do workspace |
| lumina stack config stack | Gerar docker-compose.yml (Nginx + PHP + MariaDB) |

## DevStuff — Ferramentas de Desenvolvimento

| Comando | Descrição |
| --- | --- |
| lumina dev pre | Instalar dependências base |
| lumina dev llm | Gerenciar CLIs de modelos de linguagem |
| lumina dev ide | Gerenciar ambientes de desenvolvimento |
| lumina dev term | Gerenciar emuladores de terminal |
| lumina dev mcp | Gerenciar servidores MCP |
| lumina dev update | Atualizar todas as ferramentas |

---

## DevManager — Stack em Execução

| Comando | Descrição |
| --- | --- |
| lumina stack start | Iniciar todos os contêineres |
| lumina stack end | Parar todos os contêineres |
| lumina stack log | Exibir logs em tempo real |
| lumina stack status | Status e uso de recursos |
| lumina stack db | Dados de conexão do banco de dados |
| lumina stack fix-perm | Corrigir permissões do workspace |

## DevManager — Banco de Dados

| Comando | Descrição |
| --- | --- |
| lumina db backup | Criar backup do banco de dados |
| lumina db restore | Restaurar a partir de backup |
| lumina db remove | Remover banco de dados |
| lumina db optimize | Verificar e otimizar tabelas |
| lumina db moodle | Otimizar para banco Moodle |

## DevManager — Repositórios Git

| Comando | Descrição |
| --- | --- |
| lumina repo global | Configurar identidade global do Git |
| lumina repo init | Iniciar novo repositório local |
| lumina repo clone | Clonar repositório remoto |
| lumina repo ident | Aplicar identidade a um repositório |

## DevManager — IA e Versionamento

| Comando | Descrição |
| --- | --- |
| lumina ai | Gerar contexto para assistentes de IA |
| lumina gitignore | Criar ou atualizar o .gitignore |

---

## Configurações Lumina

| Comando | Descrição |
| --- | --- |
| lumina self-update | Verificar e instalar atualização |
| lumina self-uninstall | Remover o Lumina Tools do sistema |
| lumina help | Exibir esta ajuda em modo texto |
| lumina version | Exibir a versão instalada |

---

## Configuração

**Arquivo:** ~/.lumina/config.yaml

| Campo | Descrição |
| --- | --- |
| workspace_path | Caminho do workspace de desenvolvimento |
| docker_compose_dir | Diretório do arquivo docker-compose.yml |
| theme | Tema: Lumina, Claro, Dracula, Nord, Tokyo Night, Gruvbox |
| flatpak_scope | Escopo de instalação Flatpak: system (padrão) ou user |

Exemplos de configuração via terminal:

    lumina set workspace ~/workspace
    lumina set docker ~/workspace/docker
    lumina set theme dracula
    lumina set flatpak user

`, version.Version)
}

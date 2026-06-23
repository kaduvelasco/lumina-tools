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

// ShowHelp opens a full-screen scrollable help viewer rendered with Glamour.
// The signature matches the execInteractive function type used by the TUI.
// Border, divider, breadcrumb and hint colors follow the user's selected
// theme (internal/ui.LoadTheme), like every other script that runs outside
// the main TUI render loop — never hardcode colors here.
func ShowHelp(_ context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	theme := ui.LoadTheme()
	m := helpModel{
		glamourStyle: glamourStyleFromConfig(),
		vpBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Margin(0, 1),
		primaryStyle: lipgloss.NewStyle().Foreground(theme.Primary),
		mutedStyle:   lipgloss.NewStyle().Foreground(theme.Muted),
	}
	p := tea.NewProgram(m, tea.WithInput(stdin), tea.WithOutput(stdout), tea.WithAltScreen())
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
	vpBorder     lipgloss.Style
	primaryStyle lipgloss.Style
	mutedStyle   lipgloss.Style
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		vpW := msg.Width - m.vpBorder.GetHorizontalFrameSize()
		vpH := msg.Height - helpChromeHeight
		if vpH < 4 {
			vpH = 4
		}

		content := renderHelp(vpW-glamourGutter, m.glamourStyle)

		if !m.ready {
			m.viewport = viewport.New(vpW, vpH)
			m.viewport.Style = m.vpBorder
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

	div := m.primaryStyle.Render(strings.Repeat("─", m.width))
	crumb := m.primaryStyle.Render("  Lumina Tools  ›  Configurações Lumina  ›  Ajuda")
	hints := m.mutedStyle.Render("  ↑↓/jk navegar   PgUp/PgDn página   q/esc fechar")

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
		return HelpMarkdown()
	}
	rendered, err := r.Render(HelpMarkdown())
	if err != nil {
		return HelpMarkdown()
	}
	return rendered
}

// HelpMarkdown returns the full help content in Markdown format.
func HelpMarkdown() string {
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
| lumina system toys | Instalar Linux Toys |
| lumina system megasync | Instalar MegaSync (pacote detectado pela distro) |

## Aplicativos Linux

| Comando | Descrição |
| --- | --- |
| lumina apps install | Instalar aplicativos via Flatpak |
| lumina apps uninstall | Remover aplicativos via Flatpak |
| lumina apps web | Listar WebApps sugeridos |

## Personalizar Linux

| Comando | Descrição |
| --- | --- |
| lumina theme gnome-pre | Instalar pré-requisitos GNOME |
| lumina theme gnome | Gerenciar temas GTK GNOME |
| lumina theme extensions | Listar extensões GNOME recomendadas |
| lumina theme cinnamon-pre | Instalar pré-requisitos Cinnamon |
| lumina theme cinnamon | Gerenciar temas GTK Cinnamon |
| lumina theme cursor | Gerenciar temas de cursor |
| lumina theme icons | Gerenciar pacotes de ícones |
| lumina theme flatpak | Aplicar tema GTK em apps Flatpak |

---

## DevStuff — Criar Stack

| Comando | Descrição |
| --- | --- |
| lumina stack config docker | Instalar pré-requisitos e Docker Engine |
| lumina stack config workspace | Criar estrutura de diretórios do workspace |
| lumina stack config stack | Gerar docker-compose.yml (Nginx + PHP + MariaDB) |

## DevStuff — Ferramentas de Desenvolvimento

| Comando | Descrição |
| --- | --- |
| lumina dev pre | Instalar dependências base |
| lumina dev go | Instalar ou atualizar o Go |
| lumina dev flutter | Instalar ou atualizar o Flutter + Dart |
| lumina dev llm | Gerenciar CLIs de modelos de linguagem |
| lumina dev ide | Gerenciar ambientes de desenvolvimento |
| lumina dev term | Gerenciar emuladores de terminal |
| lumina dev mcp | Gerenciar servidores MCP |
| lumina dev update | Atualizar todas as ferramentas |
| lumina dev create-workspace | Criar estrutura de workspace (atalho) |
| lumina dev create-stack-php | Gerar docker-compose.yml da stack PHP (atalho) |

---

## DevManager — Stack em Execução

| Comando | Descrição |
| --- | --- |
| lumina stack start | Iniciar todos os contêineres |
| lumina stack end | Parar todos os contêineres |
| lumina stack restart | Reiniciar todos os contêineres |
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
| lumina repo gitignore | Criar ou atualizar o .gitignore pela stack detectada |
| lumina repo conduct | Criar código de conduta |

## DevManager — Contextos IA

| Comando | Descrição |
| --- | --- |
| lumina ai context | Gerar ou atualizar contexto para assistentes de IA |
| lumina ai clear | Remover contextos de IA do diretório atual |

---

## Configurações Lumina

| Comando | Descrição |
| --- | --- |
| lumina self-update | Verificar e instalar atualização |
| lumina self-uninstall | Remover o Lumina Tools do sistema |
| lumina self-config | Configurar o Lumina interativamente |
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

package selfupdate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// helpSection is a titled group of command/description rows, shared by the
// interactive panel view (ShowHelp) and the plain-text dump (HelpText).
type helpSection struct {
	title string
	rows  [][2]string
}

var helpSections = []helpSection{
	{"Atalhos da Interface TUI", [][2]string{
		{"Tecla", "Ação"},
		{"↑ ou k", "Mover cursor para cima"},
		{"↓ ou j", "Mover cursor para baixo"},
		{"Enter ou Espaço", "Selecionar item"},
		{"Esc", "Voltar ao menu anterior"},
		{"t", "Selecionar tema com preview ao vivo"},
		{"q ou Ctrl+C", "Sair"},
	}},
	{"Gerenciamento Linux", [][2]string{
		{"lumina system pos mint", "Pós instalação do Linux Mint 22.3 Cinnamon"},
		{"lumina system pos mint-xfce", "Pós instalação do Linux Mint 22.3 XFCE"},
		{"lumina system pos zorin", "Pós instalação do ZorinOS 18.1"},
		{"lumina system pos ubuntu", "Pós instalação do Ubuntu 26.04"},
		{"lumina system pos fedora", "Pós instalação do Fedora 44"},
		{"lumina system update", "Atualizar o sistema operacional"},
		{"lumina system fonts", "Instalar fontes tipográficas"},
		{"lumina system templates", "Instalar templates de arquivo"},
		{"lumina system toys", "Instalar Linux Toys"},
		{"lumina system megasync", "Instalar MegaSync (pacote detectado pela distro)"},
	}},
	{"Aplicativos Linux", [][2]string{
		{"lumina apps install", "Instalar aplicativos via Flatpak"},
		{"lumina apps uninstall", "Remover aplicativos via Flatpak"},
		{"lumina apps web", "Listar WebApps sugeridos"},
	}},
	{"Personalizar Linux", [][2]string{
		{"lumina theme gnome-pre", "Instalar pré-requisitos GNOME"},
		{"lumina theme gnome", "Gerenciar temas GTK GNOME"},
		{"lumina theme extensions", "Listar extensões GNOME recomendadas"},
		{"lumina theme cinnamon-pre", "Instalar pré-requisitos Cinnamon"},
		{"lumina theme cinnamon", "Gerenciar temas GTK Cinnamon"},
		{"lumina theme cursor", "Gerenciar temas de cursor"},
		{"lumina theme icons", "Gerenciar pacotes de ícones"},
		{"lumina theme flatpak", "Aplicar tema GTK em apps Flatpak"},
	}},
	{"DevStuff — Criar Stack", [][2]string{
		{"lumina stack config docker", "Instalar pré-requisitos e Docker Engine"},
		{"lumina stack config workspace", "Criar estrutura de diretórios do workspace"},
		{"lumina stack config stack", "Gerar docker-compose.yml (Nginx + PHP + MariaDB)"},
	}},
	{"DevStuff — Ferramentas de Desenvolvimento", [][2]string{
		{"lumina dev pre", "Instalar pré-requisitos de desenvolvimento"},
		{"lumina dev go", "Gerenciar o Go (instalar, atualizar ou remover)"},
		{"lumina dev flutter", "Gerenciar Flutter + Dart (instalar, atualizar ou remover)"},
		{"lumina dev android-studio", "Gerenciar Android Studio (instalar, reinstalar ou remover)"},
		{"lumina dev antigravity", "Gerenciar Antigravity IDE (instalar, reinstalar ou remover)"},
		{"lumina dev llm", "Gerenciar CLIs de modelos de linguagem"},
		{"lumina dev ide", "Gerenciar ambientes de desenvolvimento"},
		{"lumina dev term", "Gerenciar emuladores de terminal"},
		{"lumina dev mcp", "Gerenciar servidores MCP"},
		{"lumina dev create-workspace", "Criar estrutura de workspace (atalho)"},
		{"lumina dev create-stack-php", "Gerar docker-compose.yml da stack PHP (atalho)"},
	}},
	{"DevManager — Stack em Execução", [][2]string{
		{"lumina stack start", "Iniciar todos os contêineres"},
		{"lumina stack end", "Parar todos os contêineres"},
		{"lumina stack restart", "Reiniciar todos os contêineres"},
		{"lumina stack log", "Exibir logs em tempo real"},
		{"lumina stack status", "Status e uso de recursos"},
		{"lumina stack db", "Dados de conexão do banco de dados"},
		{"lumina stack fix-perm", "Corrigir permissões do workspace"},
		{"lumina stack moodle", "Configurar roteamento Moodle 5.1+ sem recriar a stack"},
	}},
	{"DevManager — Banco de Dados", [][2]string{
		{"lumina db backup", "Criar backup do banco de dados"},
		{"lumina db restore", "Restaurar a partir de backup"},
		{"lumina db remove", "Remover banco de dados"},
		{"lumina db optimize", "Verificar e otimizar tabelas"},
		{"lumina db moodle", "Otimizar para banco Moodle"},
	}},
	{"DevManager — Repositórios Git", [][2]string{
		{"lumina repo global", "Configurar identidade global do Git"},
		{"lumina repo init", "Iniciar novo repositório local"},
		{"lumina repo clone", "Clonar repositório remoto"},
		{"lumina repo ident", "Aplicar identidade a um repositório"},
		{"lumina repo gitignore", "Criar ou atualizar o .gitignore pela stack detectada"},
		{"lumina repo conduct", "Criar código de conduta"},
	}},
	{"DevManager — Contextos IA", [][2]string{
		{"lumina ai context", "Gerar ou atualizar contexto para assistentes de IA"},
		{"lumina ai clear", "Remover contextos de IA do diretório atual"},
	}},
	{"Configurações Lumina", [][2]string{
		{"lumina self-update", "Verificar e instalar atualização"},
		{"lumina self-uninstall", "Remover o Lumina Tools do sistema"},
		{"lumina self-config", "Configurar o Lumina interativamente"},
		{"lumina help", "Exibir esta ajuda em modo texto"},
		{"lumina version", "Exibir a versão instalada"},
	}},
	{"Configuração — Arquivo: ~/.lumina/config.yaml", [][2]string{
		{"workspace_path", "Caminho do workspace de desenvolvimento"},
		{"docker_compose_dir", "Diretório do arquivo docker-compose.yml"},
		{"theme", "Tema: Lumina, Claro, Dracula, Nord, Tokyo Night, Gruvbox"},
		{"flatpak_scope", "Escopo de instalação Flatpak: system (padrão) ou user"},
	}},
}

// helpExamples lists CLI snippets shown after the config section, both in
// ShowHelp and HelpText.
var helpExamples = []string{
	"lumina set workspace ~/workspace",
	"lumina set docker ~/workspace/docker",
	"lumina set theme dracula",
	"lumina set flatpak user",
}

// ShowHelp prints the command reference using the same header/panel pattern
// as every other script in the app (ui.PrintHeader/Info/PrintBox), instead of
// a dedicated full-screen viewer. The signature matches the execInteractive
// function type used by the TUI, even though stdin is unused here — WaitEnter
// always reads os.Stdin directly.
func ShowHelp(_ context.Context, _ *executor.Executor, _ io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Ajuda")
	ui.Info(stdout, "Lumina Tools — Ajuda\nVersão: "+version.Version)

	for _, s := range helpSections {
		ui.Info(stdout, s.title)
		ui.PrintBox(stdout, sectionRows(s))
	}

	ui.Info(stdout, "Exemplos de configuração via terminal:\n\n  "+strings.Join(helpExamples, "\n  "))

	ui.WaitEnter(stdout)
	return nil
}

// HelpText renders the same content as ShowHelp, but as a single plain-text
// block with no panels or terminal control codes — used by "lumina help" for
// piping/redirection.
func HelpText() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Lumina Tools — Ajuda\nVersão: %s\n", version.Version)
	for _, s := range helpSections {
		sb.WriteString("\n" + s.title + "\n")
		sb.WriteString(sectionRows(s) + "\n")
	}
	sb.WriteString("\nExemplos de configuração via terminal:\n\n  " + strings.Join(helpExamples, "\n  ") + "\n")
	return sb.String()
}

func sectionRows(s helpSection) string {
	rows := make([]string, len(s.rows))
	for i, r := range s.rows {
		rows[i] = fmt.Sprintf("%-30s %s", r[0], r[1])
	}
	return strings.Join(rows, "\n")
}

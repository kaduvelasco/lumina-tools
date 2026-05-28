package tui

import "github.com/charmbracelet/bubbles/list"

type menuID int

const (
	menuMain menuID = iota + 1
	menuSystem
	menuSystemPostInstall
	menuSystemApps
	menuSystemAppsInstall
	menuSystemAppsUninstall
	menuDev
	menuDevStack
	menuDevTools
	menuManager
	menuManagerStack
	menuManagerDB
	menuManagerRepo
	menuLumina
	menuGnome
)

type actionID int

const (
	actNone actionID = iota
	actBack
	// System
	actSystemPostMint
	actSystemPostZorin
	actSystemPostUbuntu
	actSystemPostFedora
	actSystemFonts
	actSystemTemplates
	actAppsInstall
	actAppsUninstall
	actSystemUpdate
	actSystemUlauncher
	actUlauncherUninstall
	// Stack setup
	actStackSetupPrereqs
	actStackWorkspace
	actStackCompose
	// Stack lifecycle
	actStackStart
	actStackStop
	actStackLogs
	actStackStats
	actStackDB
	actStackFixPerms
	// Dev tools
	actDevDepends
	actLLMManage
	actIDEManage
	actTermManage
	actMCPManage
	actDevUpgrade
	// Manager
	actAIContext
	actGitignore
	actDBBackup
	actDBRestore
	actDBRemove
	actDBOptimize
	actDBMoodle
	actRepoGlobal
	actRepoInit
	actRepoClone
	actRepoIdent
	// Lumina
	actLuminaUpdate
	actLuminaUninstall
	actLuminaHelp
	// GNOME
	actGnomePrereqs
	actGnomeExtensions
	actGnomeThemes
	actGnomeIcons
	actGnomeCursors
)

type menuItem struct {
	label       string
	description string
	submenu     menuID
	action      actionID
}

// Title appends "  ›" to submenu items so the user can distinguish navigation from actions.
func (i menuItem) Title() string {
	if i.submenu != 0 {
		return i.label + "  ›"
	}
	return i.label
}

func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.label }

func toListItems(items []menuItem) []list.Item {
	result := make([]list.Item, len(items))
	for idx, item := range items {
		result[idx] = item
	}
	return result
}

var menuLabels = map[menuID]string{
	menuMain:                "Lumina Tools",
	menuSystem:              "Gerenciamento Linux",
	menuSystemPostInstall:   "Pós Instalação",
	menuSystemApps:          "Aplicativos",
	menuSystemAppsInstall:   "Instalar",
	menuSystemAppsUninstall: "Desinstalar",
	menuDev:                 "DevStuff",
	menuDevStack:            "Criar Stack",
	menuDevTools:            "Gerenciar Ferramentas",
	menuManager:             "DevManager",
	menuManagerStack:        "Gerenciar Stack",
	menuManagerDB:           "Banco de Dados",
	menuManagerRepo:         "Repositórios",
	menuLumina:              "Configurações Lumina",
	menuGnome:               "Customizar GNOME",
}

func itemsFor(m menuID) []menuItem {
	switch m {
	case menuMain:
		return []menuItem{
			{label: "Gerenciamento Linux", description: "Atualização do sistema, fontes, aplicativos e customização GNOME", submenu: menuSystem},
			{label: "DevStuff", description: "Ferramentas de desenvolvimento e criação de stack Docker", submenu: menuDev},
			{label: "DevManager", description: "Gerenciamento de projetos, banco de dados e repositórios", submenu: menuManager},
			{label: "Configurações Lumina", description: "Atualização, desinstalação e ajuda do Lumina Tools", submenu: menuLumina},
		}
	case menuSystem:
		return []menuItem{
			{label: "Pós Instalação", description: "Scripts de configuração inicial para diferentes distribuições Linux", submenu: menuSystemPostInstall},
			{label: "Atualizar Sistema", description: "Executa atualização completa do sistema operacional", action: actSystemUpdate},
			{label: "Instalar Fontes", description: "Seleciona e instala pacotes de fontes tipográficas no sistema", action: actSystemFonts},
			{label: "Templates de Arquivos", description: "Instala templates para criação rápida de novos arquivos", action: actSystemTemplates},
			{label: "Aplicativos", description: "Instala ou desinstala aplicativos Flatpak e Ulauncher", submenu: menuSystemApps},
			{label: "Customizar GNOME", description: "Extensões, temas, ícones e cursores para o ambiente GNOME", submenu: menuGnome},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuSystemPostInstall:
		return []menuItem{
			{label: "Linux Mint 22.3", description: "Configura o ambiente após a instalação do Linux Mint 22.3", action: actSystemPostMint},
			{label: "ZorinOS 18.1", description: "Configura o ambiente após a instalação do ZorinOS 18.1", action: actSystemPostZorin},
			{label: "Ubuntu 26.04", description: "Configura o ambiente após a instalação do Ubuntu 26.04", action: actSystemPostUbuntu},
			{label: "Fedora 44", description: "Configura o ambiente após a instalação do Fedora 44", action: actSystemPostFedora},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuSystemApps:
		return []menuItem{
			{label: "Instalar", description: "Seleciona e instala aplicativos Flatpak ou o Ulauncher", submenu: menuSystemAppsInstall},
			{label: "Desinstalar", description: "Remove aplicativos Flatpak ou o Ulauncher do sistema", submenu: menuSystemAppsUninstall},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuSystemAppsInstall:
		return []menuItem{
			{label: "Aplicativos Flatpak", description: "Abre seletor interativo para instalar aplicativos via Flatpak", action: actAppsInstall},
			{label: "Ulauncher", description: "Instala o Ulauncher, lançador de aplicativos rápido", action: actSystemUlauncher},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuSystemAppsUninstall:
		return []menuItem{
			{label: "Aplicativos Flatpak", description: "Abre seletor interativo para remover aplicativos instalados via Flatpak", action: actAppsUninstall},
			{label: "Ulauncher", description: "Remove o Ulauncher e seus dados de configuração do sistema", action: actUlauncherUninstall},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuDev:
		return []menuItem{
			{label: "Criar Stack de Desenvolvimento", description: "Configura Docker, workspace e ambiente de desenvolvimento local", submenu: menuDevStack},
			{label: "Gerenciar Ferramentas de Desenvolvimento", description: "Instala e atualiza CLIs, IDEs, terminais e servidores MCP", submenu: menuDevTools},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuDevStack:
		return []menuItem{
			{label: "Instalar Pré-requisitos", description: "Instala pacotes base e Docker Engine via gerenciador de pacotes", action: actStackSetupPrereqs},
			{label: "Criar Estrutura do Workspace", description: "Cria a estrutura de diretórios do workspace de desenvolvimento", action: actStackWorkspace},
			{label: "Criar Stack (Nginx + PHP + MariaDB)", description: "Gera docker-compose.yml com Nginx, PHP-FPM e MariaDB configurados", action: actStackCompose},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuDevTools:
		return []menuItem{
			{label: "Instalar Pré-requisitos", description: "Instala dependências base para as ferramentas de desenvolvimento", action: actDevDepends},
			{label: "Gerenciar CLIs LLM", description: "Instala ou remove CLIs de modelos de linguagem (Claude, Gemini, etc.)", action: actLLMManage},
			{label: "Gerenciar IDEs", description: "Instala ou remove ambientes de desenvolvimento integrado", action: actIDEManage},
			{label: "Gerenciar Terminais", description: "Instala ou remove emuladores de terminal alternativos", action: actTermManage},
			{label: "Gerenciar Servidores MCP", description: "Configura servidores do Model Context Protocol para assistentes IA", action: actMCPManage},
			{label: "Atualizar Ferramentas", description: "Atualiza todas as ferramentas de desenvolvimento instaladas", action: actDevUpgrade},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuManager:
		return []menuItem{
			{label: "Gerenciar Stack", description: "Inicia, finaliza e monitora a stack Docker em execução", submenu: menuManagerStack},
			{label: "Criar Contexto IA", description: "Gera arquivos de contexto para assistentes de IA no projeto atual", action: actAIContext},
			{label: "Criar/Atualizar .gitignore", description: "Gera ou atualiza o arquivo .gitignore do projeto atual", action: actGitignore},
			{label: "Gerenciar Banco de Dados", description: "Backup, restauração, remoção e otimização de bancos de dados", submenu: menuManagerDB},
			{label: "Gerenciar Repositórios", description: "Configuração de identidade e operações em repositórios Git", submenu: menuManagerRepo},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuManagerStack:
		return []menuItem{
			{label: "Iniciar", description: "Inicia todos os contêineres da stack Docker", action: actStackStart},
			{label: "Finalizar", description: "Para e remove todos os contêineres da stack Docker", action: actStackStop},
			{label: "Logs", description: "Exibe os logs em tempo real de todos os serviços da stack", action: actStackLogs},
			{label: "Status e Recursos", description: "Mostra status dos contêineres e uso de CPU, memória e rede", action: actStackStats},
			{label: "Dados do DB", description: "Exibe informações de conexão e dados do banco de dados ativo", action: actStackDB},
			{label: "Ajustar Permissões", description: "Corrige permissões de arquivos e diretórios no workspace", action: actStackFixPerms},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuManagerDB:
		return []menuItem{
			{label: "Backup", description: "Cria um arquivo de backup do banco de dados selecionado", action: actDBBackup},
			{label: "Restaurar", description: "Restaura um banco de dados a partir de um arquivo de backup", action: actDBRestore},
			{label: "Apagar", description: "Remove permanentemente um banco de dados do servidor", action: actDBRemove},
			{label: "Verificar / Otimizar", description: "Verifica a integridade e otimiza as tabelas do banco de dados", action: actDBOptimize},
			{label: "Otimizar para Moodle", description: "Aplica otimizações de desempenho específicas para bancos Moodle", action: actDBMoodle},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuManagerRepo:
		return []menuItem{
			{label: "Configurar Identidade Global", description: "Define nome de usuário e e-mail globais do Git", action: actRepoGlobal},
			{label: "Iniciar Novo Repositório", description: "Inicializa um novo repositório Git no diretório atual", action: actRepoInit},
			{label: "Clonar Repositório", description: "Clona um repositório remoto para o diretório local", action: actRepoClone},
			{label: "Aplicar Identidade", description: "Aplica uma identidade Git específica a um repositório local", action: actRepoIdent},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuLumina:
		return []menuItem{
			{label: "Atualizar Lumina Tools", description: "Verifica e instala a versão mais recente do Lumina Tools", action: actLuminaUpdate},
			{label: "Desinstalar Lumina Tools", description: "Remove o binário e todas as configurações do Lumina Tools", action: actLuminaUninstall},
			{label: "Ajuda", description: "Exibe a referência completa de todos os comandos disponíveis", action: actLuminaHelp},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	case menuGnome:
		return []menuItem{
			{label: "Instalar pré-requisitos de customização", description: "Instala pacotes necessários para personalização do GNOME", action: actGnomePrereqs},
			{label: "Mostrar extensões recomendadas", description: "Lista extensões GNOME recomendadas com links de instalação", action: actGnomeExtensions},
			{label: "Gerenciar Temas", description: "Instala ou remove temas GTK para o ambiente GNOME", action: actGnomeThemes},
			{label: "Gerenciar Ícones", description: "Instala ou remove pacotes de ícones para o ambiente GNOME", action: actGnomeIcons},
			{label: "Gerenciar Cursores", description: "Instala ou remove temas de cursor para o ambiente GNOME", action: actGnomeCursors},
			{label: "Voltar", description: "Retorna ao menu anterior", action: actBack},
		}
	}
	return nil
}

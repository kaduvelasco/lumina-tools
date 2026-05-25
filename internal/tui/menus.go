package tui

type menuID int

const (
	menuMain menuID = iota + 1
	menuSystem
	menuSystemPostInstall
	menuStack
	menuStackConfig
	menuDev
	menuManager
	menuManagerDB
	menuManagerRepo
	menuLumina
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
	// Stack
	actStackDepends
	actStackDocker
	actStackWorkspace
	actStackCompose
	actStackStart
	actStackStop
	actStackLogs
	actStackStats
	actStackDB
	actStackFixPerms
	// Dev
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
)

type menuItem struct {
	label   string
	submenu menuID
	action  actionID
}

var menuLabels = map[menuID]string{
	menuMain:              "Lumina Tools",
	menuSystem:            "Gerenciamento Linux",
	menuSystemPostInstall: "Pós Instalação",
	menuStack:             "DevStack",
	menuStackConfig:       "Configurar",
	menuDev:               "DevStuff",
	menuManager:           "DevManager",
	menuManagerDB:         "Banco de Dados",
	menuManagerRepo:       "Repositórios",
	menuLumina:            "Configurações Lumina",
}

func itemsFor(m menuID) []menuItem {
	switch m {
	case menuMain:
		return []menuItem{
			{label: "Gerenciamento Linux", submenu: menuSystem},
			{label: "DevStack", submenu: menuStack},
			{label: "DevStuff", submenu: menuDev},
			{label: "DevManager", submenu: menuManager},
			{label: "Configurações Lumina", submenu: menuLumina},
		}
	case menuSystem:
		return []menuItem{
			{label: "Pós Instalação", submenu: menuSystemPostInstall},
			{label: "Instalar Fontes", action: actSystemFonts},
			{label: "Templates de Arquivos", action: actSystemTemplates},
			{label: "Instalar Aplicativos", action: actAppsInstall},
			{label: "Desinstalar Aplicativos", action: actAppsUninstall},
			{label: "Atualizar Sistema", action: actSystemUpdate},
			{label: "Instalar Ulauncher", action: actSystemUlauncher},
			{label: "Voltar", action: actBack},
		}
	case menuSystemPostInstall:
		return []menuItem{
			{label: "Linux Mint 22.3", action: actSystemPostMint},
			{label: "ZorinOS 18.1", action: actSystemPostZorin},
			{label: "Ubuntu 26.04", action: actSystemPostUbuntu},
			{label: "Fedora 44", action: actSystemPostFedora},
			{label: "Voltar", action: actBack},
		}
	case menuStack:
		return []menuItem{
			{label: "Configurar", submenu: menuStackConfig},
			{label: "Iniciar Stack", action: actStackStart},
			{label: "Finalizar Stack", action: actStackStop},
			{label: "Visualizar Logs", action: actStackLogs},
			{label: "Status e Recursos", action: actStackStats},
			{label: "Dados do DB", action: actStackDB},
			{label: "Corrigir Permissões", action: actStackFixPerms},
			{label: "Voltar", action: actBack},
		}
	case menuStackConfig:
		return []menuItem{
			{label: "Instalar Pré-requisitos", action: actStackDepends},
			{label: "Instalar Docker", action: actStackDocker},
			{label: "Criar Workspace", action: actStackWorkspace},
			{label: "Criar Stack (docker-compose)", action: actStackCompose},
			{label: "Voltar", action: actBack},
		}
	case menuDev:
		return []menuItem{
			{label: "Instalar Pré-requisitos", action: actDevDepends},
			{label: "Gerenciar CLIs LLM", action: actLLMManage},
			{label: "Gerenciar IDEs", action: actIDEManage},
			{label: "Gerenciar Terminais", action: actTermManage},
			{label: "Gerenciar Servidores MCP", action: actMCPManage},
			{label: "Atualizar Ferramentas", action: actDevUpgrade},
			{label: "Voltar", action: actBack},
		}
	case menuManager:
		return []menuItem{
			{label: "Criar Contexto AI", action: actAIContext},
			{label: "Criar/Atualizar .gitignore", action: actGitignore},
			{label: "Gerenciar Banco de Dados", submenu: menuManagerDB},
			{label: "Gerenciar Repositórios", submenu: menuManagerRepo},
			{label: "Voltar", action: actBack},
		}
	case menuManagerDB:
		return []menuItem{
			{label: "Backup", action: actDBBackup},
			{label: "Restaurar", action: actDBRestore},
			{label: "Remover Banco", action: actDBRemove},
			{label: "Verificar / Otimizar", action: actDBOptimize},
			{label: "Otimizar para Moodle", action: actDBMoodle},
			{label: "Voltar", action: actBack},
		}
	case menuManagerRepo:
		return []menuItem{
			{label: "Configurar Identidade Global", action: actRepoGlobal},
			{label: "Iniciar Novo Repositório", action: actRepoInit},
			{label: "Clonar Repositório", action: actRepoClone},
			{label: "Aplicar Identidade", action: actRepoIdent},
			{label: "Voltar", action: actBack},
		}
	case menuLumina:
		return []menuItem{
			{label: "Atualizar Lumina Tools", action: actLuminaUpdate},
			{label: "Desinstalar Lumina Tools", action: actLuminaUninstall},
			{label: "Ajuda", action: actLuminaHelp},
			{label: "Voltar", action: actBack},
		}
	}
	return nil
}

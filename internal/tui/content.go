package tui

// ── action IDs ────────────────────────────────────────────────────────────────

type actionID int

const (
	actNone actionID = iota
	actBack
	actPrereqs
	// System
	actSystemPostMint
	actSystemPostZorin
	actSystemPostUbuntu
	actSystemPostFedora
	actSystemFonts
	actSystemTemplates
	actAppsInstall
	actAppsUninstall
	actAppsWebApps
	actSystemUpdate
	// Stack setup
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
	actGoManage
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
	actRepoConduct
	// Lumina
	actLuminaUpdate
	actLuminaUninstall
	actLuminaHelp
	// Personalizar
	actGnomePrereqs
	actGnomeExtensions
	actGnomeThemes
	actGnomeIcons
	actGnomeCursors
	actGnomeFlatpak
	// New in reorganization
	actLuminaConfig
	actStackRestart
	actAIContextRemove
	// System extras
	actLinuxToys
	actMegaSync
	// v2 only
	actQuit
)

// ── section indices ───────────────────────────────────────────────────────────
//
// These constants mirror the position of each entry in the sections slice.
// Update them whenever a section is inserted or removed.

const (
	sectionHome     = 0
	sectionLinux    = 1
	sectionApps     = 2
	sectionCustom   = 3
	sectionDev      = 4
	sectionStack    = 5
	sectionDatabase = 6
	sectionRepo     = 7
	sectionAI       = 8
)

// ── menu data ─────────────────────────────────────────────────────────────────

// submenuEntry is one actionable item in the right-panel list.
// cmd is shown as the second line; if empty, desc is shown instead.
// pending marks items that are not yet implemented — shown as disabled.
type submenuEntry struct {
	title   string
	cmd     string
	desc    string
	action  actionID
	pending bool
}

// section is one of the persistent entries in the left-panel sidebar.
type section struct {
	label string
	items []submenuEntry
}

// sections is the authoritative menu data for the v2 layout.
var sections = []section{
	{
		label: "Home",
		items: []submenuEntry{
			{title: "Sobre", cmd: "lumina", action: actNone},
			{title: "Atualizar", cmd: "lumina self-update", action: actLuminaUpdate},
			{title: "Configurar", cmd: "lumina self-config", action: actLuminaConfig},
			{title: "Ajuda", cmd: "lumina help", action: actLuminaHelp},
			{title: "Desinstalar", cmd: "lumina self-uninstall", action: actLuminaUninstall},
			{title: "Sair", desc: "Encerra o programa após confirmação.", action: actQuit},
		},
	},
	{
		label: "Gerenciamento Linux",
		items: []submenuEntry{
			{title: "Pós instalação · Mint 22.3", cmd: "lumina system pos mint", action: actSystemPostMint},
			{title: "Pós instalação · ZorinOS 18.1", cmd: "lumina system pos zorin", action: actSystemPostZorin},
			{title: "Pós instalação · Ubuntu 26.04", cmd: "lumina system pos ubuntu", action: actSystemPostUbuntu},
			{title: "Pós instalação · Fedora Gnome 44", cmd: "lumina system pos fedora", action: actSystemPostFedora},
			{title: "Gerenciar Fontes", cmd: "lumina system fonts", action: actSystemFonts},
			{title: "Gerenciar Templates de Arquivos", cmd: "lumina system templates", action: actSystemTemplates},
			{title: "Instalar Linux Toys", cmd: "lumina system toys", action: actLinuxToys},
			{title: "Instalar MegaSync", cmd: "lumina system megasync", action: actMegaSync},
		},
	},
	{
		label: "Aplicativos Linux",
		items: []submenuEntry{
			{title: "Instalar", cmd: "lumina apps install", action: actAppsInstall},
			{title: "Desinstalar", cmd: "lumina apps uninstall", action: actAppsUninstall},
			{title: "WebApps recomendados", cmd: "lumina apps web", action: actAppsWebApps},
		},
	},
	{
		label: "Personalizar Linux",
		items: []submenuEntry{
			{title: "Pré-requisitos Gnome", cmd: "lumina gnome pre", action: actGnomePrereqs},
			{title: "Temas Gnome", cmd: "lumina gnome themes", action: actGnomeThemes},
			{title: "Cursor", cmd: "lumina gnome cursor", action: actGnomeCursors},
			{title: "Ícones", cmd: "lumina gnome icons", action: actGnomeIcons},
			{title: "Aplicar tema Flathub", cmd: "lumina system gnome flatpak", action: actGnomeFlatpak},
		},
	},
	{
		label: "Ambiente de Desenvolvimento",
		items: []submenuEntry{
			{title: "Instalar pré-requisitos", cmd: "lumina dev pre", action: actPrereqs},
			{title: "Criar Workspace", cmd: "lumina dev create-workspace", action: actStackWorkspace},
			{title: "Criar Stack PHP", cmd: "lumina dev create-stack-php", action: actStackCompose},
			{title: "Gerenciar Go", cmd: "lumina dev go", action: actGoManage},
			{title: "Gerenciar CLIs LLM", cmd: "lumina dev llm", action: actLLMManage},
			{title: "Gerenciar IDEs", cmd: "lumina dev ide", action: actIDEManage},
			{title: "Gerenciar Terminais", cmd: "lumina dev term", action: actTermManage},
			{title: "Gerenciar MCPs", cmd: "lumina dev mcp", action: actMCPManage},
			{title: "Atualizar ferramentas", cmd: "lumina dev update", action: actDevUpgrade},
		},
	},
	{
		label: "Gerenciar Stack PHP",
		items: []submenuEntry{
			{title: "Iniciar Stack PHP", cmd: "lumina stack start", action: actStackStart},
			{title: "Parar Stack PHP", cmd: "lumina stack end", action: actStackStop},
			{title: "Reiniciar Stack PHP", cmd: "lumina stack restart", action: actStackRestart},
			{title: "Status da Stack PHP", cmd: "lumina stack status", action: actStackStats},
			{title: "Logs da Stack PHP", cmd: "lumina stack log", action: actStackLogs},
			{title: "Ajustar permissões", cmd: "lumina stack fix-perm", action: actStackFixPerms},
		},
	},
	{
		label: "Gerenciar banco de Dados",
		items: []submenuEntry{
			{title: "Realizar Backup", cmd: "lumina db backup", action: actDBBackup},
			{title: "Realizar Restore", cmd: "lumina db restore", action: actDBRestore},
			{title: "Remover Banco", cmd: "lumina db remove", action: actDBRemove},
			{title: "Otimizar", cmd: "lumina db optimize", action: actDBOptimize},
			{title: "Otimizar para Moodle", cmd: "lumina db moodle", action: actDBMoodle},
		},
	},
	{
		label: "Gerenciar Repositórios",
		items: []submenuEntry{
			{title: "Aplicar Identidade Global", cmd: "lumina repo global", action: actRepoGlobal},
			{title: "Iniciar repositório nessa pasta", cmd: "lumina repo init", action: actRepoInit},
			{title: "Clonar Repositório", cmd: "lumina repo clone", action: actRepoClone},
			{title: "Aplicar identificação", cmd: "lumina repo ident", action: actRepoIdent},
			{title: "Criar/Atualizar .gitignore", cmd: "lumina repo gitignore", action: actGitignore},
			{title: "Criar Código de Conduta", cmd: "lumina repo conduct", action: actRepoConduct},
		},
	},
	{
		label: "Gerenciar Contextos IA",
		items: []submenuEntry{
			{title: "Aplicar/Atualizar contextos de IA", cmd: "lumina ai context", action: actAIContext},
			{title: "Remover contextos de IA", cmd: "lumina ai clear", action: actAIContextRemove},
		},
	},
}

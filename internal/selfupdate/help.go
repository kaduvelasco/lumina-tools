package selfupdate

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// ShowHelp prints the full command and menu reference for Lumina Tools.
func ShowHelp(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Ajuda — Lumina Tools")

	ui.Info(stdout, "Versão: "+version.Version)

	ui.Info(stdout, "COMANDOS (terminal)\n"+
		"  lumina                   Abre a interface TUI interativa\n"+
		"  lumina self-update       Verifica e instala atualizações\n"+
		"  lumina self-uninstall    Remove o binário e as configurações\n"+
		"  lumina version           Exibe a versão instalada\n"+
		"  lumina help              Exibe esta referência")

	ui.Info(stdout, "ATALHOS (TUI)\n"+
		"  ↑ / k          Mover cursor para cima\n"+
		"  ↓ / j          Mover cursor para baixo\n"+
		"  Enter / Espaço Confirmar / selecionar\n"+
		"  t              Seletor de temas (preview ao vivo)\n"+
		"  Esc            Voltar ao menu anterior\n"+
		"  q / Ctrl+C     Sair")

	ui.Info(stdout, "MENUS\n"+
		"  Gerenciamento Linux\n"+
		"    lumina system pos [mint|zorin|ubuntu|fedora]\n"+
		"    lumina system fonts\n"+
		"    lumina system templates\n"+
		"    lumina system apps install\n"+
		"    lumina system apps uninstall\n"+
		"    lumina system update\n"+
		"    lumina system ulauncher\n"+
		"\n"+
		"  DevStack\n"+
		"    lumina stack config [pre|docker|workspace|stack]\n"+
		"    lumina stack start\n"+
		"    lumina stack end\n"+
		"    lumina stack log\n"+
		"    lumina stack status\n"+
		"    lumina stack db\n"+
		"    lumina stack fix-perm\n"+
		"\n"+
		"  DevStuff\n"+
		"    lumina dev pre\n"+
		"    lumina dev llm\n"+
		"    lumina dev ide\n"+
		"    lumina dev term\n"+
		"    lumina dev mcp\n"+
		"    lumina dev update\n"+
		"\n"+
		"  DevManager\n"+
		"    lumina ai\n"+
		"    lumina gitignore\n"+
		"    lumina db [backup|restore|remove|optimize|moodle]\n"+
		"    lumina repo [global|init|clone|ident]\n"+
		"\n"+
		"  Configurações Lumina\n"+
		"    lumina self-update\n"+
		"    lumina self-uninstall\n"+
		"    lumina help")

	ui.Info(stdout, "CONFIGURAÇÃO\n"+
		"  Arquivo  : ~/.lumina/config.yaml\n"+
		"\n"+
		"  Campos:\n"+
		"    workspace_path     – Caminho do workspace de desenvolvimento\n"+
		"    docker_compose_dir – Diretório do docker-compose.yml\n"+
		"    theme              – Tema: Lumina | Claro | Dracula | Nord | Tokyo Night | Gruvbox\n"+
		"    flatpak_scope      – Escopo Flatpak: system (padrão) | user\n"+
		"\n"+
		"  Comandos:\n"+
		"    lumina set workspace <caminho>\n"+
		"    lumina set docker <caminho>\n"+
		"    lumina set theme [lumina|light|dracula|nord|tokyo|gruvbox]\n"+
		"    lumina set flatpak [user|system]")

	ui.WaitEnter(stdout)
	return nil
}

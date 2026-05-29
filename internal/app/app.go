package app

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	"github.com/kaduvelasco/lumina-tools/internal/tui"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

// Run is the single entry point called by main.
// With no args it launches the interactive TUI.
// With args it dispatches to the appropriate domain function.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return tui.Run(ctx, stdin, stdout, stderr)
	}
	return dispatch(ctx, args, stdin, stdout, stderr)
}

func dispatch(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "lumina %s\n", version.Version)
		return nil

	case "update", "self-update":
		exe := executor.New(stdout, stderr)
		return selfupdate.Run(ctx, exe, stdout)

	case "self-uninstall":
		exe := executor.New(stdout, stderr)
		err := selfupdate.Uninstall(ctx, exe, stdout)
		if errors.Is(err, selfupdate.ErrUninstalled) {
			return nil
		}
		return err

	case "-h", "--help", "help":
		printHelp(stdout)
		return nil

	case "system":
		return dispatchSystem(ctx, args[1:], stdin, stdout, stderr)

	case "stack":
		return dispatchStack(ctx, args[1:], stdin, stdout, stderr)

	case "dev":
		return dispatchDev(ctx, args[1:], stdin, stdout, stderr)

	case "ai":
		exe := executor.New(stdout, stderr)
		return managerai.GenerateContext(ctx, exe, stdin, stdout)

	case "gitignore":
		exe := executor.New(stdout, stderr)
		return managergitignore.Generate(ctx, exe, stdout)

	case "db":
		return dispatchDB(ctx, args[1:], stdout, stderr)

	case "repo":
		return dispatchRepo(ctx, args[1:], stdout, stderr)

	case "set":
		return dispatchSet(args[1:], stdout)

	default:
		return fmt.Errorf("subcomando desconhecido: %s\n\nExecute 'lumina help' para ver os comandos disponíveis.", args[0])
	}
}

func dispatchSystem(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina system <pos|gnome|fonts|templates|apps|update|ulauncher>")
	}
	exe := executor.New(stdout, stderr)
	switch args[0] {
	case "pos":
		if len(args) < 2 {
			return tui.RunAtSystemPostInstall(ctx, stdin, stdout, stderr)
		}
		switch args[1] {
		case "mint":
			return postinstall.Mint(ctx, exe, stdout)
		case "zorin":
			return postinstall.Zorin(ctx, exe, stdout)
		case "ubuntu":
			return postinstall.Ubuntu(ctx, exe, stdout)
		case "fedora":
			return postinstall.Fedora(ctx, exe, stdout)
		default:
			return fmt.Errorf("distro desconhecida: %s\nuso: lumina system pos [mint|zorin|ubuntu|fedora]", args[1])
		}
	case "gnome":
		return dispatchGnome(ctx, args[1:], stdin, stdout, stderr)
	case "fonts":
		return fonts.Select(ctx, exe, stdin, stdout)
	case "templates":
		return templates.Select(ctx, exe, stdin, stdout)
	case "apps":
		if len(args) < 2 {
			return fmt.Errorf("uso: lumina system apps <install|uninstall|webapps> [flatpak|ulauncher]")
		}
		switch args[1] {
		case "install":
			if len(args) > 2 && args[2] == "ulauncher" {
				return ulauncher.Install(ctx, exe, stdout)
			}
			return apps.SelectInstall(ctx, exe, stdin, stdout)
		case "uninstall":
			if len(args) > 2 && args[2] == "ulauncher" {
				return ulauncher.Uninstall(ctx, exe, stdout)
			}
			return apps.SelectUninstall(ctx, exe, stdin, stdout)
		case "webapps":
			return apps.ShowWebApps(ctx, exe, stdout)
		default:
			return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina system apps <install|uninstall|webapps> [flatpak|ulauncher]", args[1])
		}
	case "update":
		return update.Run(ctx, exe, stdout)
	case "ulauncher":
		if len(args) > 1 && args[1] == "uninstall" {
			return ulauncher.Uninstall(ctx, exe, stdout)
		}
		return ulauncher.Install(ctx, exe, stdout)
	default:
		return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina system <pos|gnome|fonts|templates|apps|update|ulauncher>", args[0])
	}
}

func dispatchGnome(ctx context.Context, args []string, stdin io.Reader, stdout, _ io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina system gnome <pre|ext|themes|icons|cursors|flatpak>")
	}
	exe := executor.New(stdout, stdout)
	switch args[0] {
	case "pre":
		return gnome.InstallPrereqs(ctx, exe, stdout)
	case "ext":
		return gnome.ShowExtensions(ctx, exe, stdout)
	case "themes":
		return gnome.ManageThemes(ctx, exe, stdin, stdout)
	case "icons":
		return gnome.ManageIcons(ctx, exe, stdin, stdout)
	case "cursors":
		return gnome.ManageCursors(ctx, exe, stdin, stdout)
	case "flatpak":
		return gnome.ApplyFlatpakTheme(ctx, exe, stdin, stdout)
	default:
		return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina system gnome <pre|ext|themes|icons|cursors|flatpak>", args[0])
	}
}

func dispatchStack(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina stack <config|start|end|log|status|db|fix-perm>")
	}
	exe := executor.New(stdout, stderr)
	switch args[0] {
	case "config":
		if len(args) < 2 {
			return tui.RunAtStackConfig(ctx, stdin, stdout, stderr)
		}
		switch args[1] {
		case "pre":
			return stackconfig.Depends(ctx, exe, stdout)
		case "docker":
			return stackconfig.Docker(ctx, exe, stdout)
		case "workspace":
			return stackconfig.Workspace(ctx, exe, stdout)
		case "stack":
			return stackconfig.Compose(ctx, exe, stdout)
		default:
			return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina stack config [pre|docker|workspace|stack]", args[1])
		}
	default:
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("carregar config: %w", err)
		}
		switch args[0] {
		case "start":
			return stack.Start(ctx, exe, stdout, cfg.DockerComposeDir)
		case "end":
			return stack.Stop(ctx, exe, stdout, cfg.DockerComposeDir)
		case "log":
			return stack.Logs(ctx, exe, stdout, cfg.DockerComposeDir)
		case "status":
			return stack.Stats(ctx, exe, stdout)
		case "db":
			return stack.DBInfo(ctx, exe, stdout)
		case "fix-perm":
			return stack.FixPerms(ctx, exe, stdout, cfg.WorkspacePath)
		default:
			return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina stack <config|start|end|log|status|db|fix-perm>", args[0])
		}
	}
}

func dispatchDev(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina dev <pre|llm|ide|term|mcp|update>")
	}
	exe := executor.New(stdout, stderr)
	switch args[0] {
	case "pre":
		return depends.Install(ctx, exe, stdout)
	case "llm":
		return llm.Select(ctx, exe, stdin, stdout)
	case "ide":
		return ide.Select(ctx, exe, stdin, stdout)
	case "term":
		return devterminal.Select(ctx, exe, stdin, stdout)
	case "mcp":
		return mcp.Select(ctx, exe, stdin, stdout)
	case "update":
		return upgrade.Update(ctx, exe, stdout)
	default:
		return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina dev <pre|llm|ide|term|mcp|update>", args[0])
	}
}

func dispatchDB(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina db <backup|restore|remove|optimize|moodle>")
	}
	exe := executor.New(stdout, stderr)
	switch args[0] {
	case "backup":
		return managerdb.Backup(ctx, exe, stdout)
	case "restore":
		return managerdb.Restore(ctx, exe, stdout)
	case "remove":
		return managerdb.Remove(ctx, exe, stdout)
	case "optimize":
		return managerdb.Optimize(ctx, exe, stdout)
	case "moodle":
		return managerdb.OptimizeMoodle(ctx, exe, stdout)
	default:
		return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina db <backup|restore|remove|optimize|moodle>", args[0])
	}
}

func dispatchRepo(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina repo <global|init|clone|ident>")
	}
	exe := executor.New(stdout, stderr)
	switch args[0] {
	case "global":
		return managerrepo.ConfigureGlobal(ctx, exe, stdout)
	case "init":
		return managerrepo.Init(ctx, exe, stdout)
	case "clone":
		return managerrepo.Clone(ctx, exe, stdout)
	case "ident":
		return managerrepo.ApplyIdent(ctx, exe, stdout)
	default:
		return fmt.Errorf("subcomando desconhecido: %s\nuso: lumina repo <global|init|clone|ident>", args[0])
	}
}

var themeAliases = map[string]string{
	"lumina":  "Lumina",
	"light":   "Claro",
	"dracula": "Dracula",
	"nord":    "Nord",
	"tokyo":   "Tokyo Night",
	"gruvbox": "Gruvbox",
}

func dispatchSet(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: lumina set <workspace|docker|theme|flatpak> <valor>")
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregar config: %w", err)
	}
	switch args[0] {
	case "workspace":
		if len(args) < 2 {
			return fmt.Errorf("uso: lumina set workspace <caminho>")
		}
		path, err := config.ExpandPath(args[1])
		if err != nil {
			return err
		}
		cfg.WorkspacePath = path
	case "docker":
		if len(args) < 2 {
			return fmt.Errorf("uso: lumina set docker <caminho>")
		}
		path, err := config.ExpandPath(args[1])
		if err != nil {
			return err
		}
		cfg.DockerComposeDir = path
	case "theme":
		if len(args) < 2 {
			return fmt.Errorf("uso: lumina set theme [lumina|light|dracula|nord|tokyo|gruvbox]")
		}
		name, ok := themeAliases[args[1]]
		if !ok {
			return fmt.Errorf("tema desconhecido: %s\nopções: lumina|light|dracula|nord|tokyo|gruvbox", args[1])
		}
		cfg.Theme = name
	case "flatpak":
		if len(args) < 2 {
			return fmt.Errorf("uso: lumina set flatpak [user|system]")
		}
		if args[1] != "user" && args[1] != "system" {
			return fmt.Errorf("escopo desconhecido: %s\nopções: user|system", args[1])
		}
		cfg.FlatpakScope = args[1]
	default:
		return fmt.Errorf("campo desconhecido: %s\nuso: lumina set <workspace|docker|theme|flatpak> <valor>", args[0])
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("salvar config: %w", err)
	}
	fmt.Fprintln(stdout, "Configuração atualizada.")
	return nil
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `lumina — Ferramentas Linux unificadas

COMANDOS
  lumina                   Abre a interface TUI interativa
  lumina self-update       Verifica e instala atualizações
  lumina self-uninstall    Remove o binário e as configurações
  lumina version           Exibe a versão instalada
  lumina help              Exibe esta referência

GERENCIAMENTO LINUX
  lumina system pos [mint|zorin|ubuntu|fedora]
  lumina system fonts
  lumina system templates
  lumina system apps install
  lumina system apps uninstall
  lumina system apps webapps
  lumina system update
  lumina system ulauncher
  lumina system ulauncher uninstall
  lumina system gnome [pre|ext|themes|icons|cursors|flatpak]

DEVSTACK
  lumina stack config [pre|docker|workspace|stack]
  lumina stack start
  lumina stack end
  lumina stack log
  lumina stack status
  lumina stack db
  lumina stack fix-perm

DEVSTUFF
  lumina dev pre
  lumina dev llm
  lumina dev ide
  lumina dev term
  lumina dev mcp
  lumina dev update

DEVMANAGER
  lumina ai
  lumina gitignore
  lumina db [backup|restore|remove|optimize|moodle]
  lumina repo [global|init|clone|ident]

CONFIGURAÇÃO
  lumina set workspace <caminho>
  lumina set docker <caminho>
  lumina set theme [lumina|light|dracula|nord|tokyo|gruvbox]
  lumina set flatpak [user|system]`)
}

package mcp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install shows available MCP servers and installs selected ones.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Servidores MCP :: Instalar")

	servers, err := Catalogue()
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	if len(servers) == 0 {
		ui.Warning(stdout, "Catálogo vazio. Adicione servidores em assets/mcp/servers.yaml.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Verificando servidores MCP instalados...")
	installed := InstalledMap(ctx, exe, servers)

	for i, s := range servers {
		status := ""
		if installed[s.Name] {
			status = " (instalado)"
		}
		fmt.Fprintf(stdout, "  %d. %s%s\n", i+1, s.Name, status)
		if s.Description != "" {
			fmt.Fprintf(stdout, "     %s\n", s.Description)
		}
	}

	fmt.Fprintln(stdout, "\nDigite os números para instalar, ou Enter para cancelar:")
	fmt.Fprint(stdout, "> ")

	line := prompt.ReadLine()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	selected := prompt.ParseSelection(line, len(servers))
	if len(selected) == 0 {
		ui.Warning(stdout, "Nenhuma seleção válida.")
		ui.WaitEnter(stdout)
		return nil
	}

	for _, idx := range selected {
		s := servers[idx]
		ui.Info(stdout, "Instalando "+s.Name+"...")
		if err := npmInstallMCP(ctx, exe, stdout, s.Package); err != nil {
			ui.Warning(stdout, "Falha: "+err.Error())
		} else {
			ui.Success(stdout, s.Name+" instalado.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

// npmInstallMCP installs an npm package globally, sourcing nvm when available.
// Global npm installs via nvm are user-local and do not require sudo.
func npmInstallMCP(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkg string) error {
	script := fmt.Sprintf(`
set -e
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
npm install -g %s
`, pkg)
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
}

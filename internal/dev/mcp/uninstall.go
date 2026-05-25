package mcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Uninstall shows installed MCP servers and removes selected ones.
func Uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Servidores MCP :: Desinstalar")

	servers, err := Catalogue()
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	if len(servers) == 0 {
		ui.Warning(stdout, "Catálogo vazio. Nenhum servidor para remover.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Verificando servidores MCP instalados...")
	installed := InstalledMap(ctx, exe, servers)

	var present []Server
	for _, s := range servers {
		if installed[s.Name] {
			present = append(present, s)
		}
	}

	if len(present) == 0 {
		ui.Info(stdout, "Nenhum servidor MCP instalado.")
		ui.WaitEnter(stdout)
		return nil
	}

	for i, s := range present {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, s.Name)
	}

	fmt.Fprintln(stdout, "\nDigite os números para desinstalar, ou Enter para cancelar:")
	fmt.Fprint(stdout, "> ")

	line := prompt.ReadLine()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	selected := prompt.ParseSelection(line, len(present))
	if len(selected) == 0 {
		ui.Warning(stdout, "Nenhuma seleção válida.")
		ui.WaitEnter(stdout)
		return nil
	}

	for _, idx := range selected {
		s := present[idx]
		ui.Info(stdout, "Desinstalando "+s.Name+"...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"env", "PATH="+os.Getenv("PATH"), "npm", "uninstall", "-g", s.Package,
		); err != nil {
			ui.Warning(stdout, "Falha: "+err.Error())
		} else {
			ui.Success(stdout, s.Name+" removido.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

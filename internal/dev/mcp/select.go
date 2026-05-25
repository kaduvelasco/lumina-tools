package mcp

import (
	"context"
	"io"
	"os"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Select shows a multiselect of all MCP servers from the embedded catalogue.
// Items already installed start selected. Deselecting one uninstalls it;
// selecting one installs it.
func Select(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Servidores MCP")

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

	items := make([]ui.SelectItem, len(servers))
	for i, s := range servers {
		label := s.Name
		if s.Description != "" {
			label += " — " + s.Description
		}
		items[i] = ui.SelectItem{Label: label, ID: s.Name, Selected: installed[s.Name]}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var toInstall, toRemove []Server
	for i, item := range finalItems {
		s := servers[i]
		switch {
		case item.Selected && !installed[s.Name]:
			toInstall = append(toInstall, s)
		case !item.Selected && installed[s.Name]:
			toRemove = append(toRemove, s)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Servidores MCP")
	npmPath := os.Getenv("PATH")

	for _, s := range toRemove {
		ui.Info(stdout, "Desinstalando "+s.Name+"...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"env", "PATH="+npmPath, "npm", "uninstall", "-g", s.Package,
		); err != nil {
			ui.Warning(stdout, "Falha ao remover "+s.Name+": "+err.Error())
		} else {
			ui.Success(stdout, s.Name+" removido.")
		}
	}

	for _, s := range toInstall {
		ui.Info(stdout, "Instalando "+s.Name+"...")
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"env", "PATH="+npmPath, "npm", "install", "-g", s.Package,
		); err != nil {
			ui.Warning(stdout, "Falha ao instalar "+s.Name+": "+err.Error())
		} else {
			ui.Success(stdout, s.Name+" instalado.")
		}
	}

	ui.Success(stdout, "Gerenciamento de servidores MCP concluído.")
	ui.WaitEnter(stdout)
	return nil
}

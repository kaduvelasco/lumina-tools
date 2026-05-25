package upgrade

import (
	"context"
	"fmt"
	"io"
	"strings"

	devide "github.com/kaduvelasco/lumina-tools/internal/dev/ide"
	devllm "github.com/kaduvelasco/lumina-tools/internal/dev/llm"
	devmcp "github.com/kaduvelasco/lumina-tools/internal/dev/mcp"
	devterminal "github.com/kaduvelasco/lumina-tools/internal/dev/terminal"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Update lists all installed dev tools and offers to update them to their latest versions.
func Update(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Atualizar Ferramentas")
	ui.Info(stdout, "Verificando ferramentas instaladas...")

	llmInstalled := devllm.InstalledMap(ctx, exe)
	ideInstalled := devide.InstalledMap(ctx, exe)
	termInstalled := devterminal.InstalledMap(ctx, exe)
	mcpServers, _ := devmcp.Catalogue()
	mcpInstalled := devmcp.InstalledMap(ctx, exe, mcpServers)

	var lines []string
	for _, l := range devllm.Catalogue {
		if llmInstalled[l.Name] {
			lines = append(lines, "  CLI       "+l.Name)
		}
	}
	for _, e := range devide.Catalogue {
		if ideInstalled[e.Name] {
			lines = append(lines, "  IDE       "+e.Name)
		}
	}
	for _, t := range devterminal.Catalogue {
		if termInstalled[t.Name] {
			lines = append(lines, "  Terminal  "+t.Name)
		}
	}
	for _, s := range mcpServers {
		if mcpInstalled[s.Name] {
			lines = append(lines, "  MCP       "+s.Name)
		}
	}

	if len(lines) == 0 {
		ui.Info(stdout, "Nenhuma ferramenta instalada.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.Info(stdout, "Ferramentas instaladas:\n"+strings.Join(lines, "\n"))

	fmt.Fprint(stdout, "\nAtualizar todas as ferramentas listadas? (s/N): ")
	if c := strings.TrimSpace(prompt.ReadLine()); c != "s" && c != "S" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "DevStuff :: Atualizar Ferramentas")

	if err := devllm.Update(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Erro ao atualizar CLIs: "+err.Error())
	}
	if err := devide.Update(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Erro ao atualizar IDEs: "+err.Error())
	}
	if err := devterminal.Update(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Erro ao atualizar Terminais: "+err.Error())
	}
	if err := devmcp.Update(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Erro ao atualizar Servidores MCP: "+err.Error())
	}

	ui.Success(stdout, "Atualização de ferramentas concluída.")
	ui.WaitEnter(stdout)
	return nil
}

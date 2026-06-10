package selfupdate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Configure lets the user view and update Lumina's settings interactively.
func Configure(ctx context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Configurar Lumina")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar configuração: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.PrintBox(stdout, fmt.Sprintf(
		"Workspace:       %s\nDocker Compose:  %s\nTema:            %s\nEscopo Flatpak:  %s",
		display(cfg.WorkspacePath),
		display(cfg.DockerComposeDir),
		cfg.Theme,
		cfg.FlatpakScope,
	))

	// Workspace path
	fmt.Fprintf(stdout, "\n  Workspace [%s]: ", display(cfg.WorkspacePath))
	if val, _ := prompt.ReadLineFrom(stdin); strings.TrimSpace(val) != "" {
		if expanded, err := config.ExpandPath(strings.TrimSpace(val)); err != nil {
			ui.Warning(stdout, "Caminho inválido, mantendo o anterior: "+err.Error())
		} else {
			cfg.WorkspacePath = expanded
		}
	}

	// Docker Compose Dir
	fmt.Fprintf(stdout, "  Docker Compose [%s]: ", display(cfg.DockerComposeDir))
	if val, _ := prompt.ReadLineFrom(stdin); strings.TrimSpace(val) != "" {
		if expanded, err := config.ExpandPath(strings.TrimSpace(val)); err != nil {
			ui.Warning(stdout, "Caminho inválido, mantendo o anterior: "+err.Error())
		} else {
			cfg.DockerComposeDir = expanded
		}
	}

	// Theme
	ui.Info(stdout, "Tema — selecione ou Esc para manter o atual:")
	themeItems := []ui.SelectItem{
		{Label: "Lumina", ID: "Lumina"},
		{Label: "Claro", ID: "Claro"},
		{Label: "Dracula", ID: "Dracula"},
		{Label: "Nord", ID: "Nord"},
		{Label: "Tokyo Night", ID: "Tokyo Night"},
		{Label: "Gruvbox", ID: "Gruvbox"},
	}
	for i := range themeItems {
		if themeItems[i].ID == cfg.Theme {
			themeItems[i].Label += " ✓"
			break
		}
	}
	if idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, themeItems); err != nil {
		return err
	} else if ok && idx >= 0 {
		cfg.Theme = themeItems[idx].ID
	}

	// Flatpak scope
	ui.Info(stdout, "Escopo Flatpak — selecione ou Esc para manter o atual:")
	scopeItems := []ui.SelectItem{
		{Label: "user", ID: "user"},
		{Label: "system", ID: "system"},
	}
	for i := range scopeItems {
		if scopeItems[i].ID == cfg.FlatpakScope {
			scopeItems[i].Label += " ✓"
			break
		}
	}
	if idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, scopeItems); err != nil {
		return err
	} else if ok && idx >= 0 {
		cfg.FlatpakScope = scopeItems[idx].ID
	}

	if err := config.Save(cfg); err != nil {
		ui.Err(stdout, "Falha ao salvar configuração: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Configuração salva com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

func display(s string) string {
	if s == "" {
		return "(não definido)"
	}
	return s
}

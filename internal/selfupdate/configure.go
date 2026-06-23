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

// distroLabels maps a saved cfg.Distro token to its human-readable label,
// shown in the settings summary and as the picker's default selection.
var distroLabels = map[string]string{
	"mint":   "Linux Mint",
	"zorin":  "Zorin OS",
	"ubuntu": "Ubuntu / Kubuntu",
	"fedora": "Fedora",
	"outro":  "Outro",
}

// deLabels maps a saved cfg.DE token to its human-readable label.
var deLabels = map[string]string{
	"cinnamon": "Cinnamon",
	"gnome":    "GNOME",
	"other":    "Outro",
}

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
		"Workspace:       %s\nDocker Compose:  %s\nTema:            %s\nEscopo Flatpak:  %s\nDistribuição:    %s\nAmbiente:        %s",
		display(cfg.WorkspacePath),
		display(cfg.DockerComposeDir),
		cfg.Theme,
		cfg.FlatpakScope,
		display(distroLabels[cfg.Distro]),
		display(deLabels[cfg.DE]),
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

	// Distribuição
	ui.Info(stdout, "Distribuição — selecione ou Esc para manter a atual:")
	distroItems := []ui.SelectItem{
		{Label: "Linux Mint", ID: "mint"},
		{Label: "Zorin OS", ID: "zorin"},
		{Label: "Ubuntu / Kubuntu", ID: "ubuntu"},
		{Label: "Fedora", ID: "fedora"},
		{Label: "Outro", ID: "outro"},
	}
	for i := range distroItems {
		if distroItems[i].ID == cfg.Distro {
			distroItems[i].Label += " ✓"
			break
		}
	}
	if idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, distroItems); err != nil {
		return err
	} else if ok && idx >= 0 {
		cfg.Distro = distroItems[idx].ID
	}

	// Ambiente de desktop
	ui.Info(stdout, "Ambiente de desktop — selecione ou Esc para manter o atual:")
	deItems := []ui.SelectItem{
		{Label: "Cinnamon", ID: "cinnamon"},
		{Label: "GNOME", ID: "gnome"},
		{Label: "Outro", ID: "other"},
	}
	for i := range deItems {
		if deItems[i].ID == cfg.DE {
			deItems[i].Label += " ✓"
			break
		}
	}
	if idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, deItems); err != nil {
		return err
	} else if ok && idx >= 0 {
		cfg.DE = deItems[idx].ID
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

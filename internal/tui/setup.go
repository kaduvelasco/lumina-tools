package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var distroNames = map[string]string{
	"mint":   "Linux Mint",
	"zorin":  "Zorin OS",
	"ubuntu": "Ubuntu",
	"fedora": "Fedora",
	"outro":  "Outro",
}

// distroDisplayName returns the human-readable label for a distro token.
// Returns "" for unknown tokens (header omits the segment when empty).
func distroDisplayName(d string) string {
	return distroNames[d]
}

var deNames = map[string]string{
	"cinnamon": "Cinnamon",
	"gnome":    "GNOME",
	"other":    "Outro",
}

// ensureSystemInfo checks if distro and DE are recorded in cfg.
// If either is missing, detects the values, asks the user to confirm, and
// saves the result before the TUI starts.
func ensureSystemInfo(ctx context.Context, stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	if cfg.Distro != "" && cfg.DE != "" {
		return nil
	}

	detectedDistro := distro.NormalizedDistro()
	detectedDE := distro.DetectDE()

	dName := distroDisplayName(detectedDistro)
	if dName == "" {
		dName = "Desconhecida"
	}
	deName := deNames[detectedDE]
	if deName == "" {
		deName = "Outro"
	}

	fmt.Fprintln(stdout)
	ui.Info(stdout, fmt.Sprintf("Distribuição detectada: %s  |  Ambiente: %s", dName, deName))
	fmt.Fprintf(stdout, "\nEssas informações estão corretas? (s/N): ")

	answer, _ := prompt.ReadLineFrom(stdin)
	answer = strings.ToLower(strings.TrimSpace(answer))
	confirmed := answer == "s" || answer == "sim"

	// Force manual selection when: user denied auto-detection OR distro was not recognised.
	if !confirmed || detectedDistro == "" {
		if confirmed && detectedDistro == "" {
			ui.Warning(stdout, "Distribuição não identificada automaticamente. Selecione manualmente:")
		}

		distroItems := []ui.SelectItem{
			{Label: "Linux Mint", ID: "mint"},
			{Label: "Zorin OS", ID: "zorin"},
			{Label: "Ubuntu / Kubuntu", ID: "ubuntu"},
			{Label: "Fedora", ID: "fedora"},
			{Label: "Outro", ID: "outro"},
		}
		fmt.Fprintln(stdout, "\nSelecione a distribuição:")
		idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, distroItems)
		if err != nil {
			return fmt.Errorf("selecionar distribuição: %w", err)
		}
		switch {
		case ok && idx >= 0:
			detectedDistro = distroItems[idx].ID
		case detectedDistro == "":
			// User cancelled and auto-detection found nothing usable — fall back
			// to "outro" instead of leaving the field empty, otherwise this
			// prompt would resurface on every single startup from now on.
			detectedDistro = "outro"
		}

		deItems := []ui.SelectItem{
			{Label: "Cinnamon", ID: "cinnamon"},
			{Label: "GNOME", ID: "gnome"},
			{Label: "Outro", ID: "other"},
		}
		fmt.Fprintln(stdout, "\nSelecione o ambiente de desktop:")
		idx, ok, err = ui.RunSingleSelect(ctx, stdin, stdout, deItems)
		if err != nil {
			return fmt.Errorf("selecionar ambiente: %w", err)
		}
		if ok && idx >= 0 {
			detectedDE = deItems[idx].ID
		}
		// If cancelled, detectedDE keeps the auto-detected value — DetectDE
		// never returns "", so DE is always saved with something usable.
	}

	cfg.Distro = detectedDistro
	cfg.DE = detectedDE
	return config.Save(cfg)
}

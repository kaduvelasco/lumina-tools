package gnome

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type extension struct {
	Name        string
	Description string
	URL         string
}

var recommendedExtensions = []extension{
	{
		Name:        "Tiling Shell",
		Description: "Gerenciamento de janelas em mosaico para GNOME",
		URL:         "https://extensions.gnome.org/extension/7065/tiling-shell/",
	},
	{
		Name:        "User Themes",
		Description: "Permite aplicar temas de shell personalizados via GNOME Tweaks",
		URL:         "https://extensions.gnome.org/extension/19/user-themes/",
	},
	{
		Name:        "ArcMenu",
		Description: "Menu de aplicativos alternativo com múltiplos layouts",
		URL:         "https://extensions.gnome.org/extension/3628/arcmenu/",
	},
	{
		Name:        "Dash to Panel",
		Description: "Combina a dock e a barra superior em um único painel",
		URL:         "https://extensions.gnome.org/extension/1160/dash-to-panel/",
	},
}

// ShowExtensions displays recommended GNOME extensions with installation links.
func ShowExtensions(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Customizar GNOME — Extensões Recomendadas")

	if !isGnome() {
		ui.Err(stdout, ErrNotGnome.Error())
		ui.WaitEnter(stdout)
		return nil
	}

	fmt.Fprintln(stdout)
	ui.Info(stdout, "Instale as extensões pelo Gerenciador de Extensões ou pelos links abaixo.")
	fmt.Fprintln(stdout)

	for _, ext := range recommendedExtensions {
		fmt.Fprintf(stdout, "  ► %-20s %s\n", ext.Name, ext.Description)
		fmt.Fprintf(stdout, "    %s\n\n", ext.URL)
	}

	ui.Info(stdout, "Dica: instale os Pré-requisitos primeiro para ter o Gerenciador de Extensões.")
	ui.WaitEnter(stdout)
	return nil
}

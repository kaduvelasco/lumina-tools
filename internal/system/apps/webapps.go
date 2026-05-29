package apps

import (
	"context"
	"fmt"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type webApp struct {
	Name string
	URL  string
}

var suggestedWebApps = []webApp{
	{Name: "WhatsApp", URL: "https://web.whatsapp.com"},
	{Name: "Vectorpea", URL: "https://www.vectorpea.com/"},
}

// ShowWebApps displays suggested web apps with their URLs.
func ShowWebApps(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Aplicativos — WebApps Sugeridos")

	fmt.Fprintln(stdout)
	ui.Info(stdout, "Abra os links abaixo no navegador e use 'Instalar como aplicativo' ou 'Criar atalho'.")
	fmt.Fprintln(stdout)

	for _, app := range suggestedWebApps {
		fmt.Fprintf(stdout, "  ► %-20s %s\n\n", app.Name, app.URL)
	}

	ui.WaitEnter(stdout)
	return nil
}

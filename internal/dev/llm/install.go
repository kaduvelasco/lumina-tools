package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install shows available LLMs and installs selected ones.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "LLMs :: Instalar")
	ui.Info(stdout, "Verificando LLMs instalados...")

	installed := InstalledMap(ctx, exe)

	for i, l := range Catalogue {
		status := ""
		if installed[l.Name] {
			status = " (instalado)"
		}
		fmt.Fprintf(stdout, "  %d. %s%s\n", i+1, l.Name, status)
	}

	fmt.Fprintln(stdout, "\nDigite os números para instalar (ex: 1 3), ou Enter para cancelar:")
	fmt.Fprint(stdout, "> ")

	line := prompt.ReadLine()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	selected := prompt.ParseSelection(line, len(Catalogue))
	if len(selected) == 0 {
		ui.Warning(stdout, "Nenhuma seleção válida.")
		ui.WaitEnter(stdout)
		return nil
	}

	hasNode := nodeAvailable(ctx, exe)
	if !hasNode {
		ui.Info(stdout, "Node.js não encontrado. Instalando via nvm...")
		if err := installNode(ctx, exe, stdout); err != nil {
			ui.Err(stdout, "Falha ao instalar Node.js: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("instalar node: %w", err)
		}
	}

	for _, idx := range selected {
		l := Catalogue[idx]
		if installed[l.Name] {
			ui.Info(stdout, l.Name+" já instalado. Atualizando...")
		} else {
			ui.Info(stdout, "Instalando "+l.Name+"...")
		}
		if err := installOne(ctx, exe, stdout, l); err != nil {
			ui.Warning(stdout, "Falha em "+l.Name+": "+err.Error())
		} else {
			ui.Success(stdout, l.Name+" instalado.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func installOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, l LLM) error {
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	switch l.Cmd {
	case "claude":
		script := `curl -fsSL https://claude.ai/install.sh | bash`
		return exe.Run(ctx, opts, "bash", "-c", script)
	case "agy":
		script := `curl -fsSL https://antigravity.google/cli/install.sh | bash`
		return exe.Run(ctx, opts, "bash", "-c", script)
	case "codex":
		return npmInstall(ctx, exe, stdout, "@openai/codex")
	case "opencode":
		return npmInstall(ctx, exe, stdout, "opencode-ai@latest")
	}
	return fmt.Errorf("instalador desconhecido para %s", l.Name)
}

func npmInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkg string) error {
	// Source NVM when available so a freshly-installed Node is on PATH.
	// NVM global installs are user-local and do not require sudo.
	script := fmt.Sprintf(`
set -e
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
npm install -g %s
`, pkg)
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
}

func nodeAvailable(ctx context.Context, exe *executor.Executor) bool {
	_, err := exe.Output(ctx, executor.Options{}, "which", "node")
	return err == nil
}

func installNode(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	script := `
set -e
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
nvm install --lts
nvm use --lts
`
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
}

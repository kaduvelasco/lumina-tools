package localbin

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const exportLine = `export PATH="$HOME/.local/bin:$PATH"`

// EnsureInPath adds $HOME/.local/bin to ~/.bashrc when it is not already
// present in PATH. Several tools (Claude Code, Kitty, Antigravity CLI) install
// their binaries there and require the directory to be on PATH to be found.
func EnsureInPath(stdout io.Writer) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	localBin := filepath.Join(home, ".local", "bin")

	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == localBin {
			return
		}
	}

	bashrc := filepath.Join(home, ".bashrc")
	if data, readErr := os.ReadFile(bashrc); readErr == nil && strings.Contains(string(data), exportLine) {
		return
	}

	f, openErr := os.OpenFile(bashrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		ui.Warning(stdout, "Adicione manualmente ao ~/.bashrc: "+exportLine)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "\n%s\n", exportLine)
}

// RunNPMGlobal runs `npm <action> -g <pkg>`, sourcing nvm when available.
// action must be "install" or "uninstall". Never requires sudo since nvm
// installs are user-local.
func RunNPMGlobal(ctx context.Context, exe *executor.Executor, stdout io.Writer, action, pkg string) error {
	script := fmt.Sprintf(`
set -e
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
if ! command -v npm &>/dev/null; then
    printf 'Node.js não encontrado. Execute DevStuff → Instalar Pré-requisitos primeiro.\n' >&2
    exit 1
fi
npm %s -g %s
`, action, pkg)
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script)
}

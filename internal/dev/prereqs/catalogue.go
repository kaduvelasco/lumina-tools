package prereqs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/dev/localbin"
	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Prereq describes a prerequisite group managed by lumina.
type Prereq struct {
	Name        string
	Description string
	ID          string
}

// Catalogue lists all prerequisite groups managed by lumina.
var Catalogue = []Prereq{
	{Name: "Pacotes base", ID: "base", Description: "curl, git, openssl, lsof"},
	{Name: "Ferramentas DevStuff", ID: "devtools", Description: "libsecret, gnome-keyring"},
	{Name: "GitHub CLI", ID: "gh", Description: "gh — interface de linha de comando para o GitHub"},
	{Name: "Docker Engine", ID: "docker", Description: "Docker Engine + buildx, serviço habilitado"},
	{Name: "Node.js", ID: "node", Description: "Node.js LTS via nvm"},
}

// InstalledMap returns which prerequisite groups are currently installed (by Name).
func InstalledMap(ctx context.Context, exe *executor.Executor) map[string]bool {
	result := make(map[string]bool, len(Catalogue))
	for _, p := range Catalogue {
		result[p.Name] = isInstalled(ctx, exe, p.ID)
	}
	return result
}

func isInstalled(ctx context.Context, exe *executor.Executor, id string) bool {
	which := func(cmd string) bool {
		_, err := exe.Output(ctx, executor.Options{}, "which", cmd)
		return err == nil
	}
	switch id {
	case "base":
		return which("curl")
	case "devtools":
		return which("secret-tool")
	case "gh":
		return which("gh")
	case "docker":
		return which("docker")
	case "node":
		script := `export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"; command -v node && command -v npm`
		_, err := exe.Output(ctx, executor.Options{}, "bash", "-c", script)
		return err == nil
	}
	return false
}

func installOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, p Prereq) error {
	family := distro.Detect()
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch p.ID {
	case "base":
		return distro.InstallPkgs(ctx, exe, stdout, family, "curl", "git", "openssl", "lsof")
	case "devtools":
		return installDevTools(ctx, exe, stdout, family)
	case "gh":
		return installGH(ctx, exe, stdout, family, opts)
	case "docker":
		return installDocker(ctx, exe, stdout, family, opts)
	case "node":
		return installNode(ctx, exe, stdout)
	}
	return fmt.Errorf("instalador desconhecido para %s", p.Name)
}

func uninstallOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, p Prereq) error {
	family := distro.Detect()
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch p.ID {
	case "base":
		return removePkgs(ctx, exe, opts, family, "curl", "git", "openssl", "lsof")
	case "devtools":
		return uninstallDevTools(ctx, exe, opts, family)
	case "gh":
		return uninstallGH(ctx, exe, stdout, family, opts)
	case "docker":
		return removePkgs(ctx, exe, opts, family, dockerPkgs(family)...)
	case "node":
		return uninstallNode(stdout)
	}
	return fmt.Errorf("desinstalador desconhecido para %s", p.Name)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func removePkgs(ctx context.Context, exe *executor.Executor, opts executor.Options, family string, pkgs ...string) error {
	switch family {
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", append([]string{"remove", "-y", "--"}, pkgs...)...)
	default:
		return exe.Run(ctx, opts, "apt-get", append([]string{"purge", "-y", "--"}, pkgs...)...)
	}
}

func dockerPkgs(family string) []string {
	switch family {
	case distro.Fedora:
		return []string{"docker", "docker-compose", "docker-buildx-plugin"}
	default:
		return []string{"docker.io", "docker-compose-v2", "docker-buildx"}
	}
}

// ── devtools ──────────────────────────────────────────────────────────────────

func devToolsPkgs(family string) []string {
	switch family {
	case distro.Debian:
		return []string{"libsecret-1-0", "libsecret-tools", "gnome-keyring"}
	case distro.Fedora:
		return []string{"libsecret", "libsecret-devel", "gnome-keyring"}
	default:
		return []string{"libsecret", "gnome-keyring"}
	}
}

func installDevTools(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string) error {
	return distro.InstallPkgs(ctx, exe, stdout, family, devToolsPkgs(family)...)
}

func uninstallDevTools(ctx context.Context, exe *executor.Executor, opts executor.Options, family string) error {
	return removePkgs(ctx, exe, opts, family, devToolsPkgs(family)...)
}

// ── GitHub CLI ────────────────────────────────────────────────────────────────

func installGH(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, opts executor.Options) error {
	ui.Info(stdout, "Instalando GitHub CLI...")
	switch family {
	case distro.Debian:
		script := `set -e
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list
apt-get update -q
apt-get install -y -- gh`
		return exe.Run(ctx, opts, "bash", "-c", script)
	case distro.Fedora:
		script := `set -e
dnf install -y 'dnf-command(config-manager)'
dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo
dnf install -y -- gh`
		return exe.Run(ctx, opts, "bash", "-c", script)
	default:
		return exe.Run(ctx, opts, "pacman", "-S", "--noconfirm", "--", "github-cli")
	}
}

func uninstallGH(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, opts executor.Options) error {
	_ = stdout
	switch family {
	case distro.Debian:
		script := `set -e
apt-get purge -y -- gh
rm -f /etc/apt/sources.list.d/github-cli.list /usr/share/keyrings/githubcli-archive-keyring.gpg
apt-get update -q`
		return exe.Run(ctx, opts, "bash", "-c", script)
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", "remove", "-y", "--", "gh")
	default:
		return exe.Run(ctx, opts, "pacman", "-R", "--noconfirm", "--", "github-cli")
	}
}

// ── Docker ────────────────────────────────────────────────────────────────────

func installDocker(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, opts executor.Options) error {
	ui.Info(stdout, "Instalando Docker via gerenciador de pacotes...")
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			return err
		}
		if err := exe.Run(ctx, opts, "apt-get", "install", "-y", "--", "docker.io", "docker-compose-v2", "docker-buildx"); err != nil {
			return err
		}
	case distro.Fedora:
		if err := exe.Run(ctx, opts, "dnf", "install", "-y", "--", "docker", "docker-compose", "docker-buildx-plugin"); err != nil {
			return err
		}
	default:
		if err := exe.Run(ctx, opts, "pacman", "-S", "--noconfirm", "--", "docker", "docker-compose"); err != nil {
			return err
		}
	}

	_ = exe.Run(ctx, opts, "systemctl", "enable", "--now", "docker")

	user := executor.CurrentUser()
	if user != "" {
		out, _ := exe.Output(ctx, executor.Options{}, "groups", user)
		if !strings.Contains(out, "docker") {
			ui.Info(stdout, "Adicionando "+user+" ao grupo docker...")
			if err := exe.Run(ctx, opts, "usermod", "-aG", "docker", user); err != nil {
				ui.Warning(stdout, "usermod: "+err.Error())
			} else {
				ui.Warning(stdout, "Reinicie a sessão para aplicar as permissões do grupo docker.")
			}
		}
	}

	if family == distro.Fedora {
		ui.Warning(stdout, "Fedora: se volumes não funcionarem, execute:\n  sudo setsebool -P container_manage_cgroup on")
	}
	return nil
}

// ── Node.js ───────────────────────────────────────────────────────────────────

func installNode(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	checkScript := `export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"; command -v node && command -v npm`
	if _, e := exe.Output(ctx, executor.Options{}, "bash", "-c", checkScript); e == nil {
		ui.Info(stdout, "Node.js já disponível.")
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("obter diretório home: %w", err)
	}
	_, statErr := os.Stat(filepath.Join(home, ".nvm", "nvm.sh"))
	freshNVM := statErr != nil

	script := `set -e
export NVM_DIR="$HOME/.nvm"
if [ ! -s "$NVM_DIR/nvm.sh" ]; then
    curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
    export NVM_DIR="$HOME/.nvm"
fi
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
nvm install --lts
nvm use --lts
`
	ui.Info(stdout, "Instalando Node.js LTS via nvm...")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", script); err != nil {
		return fmt.Errorf("instalar Node.js: %w", err)
	}

	localbin.EnsureInPath(stdout)

	if freshNVM {
		ui.Warning(stdout, "Reinicie o terminal para ativar o nvm e o Node.js.")
	}
	return nil
}

func uninstallNode(stdout io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("obter diretório home: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(home, ".nvm")); err != nil {
		return fmt.Errorf("remover ~/.nvm: %w", err)
	}
	ui.Warning(stdout, "Node.js removido. Remova manualmente as linhas NVM_DIR do ~/.bashrc.")
	return nil
}

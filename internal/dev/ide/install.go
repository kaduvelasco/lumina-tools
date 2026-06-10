package ide

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Install shows available IDEs and installs selected ones.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "IDEs :: Instalar")
	ui.Info(stdout, "Verificando IDEs instalados...")

	installed := InstalledMap(ctx, exe)

	for i, e := range Catalogue {
		status := ""
		if installed[e.Name] {
			status = " (instalado)"
		}
		fmt.Fprintf(stdout, "  %d. %s%s\n", i+1, e.Name, status)
	}

	fmt.Fprintln(stdout, "\nDigite os números para instalar, ou Enter para cancelar:")
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

	family := distro.Detect()

	for _, idx := range selected {
		e := Catalogue[idx]
		ui.Info(stdout, "Instalando "+e.Name+"...")
		if err := installOne(ctx, exe, stdout, e, family); err != nil {
			ui.Warning(stdout, "Falha em "+e.Name+": "+err.Error())
		} else {
			ui.Success(stdout, e.Name+" instalado.")
		}
	}

	ui.WaitEnter(stdout)
	return nil
}

func installOne(ctx context.Context, exe *executor.Executor, stdout io.Writer, e IDE, family string) error {
	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch e.Cmd {
	case "zed":
		script := `curl -fsSL https://zed.dev/install.sh | sh`
		return exe.Run(ctx, opts, "bash", "-c", script)

	case "windsurf":
		return installWindsurf(ctx, exe, stdout, family, sudo)

	case "code":
		return installVSCode(ctx, exe, stdout, family, sudo)

	case "codium":
		return installVSCodium(ctx, exe, stdout, family)

	case "dbeaver":
		return installDBeaver(ctx, exe, stdout, family, sudo)
	}
	return fmt.Errorf("instalador desconhecido para %s", e.Name)
}

func installWindsurf(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, sudo executor.Options) error {
	switch family {
	case distro.Debian:
		script := `
set -e
KEYRING="/etc/apt/keyrings/windsurf-stable.gpg"
SOURCES="/etc/apt/sources.list.d/windsurf.list"
mkdir -p /etc/apt/keyrings
wget -qO- https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/windsurf.gpg \
  | gpg --dearmor | install -D -o root -g root -m 644 /dev/stdin "$KEYRING"
echo "deb [arch=amd64 signed-by=${KEYRING}] https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/apt stable main" \
  | tee "$SOURCES" > /dev/null
apt-get update -qq
apt-get install -y windsurf
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	case distro.Fedora:
		script := `
set -e
rpm --import https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/yum/RPM-GPG-KEY-windsurf
printf '[windsurf]\nname=Windsurf Repository\nbaseurl=https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/yum/repo/\nenabled=1\ngpgcheck=1\ngpgkey=https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/yum/RPM-GPG-KEY-windsurf\n' \
  | tee /etc/yum.repos.d/windsurf.repo > /dev/null
dnf install -y windsurf
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	default:
		return exe.Run(ctx, sudo, "pacman", "-S", "--noconfirm", "windsurf")
	}
}

func installVSCode(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, sudo executor.Options) error {
	switch family {
	case distro.Debian:
		script := `
set -e
KEYRING="/usr/share/keyrings/microsoft-archive-keyring.gpg"
[ -f "$KEYRING" ] || wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | dd of="$KEYRING" status=none
if [ ! -f /etc/apt/sources.list.d/vscode.list ]; then
  echo "deb [arch=amd64,arm64 signed-by=$KEYRING] https://packages.microsoft.com/repos/code stable main" \
    | tee /etc/apt/sources.list.d/vscode.list > /dev/null
fi
apt-get update -qq
apt-get install -y code
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	case distro.Fedora:
		script := `
set -e
rpm --import https://packages.microsoft.com/keys/microsoft.asc
[ -f /etc/yum.repos.d/vscode.repo ] || printf '[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc\n' | tee /etc/yum.repos.d/vscode.repo > /dev/null
dnf install -y code
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	default:
		return fmt.Errorf("instale o VS Code manualmente no Arch via AUR (yay -S visual-studio-code-bin)")
	}
}

func installDBeaver(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, _ executor.Options) error {
	switch family {
	case distro.Debian:
		// BUG-02: guard prevents overwriting the keyring on every reinstall.
		// BUG-03: DEBIAN_FRONTEND and dpkg flags suppress interactive prompts and ESC sequences.
		sudo := executor.Options{
			RequiresSudo: true,
			Stdout:       stdout,
			Stderr:       stdout,
			Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
		}
		script := `
set -e
KEYRING="/usr/share/keyrings/dbeaver.gpg.key"
SOURCES="/etc/apt/sources.list.d/dbeaver.list"
[ -f "$KEYRING" ] || wget -qO- https://dbeaver.io/debs/dbeaver.gpg.key \
  | gpg --dearmor -o "$KEYRING"
[ -f "$SOURCES" ] || echo "deb [signed-by=${KEYRING}] https://dbeaver.io/debs/dbeaver-ce /" \
  | tee "$SOURCES" > /dev/null
apt-get update -qq
apt-get install -y -o Dpkg::Use-Pty=0 -o Dpkg::Progress-Fancy=0 -o APT::Color=0 dbeaver-ce
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	case distro.Fedora:
		// DEBT-02: the RPM is x86_64-only; return a clear message on other architectures.
		if runtime.GOARCH != "amd64" {
			return fmt.Errorf("instalação automática do DBeaver no Fedora disponível apenas em amd64 — instale manualmente de https://dbeaver.io")
		}
		const rpmURL = "https://dbeaver.io/files/dbeaver-ce-latest-linux-x86_64.rpm"
		return exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", "install", "-y", "--", rpmURL,
		)
	default:
		return fmt.Errorf("instale o DBeaver manualmente no Arch via AUR (yay -S dbeaver)")
	}
}

func installVSCodium(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string) error {
	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch family {
	case distro.Debian:
		script := `
set -e
KEYRING="/usr/share/keyrings/vscodium-archive-keyring.gpg"
[ -f "$KEYRING" ] || wget -qO- https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg | gpg --dearmor | dd of="$KEYRING" status=none
if [ ! -f /etc/apt/sources.list.d/vscodium.sources ]; then
  printf 'Types: deb\nURIs: https://download.vscodium.com/debs\nSuites: vscodium\nComponents: main\nArchitectures: amd64 arm64\nSigned-by: /usr/share/keyrings/vscodium-archive-keyring.gpg\n' | tee /etc/apt/sources.list.d/vscodium.sources > /dev/null
fi
apt-get update -qq
apt-get install -y codium
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	case distro.Fedora:
		script := `
set -e
rpm --import https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg
[ -f /etc/yum.repos.d/vscodium.repo ] || printf '[vscodium]\nname=download.vscodium.com\nbaseurl=https://download.vscodium.com/rpms/\nenabled=1\ngpgcheck=1\ngpgkey=https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg\n' | tee /etc/yum.repos.d/vscodium.repo > /dev/null
dnf install -y codium
`
		return exe.Run(ctx, sudo, "bash", "-c", script)
	default:
		return fmt.Errorf("instale o VSCodium manualmente no Arch via AUR (yay -S vscodium-bin)")
	}
}

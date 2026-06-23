package postinstall

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// step prints a step panel and runs a sudo command.
func step(ctx context.Context, exe *executor.Executor, stdout io.Writer, msg, name string, args ...string) error {
	ui.Info(stdout, msg)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, name, args...)
}

// aptInstall installs one or more apt packages with sudo.
func aptInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkgs ...string) error {
	args := append([]string{"install", "-y", "-o", "Dpkg::Use-Pty=0", "-o", "Dpkg::Progress-Fancy=0", "-o", "APT::Color=0", "--"}, pkgs...)
	return exe.Run(ctx, executor.Options{
		RequiresSudo: true,
		Stdout:       stdout,
		Stderr:       stdout,
		Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
	}, "apt-get", args...)
}

// dnfInstall installs one or more dnf packages with sudo.
func dnfInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, pkgs ...string) error {
	args := append([]string{"install", "-y", "--"}, pkgs...)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}, "dnf", args...)
}

// ensureFlatpakReady checks if flatpak is present and adds the Flathub remote if needed.
func ensureFlatpakReady(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "which", "flatpak"); err != nil {
		ui.Info(stdout, "Instalando Flatpak...")
		if err := aptInstall(ctx, exe, stdout, "flatpak"); err != nil {
			return fmt.Errorf("instalar flatpak: %w", err)
		}
	}
	ui.Info(stdout, "Configurando repositório Flathub...")
	scope := config.FlatpakFlag()
	return exe.Run(ctx,
		executor.Options{RequiresSudo: scope == "--system", Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}},
		"flatpak", "remote-add", scope, "--if-not-exists", "flathub",
		"https://dl.flathub.org/repo/flathub.flatpakrepo",
	)
}

// flatpakInstall installs Flatpak apps from Flathub using the configured scope.
func flatpakInstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, appIDs ...string) error {
	args := append([]string{"install", "--noninteractive", config.FlatpakFlag(), "-y", "flathub"}, appIDs...)
	return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout, Env: []string{"TERM=dumb"}}, "flatpak", args...)
}

// configureSysctl sets swappiness, inotify and applies sysctl.
func configureSysctl(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.Info(stdout, "Aplicando configurações de kernel (sysctl)...")
	conf := "vm.swappiness=10\nfs.inotify.max_user_watches=524288\n"
	path := "/etc/sysctl.d/99-lumina.conf"
	cmd := fmt.Sprintf("printf '%%b' %q > %s && sysctl -p %s", conf, path, path)
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"bash", "-c", cmd,
	)
}

// acceptMsttFontsEula pre-accepts the EULA for ttf-mscorefonts-installer.
func acceptMsttFontsEula(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	sel := "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true"
	return exe.Run(ctx, executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"bash", "-c", fmt.Sprintf("printf '%%s\\n' %q | debconf-set-selections", sel),
	)
}

// failWith shows err as a UI error panel, waits for Enter, and returns the error.
func failWith(stdout io.Writer, err error) error {
	ui.Err(stdout, "Falha: "+err.Error())
	ui.WaitEnter(stdout)
	return err
}

// installVAAPI prompts for hardware video acceleration (Intel / AMD / skip)
// and installs the appropriate packages via the provided installer function.
// Failures are non-fatal: shown as warnings and execution continues.
func installVAAPI(
	ctx context.Context,
	exe *executor.Executor,
	stdout io.Writer,
	intel, amd []string,
	install func(context.Context, *executor.Executor, io.Writer, ...string) error,
) {
	ui.PrintBox(stdout, "1. Intel\n2. AMD\n3. Não instalar")
	fmt.Fprint(stdout, "Aceleração de vídeo (1/2/3): ")
	choice := strings.TrimSpace(prompt.ReadLine())
	switch choice {
	case "1":
		ui.Info(stdout, "Instalando drivers VA-API para Intel...")
		if err := install(ctx, exe, stdout, intel...); err != nil {
			ui.Warning(stdout, "Falha ao instalar VA-API Intel: "+err.Error())
		}
	case "2":
		ui.Info(stdout, "Instalando drivers VA-API para AMD...")
		if err := install(ctx, exe, stdout, amd...); err != nil {
			ui.Warning(stdout, "Falha ao instalar VA-API AMD: "+err.Error())
		}
	default:
		ui.Info(stdout, "Aceleração de vídeo por hardware ignorada.")
	}
}

// aptComponentEnabled reports whether the given apt component (e.g. "universe")
// is already present in at least one enabled source. Checks both legacy .list
// and DEB822 .sources formats. Returns false on any read error (safe to retry).
func aptComponentEnabled(ctx context.Context, exe *executor.Executor, component string) bool {
	out, _ := exe.Output(ctx, executor.Options{},
		"bash", "-c",
		"grep -rEh '^deb[^-]|^Components:' /etc/apt/sources.list /etc/apt/sources.list.d/ 2>/dev/null")
	for _, line := range strings.Split(out, "\n") {
		for _, field := range strings.Fields(line) {
			if field == component {
				return true
			}
		}
	}
	return false
}

// rpmFusionEnabled reports whether the rpmfusion-free-release package is installed.
func rpmFusionEnabled(ctx context.Context, exe *executor.Executor) bool {
	_, err := exe.Output(ctx, executor.Options{}, "rpm", "-q", "rpmfusion-free-release")
	return err == nil
}

func stripNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// removeSnaps lists installed snaps, presents a multi-select to the user, and removes
// the selected ones with multiple passes to handle dependency ordering.
// Non-fatal: skips silently if snap is not available.
func removeSnaps(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) {
	out, err := exe.Output(ctx, executor.Options{}, "snap", "list")
	if err != nil {
		return // snap not installed or unavailable
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= 1 {
		ui.Info(stdout, "Nenhum snap instalado.")
		return
	}

	items := make([]ui.SelectItem, 0, len(lines)-1)
	for _, line := range lines[1:] { // skip header row
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		items = append(items, ui.SelectItem{
			ID:    fields[0],
			Label: fields[0] + "  " + fields[1],
		})
	}
	if len(items) == 0 {
		ui.Info(stdout, "Nenhum snap instalado.")
		return
	}

	ui.Info(stdout, "Snaps instalados. Selecione os que deseja remover:")
	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		ui.Warning(stdout, "Falha na seleção de snaps: "+err.Error())
		return
	}
	if !confirmed {
		ui.Info(stdout, "Remoção de snaps cancelada.")
		return
	}

	var toRemove []string
	for _, item := range finalItems {
		if item.Selected {
			toRemove = append(toRemove, item.ID)
		}
	}
	if len(toRemove) == 0 {
		ui.Info(stdout, "Nenhum snap selecionado.")
		return
	}

	// Multiple passes handle implicit dependency ordering: if A depends on B and B
	// is removed first, the next pass will successfully remove A.
	removed := make(map[string]bool, len(toRemove))
	for range 3 {
		progress := false
		for _, name := range toRemove {
			if removed[name] {
				continue
			}
			if err := exe.Run(ctx,
				executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
				"snap", "remove", "--purge", "--", name,
			); err == nil {
				removed[name] = true
				progress = true
			}
		}
		if !progress {
			break
		}
	}

	for _, name := range toRemove {
		if !removed[name] {
			ui.Warning(stdout, "Snap não removido (verifique dependências): "+name)
		}
	}
}

// setupSwapfile creates a 4 GB swapfile at /swapfile and persists it in /etc/fstab.
// Skips silently if /swapfile already exists. Non-fatal: shows warnings on failure.
func setupSwapfile(ctx context.Context, exe *executor.Executor, stdout io.Writer) {
	ui.Info(stdout, "Configurando swapfile...")

	if _, err := os.Stat("/swapfile"); err == nil {
		ui.Info(stdout, "Swapfile já existe em /swapfile, pulando.")
		return
	}

	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	sudoSilent := executor.Options{RequiresSudo: true}

	if err := exe.Run(ctx, sudo, "fallocate", "-l", "4G", "/swapfile"); err != nil {
		ui.Warning(stdout, "Falha ao criar swapfile: "+err.Error())
		return
	}
	if err := exe.Run(ctx, sudo, "chmod", "600", "/swapfile"); err != nil {
		ui.Warning(stdout, "Falha ao definir permissões do swapfile: "+err.Error())
		_ = exe.Run(ctx, sudoSilent, "rm", "-f", "--", "/swapfile")
		return
	}
	if err := exe.Run(ctx, sudo, "mkswap", "/swapfile"); err != nil {
		ui.Warning(stdout, "Falha ao formatar swapfile: "+err.Error())
		_ = exe.Run(ctx, sudoSilent, "rm", "-f", "--", "/swapfile")
		return
	}
	if err := exe.Run(ctx, sudo, "swapon", "/swapfile"); err != nil {
		ui.Warning(stdout, "Falha ao ativar swapfile: "+err.Error())
		return
	}

	// Append /etc/fstab entry only if absent, preventing duplicates on reruns.
	const fstabScript = `grep -qF '/swapfile' /etc/fstab || printf '/swapfile none swap sw 0 0\n' >> /etc/fstab`
	if err := exe.Run(ctx, sudo, "bash", "-c", fstabScript); err != nil {
		ui.Warning(stdout, "Falha ao atualizar /etc/fstab: "+err.Error())
	}

	ui.Success(stdout, "Swapfile de 4 GB criado e ativado.")
}

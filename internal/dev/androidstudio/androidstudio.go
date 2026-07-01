package androidstudio

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const (
	installDir  = "android-studio"
	pathComment = "# Android Studio"
	pathEntry   = `export PATH="$PATH:$HOME/android-studio/bin"`
	downloadURL = "https://developer.android.com/studio?hl=pt-br"
)

// Manage checks whether Android Studio is installed and shows a menu to
// install, reinstall or remove it.
func Manage(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Android Studio")

	home, err := os.UserHomeDir()
	if err != nil {
		ui.Err(stdout, "Falha ao localizar diretório home: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	studioDir := filepath.Join(home, installDir)
	studioBin := filepath.Join(studioDir, "bin", "studio")

	var items []ui.SelectItem
	if isInstalled(studioBin) {
		ui.Info(stdout, "Android Studio já instalado em: "+studioDir)
		items = []ui.SelectItem{
			{Label: "Reinstalar", ID: "reinstall"},
			{Label: "Desinstalar", ID: "uninstall"},
			{Label: "Sair", ID: "exit"},
		}
	} else {
		ui.Info(stdout, "Android Studio não encontrado em: "+studioDir)
		items = []ui.SelectItem{
			{Label: "Instalar", ID: "install"},
			{Label: "Sair", ID: "exit"},
		}
	}

	idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !ok || idx < 0 || items[idx].ID == "exit" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}
	if items[idx].ID == "uninstall" {
		return uninstall(stdout, home, studioDir)
	}

	// install or reinstall — locate tarball first
	tarball, err := findTarball(ctx, stdin, stdout, home)
	if err != nil {
		ui.WaitEnter(stdout)
		return err
	}
	if tarball == "" {
		ui.WaitEnter(stdout)
		return nil
	}

	return install(ctx, exe, stdout, home, studioDir, tarball)
}

// isInstalled returns true when the studio binary exists.
func isInstalled(studioBin string) bool {
	info, err := os.Stat(studioBin)
	return err == nil && !info.IsDir()
}

// findTarball locates an android-studio-*-linux.tar.gz file in ~/Downloads.
// Returns the chosen path, an empty string if the user cancelled, or an error.
func findTarball(ctx context.Context, stdin io.Reader, stdout io.Writer, home string) (string, error) {
	downloadsDir := filepath.Join(home, "Downloads")
	matches, err := filepath.Glob(filepath.Join(downloadsDir, "android-studio-*-linux.tar.gz"))
	if err != nil {
		return "", fmt.Errorf("procurar tarball: %w", err)
	}

	if len(matches) == 0 {
		ui.Err(stdout, "Arquivo de instalação não foi encontrado na pasta Downloads.")
		ui.Info(stdout, "Faça o download em "+downloadURL+" e inicie a instalação novamente.")
		return "", nil
	}

	if len(matches) == 1 {
		name := filepath.Base(matches[0])
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "O arquivo %s foi encontrado. Continuar a instalação com este arquivo? (S/n): ", name)
		answer, _ := prompt.ReadLineFrom(stdin)
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "n" || answer == "não" || answer == "nao" {
			ui.Info(stdout, "Operação cancelada.")
			return "", nil
		}
		return matches[0], nil
	}

	// Multiple tarballs — let the user pick.
	ui.Info(stdout, "Múltiplos arquivos encontrados em Downloads:")
	items := make([]ui.SelectItem, len(matches))
	for i, m := range matches {
		items[i] = ui.SelectItem{Label: filepath.Base(m), ID: m}
	}
	idx, ok, err := ui.RunSingleSelect(ctx, stdin, stdout, items)
	if err != nil {
		return "", err
	}
	if !ok || idx < 0 {
		ui.Info(stdout, "Operação cancelada.")
		return "", nil
	}
	return items[idx].ID, nil
}

// install extracts the tarball to home, installs libraries, creates the
// desktop entry and adds the binary to PATH.
func install(ctx context.Context, exe *executor.Executor, stdout io.Writer, home, studioDir, tarball string) error {
	if err := installLibs(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao instalar bibliotecas de suporte: "+err.Error())
	}

	// Remove previous installation before extracting so that stale files from
	// a renamed release do not accumulate.
	if _, err := os.Stat(studioDir); err == nil {
		ui.Info(stdout, "Removendo instalação anterior...")
		if err := os.RemoveAll(studioDir); err != nil {
			ui.Err(stdout, "Falha ao remover instalação anterior: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("remover instalação anterior: %w", err)
		}
	}

	ui.Info(stdout, "Extraindo "+filepath.Base(tarball)+"...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"tar", "-xzf", tarball, "-C", home,
	); err != nil {
		ui.Err(stdout, "Falha ao extrair o tarball: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("extrair tarball: %w", err)
	}

	ensurePathInRC(stdout, home)

	if err := createDesktopEntry(home, studioDir); err != nil {
		ui.Warning(stdout, "Não foi possível criar o atalho no menu de aplicativos: "+err.Error())
	} else {
		ui.Info(stdout, "Atalho criado em ~/.local/share/applications/android-studio.desktop")
	}

	ui.Success(stdout, "Android Studio instalado com sucesso em: "+studioDir)
	ui.Warning(stdout, "Reinicie o terminal para ativar o Android Studio no PATH.")
	ui.Info(stdout, "Para iniciar: "+filepath.Join(studioDir, "bin", "studio"))
	ui.WaitEnter(stdout)
	return nil
}

// installLibs installs the 32-bit support libraries required by Android Studio
// on a 64-bit Linux machine.
func installLibs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	family := distro.Detect()
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	ui.Info(stdout, "Instalando bibliotecas de suporte (32-bit)...")
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "dpkg", "--add-architecture", "i386"); err != nil {
			return err
		}
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			return err
		}
		return exe.Run(ctx, opts, "apt-get", "install", "-y", "--",
			"libc6:i386", "libncurses6:i386", "libstdc++6:i386", "lib32z1", "libbz2-1.0:i386")
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", "install", "-y", "--",
			"zlib.i686", "ncurses-libs.i686", "bzip2-libs.i686")
	default:
		ui.Warning(stdout, "Distribuição não suportada para instalação automática de bibliotecas. Instale manualmente as bibliotecas 32-bit necessárias para o Android Studio.")
		return nil
	}
}

// uninstall removes the Android Studio directory, its PATH entry and desktop file.
func uninstall(stdout io.Writer, home, studioDir string) error {
	ui.Info(stdout, "Removendo Android Studio...")
	if err := os.RemoveAll(studioDir); err != nil {
		ui.Err(stdout, "Falha ao remover Android Studio: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	removePathFromRC(stdout, home)
	removeDesktopEntry()
	ui.Success(stdout, "Android Studio removido com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// createDesktopEntry writes a .desktop launcher for Android Studio.
func createDesktopEntry(home, studioDir string) error {
	dir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criar diretório de aplicações: %w", err)
	}
	iconPath := filepath.Join(studioDir, "bin", "studio.png")
	content := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=Android Studio
Comment=The official Android IDE
Exec=%s/bin/studio %%f
Icon=%s
Terminal=false
StartupNotify=true
Categories=Development;IDE;
MimeType=application/x-extension-iml;
`, studioDir, iconPath)

	dest := filepath.Join(dir, "android-studio.desktop")
	return os.WriteFile(dest, []byte(content), 0o644)
}

// removeDesktopEntry removes the .desktop file if it exists.
func removeDesktopEntry() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	_ = os.Remove(filepath.Join(home, ".local", "share", "applications", "android-studio.desktop"))
}

// shellRCFile returns the primary shell RC file for the current user's shell.
func shellRCFile(home string) string {
	switch filepath.Base(os.Getenv("SHELL")) {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

// ensurePathInRC adds android-studio/bin to PATH in the current shell's RC file.
func ensurePathInRC(stdout io.Writer, home string) {
	rc := shellRCFile(home)

	data, _ := os.ReadFile(rc)
	if strings.Contains(string(data), pathEntry) {
		return
	}

	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ui.Warning(stdout, "Não foi possível atualizar "+rc+": "+err.Error())
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n%s\n%s\n", pathComment, pathEntry)
	ui.Info(stdout, "PATH atualizado em "+rc)
}

// removePathFromRC removes the Android Studio PATH entry from RC files.
// Checks both ~/.bashrc (backward compat) and the current shell's RC file.
func removePathFromRC(stdout io.Writer, home string) {
	candidates := dedupStrings([]string{filepath.Join(home, ".bashrc"), shellRCFile(home)})
	for _, rc := range candidates {
		data, err := os.ReadFile(rc)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		out := make([]string, 0, len(lines))
		changed := false
		for _, l := range lines {
			if l == pathComment || l == pathEntry {
				changed = true
				continue
			}
			out = append(out, l)
		}
		if !changed {
			continue
		}
		if err := rewriteFile(rc, out); err != nil {
			ui.Warning(stdout, "Não foi possível atualizar "+rc+": "+err.Error())
			continue
		}
		ui.Info(stdout, "PATH removido de "+rc)
	}
}

func dedupStrings(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	out := ss[:0]
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// rewriteFile atomically replaces path with the content of lines joined by newlines.
func rewriteFile(path string, lines []string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".lumina-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := fmt.Fprint(tmp, strings.Join(lines, "\n")); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

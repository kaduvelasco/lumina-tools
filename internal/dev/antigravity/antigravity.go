package antigravity

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const (
	installDir  = "antigravity"
	binaryName  = "antigravity"
	symlinkPath = "/usr/local/bin/antigravity-ide"
	downloadURL = "https://antigravity.google/download"
	iconName    = "antigravity-ide"

	// realIDERelPath is where the bundled "IDE Wizard" installs the actual,
	// permanent IDE on first run (relative to $HOME) — the wizard binary
	// extracted from the tarball is a one-time bootstrapper, not the app
	// itself. Launchers must prefer this path once it exists, or the
	// shortcut/symlink stops opening anything after the first run.
	realIDERelPath = ".local/share/antigravity-ide/antigravity-ide"

	// minGlibc and minGlibcxx define the minimum runtime library versions
	// required by the Antigravity IDE on Linux.
	minGlibc   = "2.28"
	minGlibcxx = "3.4.25"
)

//go:embed icon.svg
var iconSVG []byte

// Manage checks whether Antigravity IDE is installed and shows a menu to
// install, reinstall or remove it.
func Manage(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Antigravity IDE")

	home, err := os.UserHomeDir()
	if err != nil {
		ui.Err(stdout, "Falha ao localizar diretório home: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	studioDir := filepath.Join(home, installDir)
	studioBin := filepath.Join(studioDir, binaryName)

	var items []ui.SelectItem
	if isInstalled(studioBin) {
		ui.Info(stdout, "Antigravity IDE já instalado em: "+studioDir)
		items = []ui.SelectItem{
			{Label: "Reinstalar", ID: "reinstall"},
			{Label: "Desinstalar", ID: "uninstall"},
			{Label: "Sair", ID: "exit"},
		}
	} else {
		ui.Info(stdout, "Antigravity IDE não encontrado em: "+studioDir)
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
		return uninstall(ctx, exe, stdout, home, studioDir)
	}

	// install or reinstall
	if ok, err := checkDeps(stdin, stdout); err != nil || !ok {
		if err != nil {
			ui.Err(stdout, "Falha ao verificar dependências: "+err.Error())
		}
		ui.WaitEnter(stdout)
		return err
	}

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

// isInstalled returns true when the Antigravity binary exists.
func isInstalled(bin string) bool {
	info, err := os.Stat(bin)
	return err == nil && !info.IsDir()
}

// checkDeps verifies that glibc >= 2.28 and glibcxx >= 3.4.25 are present.
// Returns (true, nil) when all requirements are met.
// Returns (false, nil) when a requirement fails but the user chose to abort.
func checkDeps(stdin io.Reader, stdout io.Writer) (bool, error) {
	ui.Info(stdout, "Verificando dependências do sistema...")

	glibcOK, glibcFound := checkGlibc(stdout)
	glibcxxOK, glibcxxFound := checkGlibcxx(stdout)

	allOK := glibcOK && glibcxxOK
	if allOK {
		ui.Success(stdout, "Dependências satisfeitas.")
		return true, nil
	}

	fmt.Fprintln(stdout)
	if !glibcOK {
		ui.Warning(stdout, fmt.Sprintf("glibc: encontrado %s, necessário >= %s", glibcFound, minGlibc))
	}
	if !glibcxxOK {
		ui.Warning(stdout, fmt.Sprintf("libstdc++ (GLIBCXX): encontrado %s, necessário >= %s", glibcxxFound, minGlibcxx))
	}
	ui.Warning(stdout, "O Antigravity IDE pode não funcionar corretamente neste sistema.")
	ui.Info(stdout, "Distribuições suportadas: Ubuntu 20+, Debian 10+, Fedora 36+, RHEL 8+.")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Continuar mesmo assim? (s/N): ")

	line, _ := prompt.ReadLineFrom(stdin)
	line = strings.ToLower(strings.TrimSpace(line))
	if line != "s" && line != "sim" {
		ui.Info(stdout, "Instalação cancelada.")
		return false, nil
	}
	return true, nil
}

// checkGlibc returns (meetsMin, versionFound).
func checkGlibc(stdout io.Writer) (bool, string) {
	out, err := exec.Command("ldd", "--version").Output()
	if err != nil {
		ui.Warning(stdout, "Não foi possível verificar a versão do glibc (ldd não encontrado).")
		return false, "desconhecida"
	}
	// First line: "ldd (Ubuntu GLIBC 2.35-0ubuntu3.6) 2.35"
	first := strings.SplitN(string(out), "\n", 2)[0]
	fields := strings.Fields(first)
	if len(fields) == 0 {
		return false, "desconhecida"
	}
	ver := fields[len(fields)-1]
	return versionAtLeast(ver, minGlibc), ver
}

// checkGlibcxx returns (meetsMin, versionFound) for libstdc++.
func checkGlibcxx(stdout io.Writer) (bool, string) {
	// Locate libstdc++.so.6 via ldconfig and grep for the highest GLIBCXX symbol.
	ldOut, err := exec.Command("ldconfig", "-p").Output()
	if err != nil {
		ui.Warning(stdout, "Não foi possível localizar libstdc++ via ldconfig.")
		return false, "desconhecida"
	}
	libPath := ""
	for _, line := range strings.Split(string(ldOut), "\n") {
		if strings.Contains(line, "libstdc++.so.6") && strings.Contains(line, "=>") {
			parts := strings.SplitN(line, "=>", 2)
			libPath = strings.TrimSpace(parts[1])
			break
		}
	}
	if libPath == "" {
		ui.Warning(stdout, "libstdc++.so.6 não encontrada no sistema.")
		return false, "desconhecida"
	}

	strOut, err := exec.Command("strings", libPath).Output()
	if err != nil {
		ui.Warning(stdout, "Não foi possível inspecionar libstdc++: "+err.Error())
		return false, "desconhecida"
	}

	highest := ""
	for _, line := range strings.Split(string(strOut), "\n") {
		if strings.HasPrefix(line, "GLIBCXX_") {
			ver := strings.TrimPrefix(line, "GLIBCXX_")
			if highest == "" || versionAtLeast(ver, highest) {
				highest = ver
			}
		}
	}
	if highest == "" {
		return false, "desconhecida"
	}
	return versionAtLeast(highest, minGlibcxx), highest
}

// versionAtLeast reports whether ver >= min, comparing dot-separated integers.
func versionAtLeast(ver, min string) bool {
	vp := strings.Split(ver, ".")
	mp := strings.Split(min, ".")
	n := len(vp)
	if len(mp) > n {
		n = len(mp)
	}
	for i := range n {
		v, m := 0, 0
		if i < len(vp) {
			v, _ = strconv.Atoi(vp[i])
		}
		if i < len(mp) {
			m, _ = strconv.Atoi(mp[i])
		}
		if v > m {
			return true
		}
		if v < m {
			return false
		}
	}
	return true
}

// findTarball locates an Antigravity*.tar.gz file in ~/Downloads.
// Returns the chosen path, empty string if the user cancelled, or an error.
func findTarball(ctx context.Context, stdin io.Reader, stdout io.Writer, home string) (string, error) {
	downloadsDir := filepath.Join(home, "Downloads")
	matches, err := filepath.Glob(filepath.Join(downloadsDir, "Antigravity*.tar.gz"))
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

// install extracts the tarball, creates the symlink and desktop entry.
func install(ctx context.Context, exe *executor.Executor, stdout io.Writer, home, studioDir, tarball string) error {
	// Remove previous installation before extracting.
	if _, err := os.Stat(studioDir); err == nil {
		ui.Info(stdout, "Removendo instalação anterior...")
		if err := os.RemoveAll(studioDir); err != nil {
			ui.Err(stdout, "Falha ao remover instalação anterior: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("remover instalação anterior: %w", err)
		}
	}

	if err := os.MkdirAll(studioDir, 0o755); err != nil {
		ui.Err(stdout, "Falha ao criar diretório de instalação: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("criar diretório: %w", err)
	}

	ui.Info(stdout, "Extraindo "+filepath.Base(tarball)+"...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"tar", "-xzf", tarball, "-C", studioDir, "--strip-components=1",
	); err != nil {
		os.RemoveAll(studioDir)
		ui.Err(stdout, "Falha ao extrair o tarball: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("extrair tarball: %w", err)
	}

	if err := createLauncher(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Não foi possível criar o launcher em /usr/local/bin: "+err.Error())
	}

	if err := installIcon(home); err != nil {
		ui.Warning(stdout, "Não foi possível instalar o ícone: "+err.Error())
	}

	if err := createDesktopEntry(home); err != nil {
		ui.Warning(stdout, "Não foi possível criar o atalho no menu de aplicativos: "+err.Error())
	} else {
		ui.Info(stdout, "Atalho criado em ~/.local/share/applications/antigravity-ide.desktop")
		_ = exec.Command("update-desktop-database",
			filepath.Join(home, ".local", "share", "applications")).Run()
	}

	ui.Success(stdout, "Antigravity IDE instalado com sucesso em: "+studioDir)
	ui.Info(stdout, "Para iniciar: "+symlinkPath)
	ui.Warning(stdout, "Na primeira execução, o instalador (\"IDE Wizard\") baixa a IDE completa para ~/"+realIDERelPath+" — feche e abra novamente pelo mesmo atalho após esse download terminar.")
	ui.WaitEnter(stdout)
	return nil
}

// createLauncher installs /usr/local/bin/antigravity-ide as a small wrapper
// script instead of a plain symlink. The tarball only ships a one-time "IDE
// Wizard" that, on first run, downloads the real, permanent IDE to
// ~/<realIDERelPath>. A plain symlink to the wizard binary stops opening
// anything once that download completes, so the wrapper prefers the real IDE
// when present and falls back to the wizard otherwise.
func createLauncher(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	script := fmt.Sprintf(`#!/usr/bin/env bash
real="$HOME/%s"
if [ -x "$real" ]; then
    exec "$real" "$@"
fi
exec "$HOME/%s" "$@"
`, realIDERelPath, filepath.Join(installDir, binaryName))

	tmp, err := os.CreateTemp("", "lumina-antigravity-launcher-*.sh")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	_ = exe.Run(ctx, sudo, "rm", "-f", "--", symlinkPath)
	return exe.Run(ctx, sudo, "install", "-m", "0755", "--", tmpPath, symlinkPath)
}

// installIcon writes the embedded Antigravity SVG to the user's icon theme
// so Icon=antigravity-ide in the .desktop entry resolves to a real image.
func installIcon(home string) error {
	dir := filepath.Join(home, ".local", "share", "icons", "hicolor", "scalable", "apps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criar diretório de ícones: %w", err)
	}
	dest := filepath.Join(dir, iconName+".svg")
	return os.WriteFile(dest, iconSVG, 0o644)
}

// createDesktopEntry writes a .desktop launcher for Antigravity IDE.
func createDesktopEntry(home string) error {
	dir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criar diretório de aplicações: %w", err)
	}
	content := fmt.Sprintf(`[Desktop Entry]
Name=Antigravity IDE
Comment=Antigravity IDE - Experience liftoff
GenericName=IDE
Exec=%s %%F
Icon=%s
Type=Application
Terminal=false
Categories=Development;IDE;
StartupWMClass=antigravity-ide
`, symlinkPath, iconName)

	dest := filepath.Join(dir, "antigravity-ide.desktop")
	return os.WriteFile(dest, []byte(content), 0o644)
}

// uninstall removes the Antigravity IDE directory, launcher, icon and desktop entry.
func uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer, home, studioDir string) error {
	ui.Info(stdout, "Removendo Antigravity IDE...")
	if err := os.RemoveAll(studioDir); err != nil {
		ui.Err(stdout, "Falha ao remover Antigravity IDE: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	if err := exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"rm", "-f", "--", symlinkPath,
	); err != nil {
		ui.Warning(stdout, "Não foi possível remover o launcher: "+err.Error())
	}
	desktopFile := filepath.Join(home, ".local", "share", "applications", "antigravity-ide.desktop")
	_ = os.Remove(desktopFile)
	iconFile := filepath.Join(home, ".local", "share", "icons", "hicolor", "scalable", "apps", iconName+".svg")
	_ = os.Remove(iconFile)
	_ = exec.Command("update-desktop-database",
		filepath.Join(home, ".local", "share", "applications")).Run()

	ui.Success(stdout, "Antigravity IDE removido com sucesso.")
	ui.Info(stdout, "A IDE completa baixada pelo assistente na primeira execução (~/"+realIDERelPath+") não foi removida — apague manualmente se não for mais usá-la.")
	ui.WaitEnter(stdout)
	return nil
}

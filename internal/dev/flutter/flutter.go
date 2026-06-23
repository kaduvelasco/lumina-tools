package flutter

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
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const (
	flutterRepo     = "https://github.com/flutter/flutter.git"
	pathEntry       = `export PATH="$PATH:$HOME/development/flutter/bin"`
	chromeFlatpakID = "com.google.Chrome"

	// androidCmdlineToolsURL is pinned to a specific build (cmdline-tools 13.0)
	// rather than scraped from Google's download page, which has no stable
	// "latest" alias — same integrity policy used for the PHARs in the PHP stack.
	androidCmdlineToolsURL = "https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip"

	androidEnvBlock = `export ANDROID_HOME="$HOME/Android/Sdk"
export ANDROID_SDK_ROOT="$HOME/Android/Sdk"
export PATH="$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools"`
)

// Manage checks whether Flutter is installed and offers to install or update it.
func Manage(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Instalar Flutter + Dart")

	home, err := os.UserHomeDir()
	if err != nil {
		ui.Err(stdout, "Falha ao localizar diretório home: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	flutterDir := filepath.Join(home, "development", "flutter")
	flutterBin := filepath.Join(flutterDir, "bin", "flutter")

	if isInstalled(flutterBin) {
		ui.Info(stdout, "Flutter já instalado em: "+flutterDir)
		fmt.Fprint(stdout, "\nDeseja atualizar? (s/N): ")
		line, _ := prompt.ReadLineFrom(stdin)
		if c := strings.TrimSpace(line); c == "s" || c == "S" {
			ui.Info(stdout, "Atualizando Flutter...")
			if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, flutterBin, "upgrade"); err != nil {
				ui.Err(stdout, "Falha ao atualizar: "+err.Error())
				ui.WaitEnter(stdout)
				return err
			}
			ui.Success(stdout, "Flutter atualizado com sucesso!")
		}
	} else {
		if err := cloneFlutter(ctx, exe, stdout, home, flutterDir); err != nil {
			ui.Err(stdout, "Falha ao clonar o Flutter: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
		ensurePathInBashrc(stdout, home)
		ui.Success(stdout, "Flutter instalado com sucesso em: "+flutterDir)
		ui.Warning(stdout, "Reinicie o terminal para ativar o Flutter no PATH.")
	}

	// Prereqs and the Android SDK are (re)checked on every run — apt/dnf installs
	// are no-ops when already present, so this also backfills toolchains for
	// installs that predate these checks, not just fresh installs.
	if err := installPrereqs(ctx, exe, stdout); err != nil {
		ui.Warning(stdout, "Falha ao instalar pré-requisitos: "+err.Error())
	}
	if err := ensureAndroidCmdlineTools(ctx, exe, stdout, home, flutterBin); err != nil {
		ui.Warning(stdout, "Falha ao configurar o Android SDK: "+err.Error())
	}
	if err := ensureChromeWrapper(ctx, exe, stdout, home); err != nil {
		ui.Warning(stdout, "Falha ao configurar o wrapper do Chrome: "+err.Error())
	}

	ui.Info(stdout, "Executando diagnóstico (flutter doctor)...")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, flutterBin, "doctor"); err != nil {
		ui.Warning(stdout, "flutter doctor reportou pendências — verifique a saída acima.")
	}

	ui.WaitEnter(stdout)
	return nil
}

func isInstalled(flutterBin string) bool {
	info, err := os.Stat(flutterBin)
	return err == nil && !info.IsDir()
}

// installPrereqs installs the native libraries Flutter needs to run on Linux,
// plus the toolchain "flutter doctor" requires to build Linux desktop apps
// (clang, cmake, ninja-build, pkg-config, GTK 3.0 dev libraries). Package
// names differ between apt and dnf, so each family is handled explicitly;
// unsupported families are skipped with a warning instead of guessing a
// package manager.
func installPrereqs(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	family := distro.Detect()
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	ui.Info(stdout, "Instalando pré-requisitos do Flutter...")
	switch family {
	case distro.Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			return err
		}
		return exe.Run(ctx, opts, "apt-get", "install", "-y", "--",
			"curl", "git", "unzip", "xz-utils", "zip", "libglu1-mesa",
			"clang", "cmake", "ninja-build", "pkg-config", "libgtk-3-dev")
	case distro.Fedora:
		// dnf check-update exits 100 when updates are available, so its result
		// is informational only — never treated as a fatal error here.
		_ = exe.Run(ctx, opts, "dnf", "check-update")
		return exe.Run(ctx, opts, "dnf", "install", "-y", "--",
			"curl", "git", "unzip", "xz-utils", "zip", "mesa-libGL",
			"clang", "cmake", "ninja-build", "pkgconf-pkg-config", "gtk3-devel")
	default:
		ui.Warning(stdout, "Distribuição não suportada para instalação automática de pré-requisitos. Instale manualmente: curl, git, unzip, xz-utils, zip, a biblioteca OpenGL (libGL), clang, cmake, ninja-build, pkg-config e as bibliotecas de desenvolvimento do GTK 3.0.")
		return nil
	}
}

// ensureAndroidCmdlineTools installs the Android SDK command-line tools that
// "flutter doctor" requires for Android development, accepts the SDK
// licenses and installs platform-tools. Skips silently if already present.
func ensureAndroidCmdlineTools(ctx context.Context, exe *executor.Executor, stdout io.Writer, home, flutterBin string) error {
	androidHome := filepath.Join(home, "Android", "Sdk")
	cmdlineToolsDir := filepath.Join(androidHome, "cmdline-tools")
	latestDir := filepath.Join(cmdlineToolsDir, "latest")
	sdkmanagerBin := filepath.Join(latestDir, "bin", "sdkmanager")

	if isInstalled(sdkmanagerBin) {
		ui.Info(stdout, "Android cmdline-tools já instalado em: "+androidHome)
		// Re-affirm the SDK path with Flutter — flutter doctor relies on this
		// setting (~/.config/flutter/config) when ANDROID_HOME isn't exported
		// in the current process, which is always true right after install.
		return exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, flutterBin, "config", "--android-sdk", androidHome)
	}

	if err := ensureJava(ctx, exe, stdout); err != nil {
		return fmt.Errorf("instalar Java: %w", err)
	}

	tmp, err := os.MkdirTemp("", "lumina-android-*")
	if err != nil {
		return fmt.Errorf("criar diretório temporário: %w", err)
	}
	defer os.RemoveAll(tmp)

	zipPath := filepath.Join(tmp, "cmdline-tools.zip")
	ui.Info(stdout, "Baixando Android cmdline-tools...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"curl", "-fSL", "--progress-bar", "-o", zipPath, androidCmdlineToolsURL,
	); err != nil {
		return fmt.Errorf("baixar cmdline-tools: %w", err)
	}

	if err := os.MkdirAll(cmdlineToolsDir, 0o755); err != nil {
		return fmt.Errorf("criar diretório do Android SDK: %w", err)
	}

	ui.Info(stdout, "Extraindo cmdline-tools...")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "unzip", "-q", "-o", zipPath, "-d", tmp); err != nil {
		return fmt.Errorf("extrair cmdline-tools: %w", err)
	}

	// The zip extracts to <tmp>/cmdline-tools/{bin,lib,...} — Android's tooling
	// expects that directory renamed to <androidHome>/cmdline-tools/latest.
	_ = os.RemoveAll(latestDir)
	if err := os.Rename(filepath.Join(tmp, "cmdline-tools"), latestDir); err != nil {
		return fmt.Errorf("mover cmdline-tools: %w", err)
	}

	ensureAndroidEnvInBashrc(stdout, home)

	ui.Info(stdout, "Aceitando licenças do Android SDK...")
	licenseScript := "yes | '" + sdkmanagerBin + "' --licenses >/dev/null"
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "bash", "-c", licenseScript); err != nil {
		return fmt.Errorf("aceitar licenças: %w", err)
	}

	ui.Info(stdout, "Instalando platform-tools...")
	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, sdkmanagerBin, "platform-tools"); err != nil {
		return fmt.Errorf("instalar platform-tools: %w", err)
	}

	if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, flutterBin, "config", "--android-sdk", androidHome); err != nil {
		return fmt.Errorf("configurar android-sdk no flutter: %w", err)
	}

	ui.Success(stdout, "Android cmdline-tools instalado em: "+androidHome)
	return nil
}

// ensureJava installs a JDK when "java" is not already on the PATH.
// sdkmanager (used by ensureAndroidCmdlineTools) requires Java to run.
func ensureJava(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	if _, err := exe.Output(ctx, executor.Options{}, "java", "-version"); err == nil {
		return nil
	}

	family := distro.Detect()
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	ui.Info(stdout, "Instalando Java (necessário para o Android SDK)...")
	switch family {
	case distro.Debian:
		return exe.Run(ctx, opts, "apt-get", "install", "-y", "--", "default-jdk")
	case distro.Fedora:
		return exe.Run(ctx, opts, "dnf", "install", "-y", "--", "java-17-openjdk-devel")
	default:
		return fmt.Errorf("distribuição não suportada para instalação automática do Java — instale um JDK manualmente")
	}
}

// ensureAndroidEnvInBashrc adds ANDROID_HOME/ANDROID_SDK_ROOT and the SDK's
// PATH entries to ~/.bashrc if not already present.
func ensureAndroidEnvInBashrc(stdout io.Writer, home string) {
	bashrc := filepath.Join(home, ".bashrc")

	data, _ := os.ReadFile(bashrc)
	if strings.Contains(string(data), "ANDROID_HOME") {
		return
	}

	f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ui.Warning(stdout, "Não foi possível atualizar ~/.bashrc: "+err.Error())
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# Android SDK\n%s\n", androidEnvBlock)
	ui.Info(stdout, "Variáveis do Android SDK adicionadas ao ~/.bashrc")
}

// ensureChromeWrapper makes "flutter doctor" find Chrome when it's only
// installed via Flatpak — Flutter's web tooling spawns "google-chrome"
// directly, and Flatpak apps aren't exposed as plain PATH binaries. Does
// nothing when a real Chrome/Chromium binary is already on PATH, or when
// Chrome isn't installed at all (native or Flatpak).
func ensureChromeWrapper(ctx context.Context, exe *executor.Executor, stdout io.Writer, home string) error {
	wrapperPath := filepath.Join(home, ".local", "bin", "google-chrome")

	if _, err := os.Stat(wrapperPath); err == nil {
		return nil
	}

	for _, cmd := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if _, err := exe.Output(ctx, executor.Options{}, "which", cmd); err == nil {
			return nil
		}
	}

	if _, err := exe.Output(ctx, executor.Options{}, "flatpak", "info", chromeFlatpakID); err != nil {
		// Flatpak not installed, or Chrome isn't installed through it.
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(wrapperPath), 0o755); err != nil {
		return fmt.Errorf("criar ~/.local/bin: %w", err)
	}

	script := "#!/usr/bin/env bash\nexec flatpak run " + chromeFlatpakID + " \"$@\"\n"
	if err := os.WriteFile(wrapperPath, []byte(script), 0o755); err != nil {
		return fmt.Errorf("criar wrapper do Chrome: %w", err)
	}

	localbin.EnsureInPath(stdout)
	ensureChromeEnvInBashrc(stdout, home, wrapperPath)

	ui.Success(stdout, "Wrapper do Chrome (Flatpak) criado em: "+wrapperPath)
	return nil
}

// ensureChromeEnvInBashrc points CHROME_EXECUTABLE at wrapperPath, exactly
// the variable "flutter doctor" suggests setting when it can't find Chrome.
func ensureChromeEnvInBashrc(stdout io.Writer, home, wrapperPath string) {
	bashrc := filepath.Join(home, ".bashrc")

	data, _ := os.ReadFile(bashrc)
	if strings.Contains(string(data), "CHROME_EXECUTABLE") {
		return
	}

	f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ui.Warning(stdout, "Não foi possível atualizar ~/.bashrc: "+err.Error())
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# Chrome (Flatpak)\nexport CHROME_EXECUTABLE=%q\n", wrapperPath)
	ui.Info(stdout, "CHROME_EXECUTABLE adicionado ao ~/.bashrc")
}

// cloneFlutter clones the stable channel into flutterDir, creating the parent
// development directory first.
func cloneFlutter(ctx context.Context, exe *executor.Executor, stdout io.Writer, home, flutterDir string) error {
	if err := os.MkdirAll(filepath.Join(home, "development"), 0o755); err != nil {
		return fmt.Errorf("criar diretório development: %w", err)
	}

	ui.Info(stdout, "Clonando "+flutterRepo+" (branch stable)...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"git", "clone", "--branch", "stable", flutterRepo, flutterDir,
	); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// ensurePathInBashrc adds flutter/bin to PATH in ~/.bashrc if not already present.
func ensurePathInBashrc(stdout io.Writer, home string) {
	bashrc := filepath.Join(home, ".bashrc")

	data, _ := os.ReadFile(bashrc)
	if strings.Contains(string(data), pathEntry) {
		return
	}

	f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ui.Warning(stdout, "Não foi possível atualizar ~/.bashrc: "+err.Error())
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# Flutter\n%s\n", pathEntry)
	ui.Info(stdout, "PATH atualizado em ~/.bashrc")
}

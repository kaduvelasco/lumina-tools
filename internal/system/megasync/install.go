package megasync

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/distro"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

type pkgInfo struct {
	url string
	rpm bool // false = deb, true = rpm
}

// Install detects the running distribution and installs MegaSync CE.
func Install(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Instalar MegaSync")

	// DEBT-02: automatic install is only available for amd64/x86_64 packages.
	if runtime.GOARCH != "amd64" {
		ui.Err(stdout, "MegaSync: instalação automática disponível apenas em amd64 (detectado: "+runtime.GOARCH+") — instale manualmente em https://mega.nz/sync")
		ui.WaitEnter(stdout)
		return fmt.Errorf("arquitetura não suportada: %s", runtime.GOARCH)
	}

	if _, err := exe.Output(ctx, executor.Options{}, "which", "megasync"); err == nil {
		ui.Info(stdout, "MegaSync já instalado, pulando.")
		ui.WaitEnter(stdout)
		return nil
	}

	// CON-01: use distro.RawID / distro.VersionID instead of a local duplicate parser.
	pkg, err := resolvePackage(distro.RawID(), distro.VersionID())
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	pattern := "megasync-*.deb"
	if pkg.rpm {
		pattern = "megasync-*.rpm"
	}

	// DRY-01: download step encapsulated in downloadToTemp.
	ui.Info(stdout, "Baixando MegaSync...")
	tmpPath, err := downloadToTemp(ctx, exe, pkg.url, pattern)
	if err != nil {
		ui.Err(stdout, "Falha ao baixar MegaSync: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("download megasync: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	ui.Info(stdout, "Instalando MegaSync...")
	var installErr error
	if pkg.rpm {
		installErr = exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"dnf", "install", "-y", "--", tmpPath,
		)
	} else {
		installErr = exe.Run(ctx,
			executor.Options{
				RequiresSudo: true,
				Stdout:       stdout,
				Stderr:       stdout,
				Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
			},
			"apt-get", "install", "-y", "-o", "Dpkg::Use-Pty=0", "-o", "Dpkg::Progress-Fancy=0", "-o", "APT::Color=0", "--", tmpPath,
		)
	}

	if installErr != nil {
		ui.Err(stdout, "Falha ao instalar MegaSync: "+installErr.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("instalar megasync: %w", installErr)
	}

	ui.Success(stdout, "MegaSync instalado com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// resolvePackage maps the given distro id and version to the appropriate package URL.
// Accepting id and ver as parameters makes the logic testable without file I/O.
func resolvePackage(id, ver string) (pkgInfo, error) {
	if id == "" {
		return pkgInfo{}, fmt.Errorf(
			"não foi possível identificar a distribuição — verifique /etc/os-release",
		)
	}

	switch {
	case id == "linuxmint" && strings.HasPrefix(ver, "22"):
		return pkgInfo{url: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb"}, nil

	case (id == "ubuntu" && ver == "24.04") ||
		(id == "zorin" && strings.HasPrefix(ver, "18")):
		return pkgInfo{url: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb"}, nil

	case id == "ubuntu" && ver == "26.04":
		return pkgInfo{url: "https://mega.nz/linux/repo/xUbuntu_26.04/amd64/megasync-xUbuntu_26.04_amd64.deb"}, nil

	case id == "fedora" && ver == "44":
		return pkgInfo{
			url: "https://mega.nz/linux/repo/Fedora_44/x86_64/megasync-Fedora_44.x86_64.rpm",
			rpm: true,
		}, nil

	default:
		return pkgInfo{}, fmt.Errorf(
			"distribuição não suportada para MegaSync: %s %s\nInstale manualmente em https://mega.nz/sync",
			id, ver,
		)
	}
}

// downloadToTemp creates a secure temporary file matching pattern, downloads url into it,
// and returns the path. The caller must remove the file when done.
// The temporary file is removed automatically if the download fails.
func downloadToTemp(ctx context.Context, exe *executor.Executor, url, pattern string) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("criar arquivo temporário: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	if err := exe.Run(ctx, executor.Options{}, "wget", "-q", "-O", tmpPath, url); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

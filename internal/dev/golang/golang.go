package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const (
	goInstallDir = "/usr/local/go"
	goVersionAPI = "https://go.dev/dl/?mode=json"
	downloadBase = "https://dl.google.com/go"
	pathEntry    = "export PATH=$PATH:/usr/local/go/bin"
)

type goRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

// Manage checks whether Go is installed and shows a menu to install, update or remove it.
func Manage(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevStuff :: Gerenciar Go")

	installed, current := installedVersion(ctx, exe)

	var items []ui.SelectItem
	if installed {
		ui.Info(stdout, "Go instalado: "+current)
		items = []ui.SelectItem{{Label: "Atualizar", ID: "update"}, {Label: "Desinstalar", ID: "uninstall"}, {Label: "Sair", ID: "exit"}}
	} else {
		ui.Info(stdout, "Go não encontrado no sistema.")
		items = []ui.SelectItem{{Label: "Instalar", ID: "install"}, {Label: "Sair", ID: "exit"}}
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
		return uninstall(ctx, exe, stdout)
	}

	ui.Info(stdout, "Verificando versão mais recente...")
	latest, err := latestVersion(ctx)
	if err != nil {
		ui.Err(stdout, "Falha ao consultar API do go.dev: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	if installed && current == latest {
		ui.Success(stdout, "Go já está na versão mais recente: "+latest)
		ui.WaitEnter(stdout)
		return nil
	}

	if installed {
		ui.Info(stdout, fmt.Sprintf("Atualizando %s → %s...", current, latest))
	} else {
		ui.Info(stdout, "Instalando "+latest+"...")
	}

	if err := installGo(ctx, exe, stdout, latest); err != nil {
		ui.Err(stdout, "Falha na instalação: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ensurePathInBashrc(stdout)

	ui.Success(stdout, "Go "+latest+" instalado com sucesso!")
	ui.Warning(stdout, "Reinicie o terminal para ativar o Go no PATH.")
	ui.WaitEnter(stdout)
	return nil
}

// installedVersion returns whether Go is present at the standard install path
// and its version string (e.g. "go1.26.3").
func installedVersion(ctx context.Context, exe *executor.Executor) (bool, string) {
	out, err := exe.Output(ctx, executor.Options{}, goInstallDir+"/bin/go", "version")
	if err != nil {
		return false, ""
	}
	// Output: "go version go1.26.3 linux/amd64"
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) >= 3 {
		return true, parts[2]
	}
	return false, ""
}

// latestVersion fetches the latest stable Go release from go.dev.
func latestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, goVersionAPI, nil)
	if err != nil {
		return "", fmt.Errorf("criar requisição: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("consultar API: %w", err)
	}
	defer resp.Body.Close()

	var releases []goRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("decodificar resposta: %w", err)
	}

	for _, r := range releases {
		if r.Stable {
			return r.Version, nil
		}
	}
	return "", fmt.Errorf("nenhuma versão estável encontrada na API")
}

// installGo downloads the tarball for version and installs it to /usr/local/go.
// Extraction happens to a staging directory first so that a failed extraction
// never leaves the system without a working Go installation.
func installGo(ctx context.Context, exe *executor.Executor, stdout io.Writer, version string) error {
	tarball := version + ".linux-amd64.tar.gz"
	url := downloadBase + "/" + tarball

	tmp, err := os.MkdirTemp("", "lumina-go-*")
	if err != nil {
		return fmt.Errorf("criar diretório temporário: %w", err)
	}
	defer os.RemoveAll(tmp)

	dest := filepath.Join(tmp, tarball)

	ui.Info(stdout, "Baixando "+url+"...")
	if err := exe.Run(ctx,
		executor.Options{Stdout: stdout, Stderr: stdout},
		"curl", "-fSL", "--progress-bar", "-o", dest, url,
	); err != nil {
		return fmt.Errorf("baixar tarball: %w", err)
	}

	// Extract to a staging directory before touching the live installation.
	// If extraction fails, the existing Go is left intact.
	staging := goInstallDir + "-staging"
	sudo := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}

	ui.Info(stdout, "Extraindo...")
	_ = exe.Run(ctx, sudo, "rm", "-rf", "--", staging)
	if err := exe.Run(ctx, sudo, "mkdir", "-p", "--", staging); err != nil {
		return fmt.Errorf("criar diretório de staging: %w", err)
	}
	if err := exe.Run(ctx, sudo, "tar", "--strip-components=1", "-C", staging, "-xzf", dest); err != nil {
		_ = exe.Run(ctx, sudo, "rm", "-rf", "--", staging)
		return fmt.Errorf("extrair tarball: %w", err)
	}

	// Extraction succeeded — replace the live installation atomically.
	// mv within /usr/local is a single rename syscall on the same filesystem.
	ui.Info(stdout, "Instalando...")
	if err := exe.Run(ctx, sudo, "rm", "-rf", "--", goInstallDir); err != nil {
		_ = exe.Run(ctx, sudo, "rm", "-rf", "--", staging)
		return fmt.Errorf("remover instalação antiga: %w", err)
	}
	if err := exe.Run(ctx, sudo, "mv", "--", staging, goInstallDir); err != nil {
		return fmt.Errorf("mover instalação: %w", err)
	}

	return nil
}

// uninstall removes the Go installation and its PATH entry in ~/.bashrc.
func uninstall(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.Info(stdout, "Removendo Go...")
	if err := exe.Run(ctx,
		executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
		"rm", "-rf", "--", goInstallDir,
	); err != nil {
		ui.Err(stdout, "Falha ao remover Go: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	removePathFromBashrc(stdout)

	ui.Success(stdout, "Go removido com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// removePathFromBashrc undoes ensurePathInBashrc, removing the "# Go" comment
// and PATH export line added when Go was installed.
func removePathFromBashrc(stdout io.Writer) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	bashrc := filepath.Join(home, ".bashrc")

	data, err := os.ReadFile(bashrc)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if l == "# Go" || l == pathEntry {
			continue
		}
		out = append(out, l)
	}

	if err := os.WriteFile(bashrc, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		ui.Warning(stdout, "Não foi possível atualizar ~/.bashrc: "+err.Error())
		return
	}
	ui.Info(stdout, "PATH removido de ~/.bashrc")
}

// ensurePathInBashrc adds /usr/local/go/bin to PATH in ~/.bashrc if not already present.
func ensurePathInBashrc(stdout io.Writer) {
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Warning(stdout, "Não foi possível localizar o diretório home.")
		return
	}

	bashrc := filepath.Join(home, ".bashrc")

	data, _ := os.ReadFile(bashrc)
	if strings.Contains(string(data), pathEntry) {
		return
	}

	f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ui.Warning(stdout, "Não foi possível atualizar ~/.bashrc: "+err.Error())
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# Go\n%s\n", pathEntry)
	ui.Info(stdout, "PATH atualizado em ~/.bashrc")
}

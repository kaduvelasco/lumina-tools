package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/version"
)

const githubRepo = "kaduvelasco/lumina-tools"

// Release holds information about a GitHub release.
type Release struct {
	Version    string
	DownloadURL string
}

// Run checks for a newer version and applies the update if one is available.
func Run(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	fmt.Fprintf(stdout, "\n=== Atualizar Lumina Tools ===\n\n")
	fmt.Fprintf(stdout, "Versao atual: %s\n", version.Version)

	rel, err := latestRelease(ctx)
	if err != nil {
		return fmt.Errorf("verificar atualizacoes: %w", err)
	}

	fmt.Fprintf(stdout, "Versao disponivel: %s\n\n", rel.Version)

	normalize := func(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }
	if normalize(rel.Version) == normalize(version.Version) {
		fmt.Fprintln(stdout, "+ Voce ja esta usando a versao mais recente.")
		return nil
	}

	fmt.Fprintf(stdout, "-> Baixando %s (%s/%s)...\n", rel.Version, runtime.GOOS, runtime.GOARCH)
	if err := apply(ctx, exe, stdout, rel); err != nil {
		return fmt.Errorf("atualizar binario: %w", err)
	}

	fmt.Fprintf(stdout, "\n+ Lumina atualizado para %s. Reinicie o programa.\n", rel.Version)
	return nil
}

// latestRelease queries the GitHub Releases API and returns the latest release.
func latestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lumina-tools/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API retornou status %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decodificar resposta: %w", err)
	}

	v := strings.TrimPrefix(payload.TagName, "v")
	assetName := fmt.Sprintf("lumina-%s-%s", runtime.GOOS, runtime.GOARCH)

	var downloadURL string
	for _, a := range payload.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("asset '%s' nao encontrado na release %s", assetName, payload.TagName)
	}

	return &Release{Version: "v" + v, DownloadURL: downloadURL}, nil
}

// apply downloads the new binary and replaces the current one.
func apply(ctx context.Context, exe *executor.Executor, stdout io.Writer, rel *Release) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("localizar binario atual: %w", err)
	}

	// Download to the system temp dir — always writable, avoids EPERM in /usr/local/bin.
	tmp, err := os.CreateTemp("", "lumina-*.new")
	if err != nil {
		return fmt.Errorf("criar arquivo temporario: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath) // no-op if already renamed/moved

	fmt.Fprintf(stdout, "   -> Baixando para %s...\n", tmpPath)
	if err := downloadFile(ctx, rel.DownloadURL, tmpPath); err != nil {
		return fmt.Errorf("baixar: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	// Try atomic rename first; fall back to sudo mv if permission denied.
	if err := os.Rename(tmpPath, currentExe); err != nil {
		fmt.Fprintln(stdout, "   -> Permissao negada. Tentando com sudo...")
		if sudoErr := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"mv", "--", tmpPath, currentExe,
		); sudoErr != nil {
			return fmt.Errorf("substituir binario: %w (original: %w)", sudoErr, err)
		}
	}

	return nil
}

// downloadFile downloads url to dest using the context for cancellation.
func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download retornou status %d", resp.StatusCode)
	}

	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(f, io.LimitReader(resp.Body, 256<<20))
	closeErr := f.Close()

	if copyErr != nil {
		_ = os.Remove(dest)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dest)
		return closeErr
	}
	return nil
}

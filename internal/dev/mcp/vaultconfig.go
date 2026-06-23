package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// vaultPackage is the npm package name of the Lumina Vault MCP server.
const vaultPackage = "lumina-vault"

func vaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".lumina-vault", "config.json"), nil
}

// loadVaultConfig reads ~/.lumina-vault/config.json as a generic map, preserving
// any keys managed by the lumina-vault MCP server itself. An empty map is
// returned when the file does not yet exist.
func loadVaultConfig() (map[string]any, error) {
	path, err := vaultConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault config: %w", err)
	}

	cfg := map[string]any{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse vault config: %w", err)
		}
	}
	return cfg, nil
}

// saveVaultConfig writes cfg to ~/.lumina-vault/config.json atomically,
// creating the directory if needed.
func saveVaultConfig(cfg map[string]any) error {
	path, err := vaultConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create vault config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vault config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "config-*.json")
	if err != nil {
		return fmt.Errorf("create temp vault config: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write vault config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close vault config: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod vault config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write vault config: %w", err)
	}
	return nil
}

// configureVaultPath checks ~/.lumina-vault/config.json for the globalVaultPath
// key after the Lumina Vault MCP server is installed. If the key is already
// set, it informs the current value and asks whether to keep or change it;
// otherwise it asks for the path right away.
func configureVaultPath(stdout io.Writer) error {
	cfg, err := loadVaultConfig()
	if err != nil {
		ui.Err(stdout, err.Error())
		return err
	}

	current, hasPath := cfg["globalVaultPath"].(string)
	hasPath = hasPath && current != ""

	var newPath string
	if hasPath {
		ui.Info(stdout, "Caminho atual do vault global: "+current)
		fmt.Fprint(stdout, "Deseja alterar este caminho? (s/N): ")
		confirm := strings.TrimSpace(prompt.ReadLine())
		if confirm != "s" && confirm != "S" {
			return nil
		}
		fmt.Fprint(stdout, "Novo caminho do vault global: ")
	} else {
		fmt.Fprint(stdout, "Caminho do vault global: ")
	}
	newPath = strings.TrimSpace(prompt.ReadLine())

	if newPath == "" {
		ui.Warning(stdout, "Nenhum caminho informado. Configuração do vault não alterada.")
		return nil
	}

	expanded, err := config.ExpandPath(newPath)
	if err != nil {
		ui.Err(stdout, err.Error())
		return err
	}

	cfg["globalVaultPath"] = expanded
	if err := saveVaultConfig(cfg); err != nil {
		ui.Err(stdout, "Falha ao salvar configuração do vault: "+err.Error())
		return err
	}

	ui.Success(stdout, "Caminho do vault global salvo: "+expanded)
	return nil
}

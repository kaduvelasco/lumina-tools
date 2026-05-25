package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	configDir  = ".lumina"
	configFile = "config.yaml"
)

// Config holds all user-configurable settings for lumina.
type Config struct {
	WorkspacePath    string      `yaml:"workspace_path"`
	DockerComposeDir string      `yaml:"docker_compose_dir"`
	Theme            string      `yaml:"theme,omitempty"`
	FlatpakScope     string      `yaml:"flatpak_scope,omitempty"`
	Stack            StackConfig `yaml:"stack,omitempty"`
}

// StackConfig holds Docker dev stack settings populated during compose generation.
type StackConfig struct {
	PHPVersions string `yaml:"php_versions,omitempty"`
	DBUser      string `yaml:"db_user,omitempty"`
	DBPass      string `yaml:"db_pass,omitempty"`
	DBRootPass  string `yaml:"db_root_pass,omitempty"`
}

// Load reads ~/.lumina/config.yaml and returns its content.
// If the file does not exist, defaults are returned without error.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg, err := defaults()
		if err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes cfg to ~/.lumina/config.yaml atomically, creating the directory if needed.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close config: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// FlatpakFlag returns the --system or --user scope flag based on the config.
// Falls back to --system when the config cannot be loaded or the scope is unset.
func FlatpakFlag() string {
	cfg, err := Load()
	if err == nil && cfg.FlatpakScope == "user" {
		return "--user"
	}
	return "--system"
}

// ExpandPath resolves a leading ~ or ~/ against the user's home directory.
func ExpandPath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

func defaults() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("get home dir: %w", err)
	}
	return Config{
		WorkspacePath:    filepath.Join(home, "workspace"),
		DockerComposeDir: filepath.Join(home, ".lumina", "stack"),
	}, nil
}

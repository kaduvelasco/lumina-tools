package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaduvelasco/lumina-tools/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WorkspacePath == "" {
		t.Error("expected non-empty WorkspacePath default")
	}
	if cfg.DockerComposeDir == "" {
		t.Error("expected non-empty DockerComposeDir default")
	}
}

func TestSaveLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	want := &config.Config{
		WorkspacePath:    "/srv/workspace",
		DockerComposeDir: "/srv/workspace/.stack",
	}

	if err := config.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.WorkspacePath != want.WorkspacePath {
		t.Errorf("WorkspacePath: got %q, want %q", got.WorkspacePath, want.WorkspacePath)
	}
	if got.DockerComposeDir != want.DockerComposeDir {
		t.Errorf("DockerComposeDir: got %q, want %q", got.DockerComposeDir, want.DockerComposeDir)
	}
}

func TestSaveCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := config.Save(&config.Config{WorkspacePath: "/tmp/ws"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	configFile := filepath.Join(tmp, ".lumina", "config.yaml")
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("config file not found at %s: %v", configFile, err)
	}
}

func TestExpandPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	got, err := config.ExpandPath("~/projects")
	if err != nil {
		t.Fatalf("ExpandPath: %v", err)
	}
	want := filepath.Join(tmp, "projects")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandPathAbsolute(t *testing.T) {
	got, err := config.ExpandPath("/absolute/path")
	if err != nil {
		t.Fatalf("ExpandPath: %v", err)
	}
	if got != "/absolute/path" {
		t.Errorf("absolute path should be unchanged, got %q", got)
	}
}

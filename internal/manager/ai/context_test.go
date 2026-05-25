package ai

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelNames(t *testing.T) {
	tests := []struct {
		name   string
		models []Model
		want   string
	}{
		{"single", []Model{{Name: "Go"}}, "Go"},
		{"multiple", []Model{{Name: "Go"}, {Name: "PHP"}}, "Go, PHP"},
		{"empty", []Model{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := modelNames(tc.models); got != tc.want {
				t.Errorf("modelNames() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectActiveModels_EmptyDir(t *testing.T) {
	t.Chdir(t.TempDir())

	got := detectActiveModels()
	for _, m := range models {
		if got[m.Name] {
			t.Errorf("expected %q not present in empty directory", m.Name)
		}
	}
}

func TestDetectActiveModels_WithInstructionFile(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	if err := os.MkdirAll(".instructions", 0o755); err != nil {
		t.Fatal(err)
	}
	// Simulate Go model instruction file present.
	goFile := filepath.Join(".instructions", filepath.Base(models[0].Instruction))
	if err := os.WriteFile(goFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := detectActiveModels()
	if !got[models[0].Name] {
		t.Errorf("expected %q to be detected", models[0].Name)
	}
	for _, m := range models[1:] {
		if got[m.Name] {
			t.Errorf("expected %q not to be detected", m.Name)
		}
	}
}

func TestGenerateSharedFiles_CreatesExpectedFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	var buf bytes.Buffer
	active := []Model{models[0]} // Go
	if err := generateSharedFiles(active, strings.NewReader(""), &buf); err != nil {
		t.Fatalf("generateSharedFiles: %v", err)
	}

	for _, name := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md", ".windsurfrules", ".cursorrules"} {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s to be created: %v", name, err)
		}
	}
	for _, name := range []string{".aiexclude", ".claudeignore", ".geminiignore"} {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s to be created: %v", name, err)
		}
	}
}

func TestGenerateSharedFiles_ContainsInstructionReference(t *testing.T) {
	t.Chdir(t.TempDir())

	var buf bytes.Buffer
	active := []Model{models[0]} // Go — instruction: templates/instructions/GOLANG.md
	if err := generateSharedFiles(active, strings.NewReader(""), &buf); err != nil {
		t.Fatalf("generateSharedFiles: %v", err)
	}

	data, err := os.ReadFile("CLAUDE.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("GOLANG.md")) {
		t.Error("CLAUDE.md should reference GOLANG.md")
	}
}

func TestGenerateSharedFiles_MultipleModels(t *testing.T) {
	t.Chdir(t.TempDir())

	var buf bytes.Buffer
	active := []Model{models[0], models[1]} // Go + Linux Bash
	if err := generateSharedFiles(active, strings.NewReader(""), &buf); err != nil {
		t.Fatalf("generateSharedFiles: %v", err)
	}

	data, err := os.ReadFile("CLAUDE.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("GOLANG.md")) {
		t.Error("CLAUDE.md should reference GOLANG.md")
	}
	if !bytes.Contains(data, []byte("BASH.md")) {
		t.Error("CLAUDE.md should reference BASH.md")
	}
}

func TestWriteInstruction_CreatesFile(t *testing.T) {
	t.Chdir(t.TempDir())

	var buf bytes.Buffer
	m := models[0] // Go
	if err := writeInstruction(m, strings.NewReader(""), &buf); err != nil {
		t.Fatalf("writeInstruction: %v", err)
	}

	dest := filepath.Join(".instructions", filepath.Base(m.Instruction))
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("expected %s to be created: %v", dest, err)
	}
}

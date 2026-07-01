package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "''"},
		{"simple", "hello", "'hello'"},
		{"with spaces", "hello world", "'hello world'"},
		{"single quote", "it's", "'it'\\''s'"},
		{"multiple quotes", "can't stop, won't stop", "'can'\\''t stop, won'\\''t stop'"},
		{"only single quote", "'", "''\\'''"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shellQuote(tt.input); got != tt.want {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestListSQLFiles(t *testing.T) {
	t.Run("nonexistent dir", func(t *testing.T) {
		_, err := listSQLFiles("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("expected error for nonexistent dir, got nil")
		}
	})

	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()
		files, err := listSQLFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected empty slice, got %v", files)
		}
	})

	t.Run("filters non-sql files", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"backup.sql", "notes.txt", "schema.sql", "readme.md"} {
			if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		files, err := listSQLFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 .sql files, got %d: %v", len(files), files)
		}
		// Results must be in reverse (descending) alphabetical order.
		if filepath.Base(files[0]) != "schema.sql" || filepath.Base(files[1]) != "backup.sql" {
			t.Errorf("unexpected order: %v", files)
		}
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "subdir.sql"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "real.sql"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		files, err := listSQLFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 || filepath.Base(files[0]) != "real.sql" {
			t.Errorf("expected only real.sql, got %v", files)
		}
	})
}

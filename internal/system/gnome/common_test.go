package gnome

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "'hello'"},
		{"a b c", "'a b c'"},
		{"it's", `'it'\''s'`},
		{"don't stop", `'don'\''t stop'`},
		{"a\nb", "'a\nb'"},
		{"", "''"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := shellQuote(tt.in); got != tt.want {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGlobExists(t *testing.T) {
	dir := t.TempDir()

	if globExists(filepath.Join(dir, "nothing*")) {
		t.Error("expected false for non-existent glob")
	}

	f := filepath.Join(dir, "Orchis-dark")
	if err := os.Mkdir(f, 0o755); err != nil {
		t.Fatal(err)
	}
	if !globExists(filepath.Join(dir, "Orchis*")) {
		t.Error("expected true for existing glob")
	}
}

func TestIsThemeInstalled(t *testing.T) {
	td := t.TempDir()

	theme := themeEntry{Name: "Nordic", DirPattern: "Nordic"}
	if isThemeInstalled(theme, td) {
		t.Error("expected false before installation")
	}

	if err := os.Mkdir(filepath.Join(td, "Nordic"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isThemeInstalled(theme, td) {
		t.Error("expected true after creating theme dir")
	}
}

func TestIsThemeInstalledAbsPattern(t *testing.T) {
	dir := t.TempDir()
	absPattern := filepath.Join(dir, "Yaru*")

	theme := themeEntry{Name: "Yaru", DirPattern: absPattern}
	if isThemeInstalled(theme, "/ignored") {
		t.Error("expected false when no absolute path matches")
	}

	if err := os.Mkdir(filepath.Join(dir, "Yaru-blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isThemeInstalled(theme, "/ignored") {
		t.Error("expected true when absolute glob matches")
	}
}

func TestIsCursorInstalled(t *testing.T) {
	id := t.TempDir()

	ce := cursorEntry{Name: "Layan", DirPattern: "Layan-cursors"}
	if isCursorInstalled(ce, id) {
		t.Error("expected false before installation")
	}

	if err := os.Mkdir(filepath.Join(id, "Layan-cursors"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isCursorInstalled(ce, id) {
		t.Error("expected true after creating cursor dir")
	}
}

func TestIsIconInstalled(t *testing.T) {
	id := t.TempDir()

	ic := iconEntry{Name: "Kora", DirPattern: "kora"}
	if isIconInstalled(ic, id) {
		t.Error("expected false before installation")
	}

	if err := os.Mkdir(filepath.Join(id, "kora"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isIconInstalled(ic, id) {
		t.Error("expected true after creating icon dir")
	}
}

package gnome

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotGnome is returned when the current session is not running GNOME.
var ErrNotGnome = errors.New("a área de trabalho atual não é GNOME — esta opção requer GNOME")

// isGnome reports whether the current desktop session is GNOME.
func isGnome() bool {
	for _, env := range []string{"XDG_CURRENT_DESKTOP", "DESKTOP_SESSION", "GDMSESSION"} {
		v := strings.ToLower(os.Getenv(env))
		if strings.Contains(v, "gnome") || v == "ubuntu" {
			return true
		}
	}
	return false
}

func themesDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".themes"), nil
}

func iconsDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "share", "icons"), nil
}

// globExists reports whether any path matches the given pattern.
func globExists(pattern string) bool {
	matches, _ := filepath.Glob(pattern)
	return len(matches) > 0
}

// shellQuote wraps s in single quotes, escaping any existing single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

package tui

import (
	"strings"
	"testing"
)

func TestRenderChromeHeaderShowsBrandAndConnectionState(t *testing.T) {
	s := buildStyles(availableThemes[0])

	cases := []struct {
		name   string
		online bool
		want   string
	}{
		{"online", true, "Conexão com internet ativa"},
		{"offline", false, "Conexão com internet inativa"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderChromeHeader(80, tc.online, s)
			// Brand is split into separately styled segments; check the parts.
			if !strings.Contains(out, "lumina") || !strings.Contains(out, ".tools") {
				t.Errorf("renderChromeHeader(%v) missing brand 'lumina.tools':\n%s", tc.online, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("renderChromeHeader(%v) missing %q:\n%s", tc.online, tc.want, out)
			}
		})
	}
}

func TestHintsForFocusAndOverlay(t *testing.T) {
	cases := []struct {
		name    string
		focus   focusState
		overlay overlayKind
		want    []string
	}{
		{"submenu focused", focusSubmenu, overlayNone, []string{"Seção", "Sair"}},
		{"content focused", focusContent, overlayNone, []string{"Navegar", "Executar", "← Seções"}},
		{"theme overlay", focusSubmenu, overlayTheme, []string{"navegar", "confirmar", "cancelar", "sair"}},
		{"quit confirmation", focusSubmenu, overlayQuitConfirm, []string{"Confirmar", "Cancelar"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hints := hintsFor(tc.focus, tc.overlay)
			for _, want := range tc.want {
				found := false
				for _, h := range hints {
					if h.key == want || h.desc == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("hintsFor(%v, %v) missing %q in %+v", tc.focus, tc.overlay, want, hints)
				}
			}
		})
	}
}

func TestRenderChromeFooterRendersKeyBadges(t *testing.T) {
	s := buildStyles(availableThemes[0])

	out := renderChromeFooter(80, focusSubmenu, overlayNone, s)
	// Key text rendered as badge (no brackets), desc rendered separately.
	for _, want := range []string{"↑/↓", "Seção", "q", "Sair"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderChromeFooter() missing %q:\n%s", want, out)
		}
	}
}

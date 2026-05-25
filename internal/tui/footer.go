package tui

import "strings"

func renderFooter(width int, s TUIStyles) string {
	return buildFooter(width, s, "  ↑↓/jk navegar   enter/espaço selecionar   t tema   esc voltar   q sair")
}

func renderThemeFooter(width int, s TUIStyles) string {
	return buildFooter(width, s, "  ↑↓/jk navegar   enter confirmar   esc cancelar   q sair")
}

func buildFooter(width int, s TUIStyles, hints string) string {
	return s.Divider.Render(strings.Repeat("─", width)) + "\n" + s.Footer.Render(hints)
}

package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kaduvelasco/lumina-tools/internal/config"
)

// newTestModelV2 returns a ModelV2 initialised with a deterministic config.
func newTestModelV2() ModelV2 {
	return NewV2(context.Background(), &config.Config{Theme: "Tokyo Night"})
}

// pressKey sends a rune key to m and returns the updated model.
func pressKey(m ModelV2, r rune) ModelV2 {
	upd, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return upd.(ModelV2)
}

// pressSpecial sends a special key to m and returns the updated model.
func pressSpecial(m ModelV2, kt tea.KeyType) ModelV2 {
	upd, _ := m.Update(tea.KeyMsg{Type: kt})
	return upd.(ModelV2)
}

// ── section navigation (left panel) ──────────────────────────────────────────

func TestModelV2SectionNavigation(t *testing.T) {
	m := newTestModelV2() // section 0, focusSubmenu
	if m.section != 0 {
		t.Fatal("initial section must be 0")
	}

	// Down moves section forward.
	m = pressSpecial(m, tea.KeyDown)
	if m.section != 1 {
		t.Errorf("Down: section = %d, want 1", m.section)
	}
	if m.cursor != 0 {
		t.Errorf("section change resets cursor: got %d, want 0", m.cursor)
	}

	// Up moves section back.
	m = pressSpecial(m, tea.KeyUp)
	if m.section != 0 {
		t.Errorf("Up: section = %d, want 0", m.section)
	}

	// Wrap-around going up from section 0.
	m = pressSpecial(m, tea.KeyUp)
	if m.section != len(sections)-1 {
		t.Errorf("Up from 0: section = %d, want %d (wrap)", m.section, len(sections)-1)
	}
}

func TestModelV2SectionNotMovedWhenContentFocused(t *testing.T) {
	m := newTestModelV2()
	m = pressSpecial(m, tea.KeyTab) // focus content
	before := m.section
	m = pressSpecial(m, tea.KeyDown) // should move item cursor, not section
	if m.section != before {
		t.Errorf("Down in focusContent must not change section: got %d, want %d", m.section, before)
	}
}

// ── item cursor navigation (right panel) ─────────────────────────────────────

func TestModelV2ItemCursorNavigation(t *testing.T) {
	m := newTestModelV2()
	m = pressSpecial(m, tea.KeyTab) // focusContent
	n := len(sections[m.section].items)

	m = pressSpecial(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("Down: cursor = %d, want 1", m.cursor)
	}

	m = pressSpecial(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Errorf("Up: cursor = %d, want 0", m.cursor)
	}

	// Wrap-around going up from 0.
	m = pressSpecial(m, tea.KeyUp)
	if m.cursor != n-1 {
		t.Errorf("Up from 0: cursor = %d, want %d (wrap)", m.cursor, n-1)
	}
}

func TestModelV2ItemCursorNotMovedWhenSubmenuFocused(t *testing.T) {
	m := newTestModelV2() // focusSubmenu
	before := m.cursor
	m = pressSpecial(m, tea.KeyDown) // moves section, not cursor
	// cursor should still be 0 (reset by moveSection)
	if m.cursor != before {
		t.Errorf("Down in focusSubmenu must not break cursor invariant: got %d", m.cursor)
	}
}

// ── focus switching ───────────────────────────────────────────────────────────

func TestModelV2FocusToggleTab(t *testing.T) {
	m := newTestModelV2()
	if m.focus != focusSubmenu {
		t.Fatal("initial focus must be focusSubmenu")
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.focus != focusContent {
		t.Error("after Tab: focus should be focusContent")
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.focus != focusSubmenu {
		t.Error("after Tab again: focus should return to focusSubmenu")
	}
}

func TestModelV2FocusShiftTab(t *testing.T) {
	m := newTestModelV2()
	m = pressSpecial(m, tea.KeyShiftTab)
	if m.focus != focusContent {
		t.Error("ShiftTab from submenu should go to focusContent")
	}
	m = pressSpecial(m, tea.KeyShiftTab)
	if m.focus != focusSubmenu {
		t.Error("ShiftTab from content should return to focusSubmenu")
	}
}

func TestModelV2EnterInSubmenuMovesFocusToContent(t *testing.T) {
	m := newTestModelV2()
	if m.focus != focusSubmenu {
		t.Fatal("initial focus must be focusSubmenu")
	}
	m = pressSpecial(m, tea.KeyEnter)
	if m.focus != focusContent {
		t.Error("Enter in focusSubmenu should move focus to focusContent")
	}
}

func TestModelV2EscFromContent(t *testing.T) {
	m := newTestModelV2()
	m = pressSpecial(m, tea.KeyTab)
	if m.focus != focusContent {
		t.Fatal("prerequisite: focus should be on content after Tab")
	}
	m = pressSpecial(m, tea.KeyEsc)
	if m.focus != focusSubmenu {
		t.Error("Esc from content should return focus to submenu")
	}
}

func TestModelV2EscFromSubmenuNoEffect(t *testing.T) {
	m := newTestModelV2()
	m = pressSpecial(m, tea.KeyEsc)
	if m.focus != focusSubmenu || m.section != 0 {
		t.Error("Esc from submenu should be a no-op")
	}
}

// ── window resize ─────────────────────────────────────────────────────────────

func TestModelV2WindowResize(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = upd.(ModelV2)
	if m.width != 120 || m.height != 40 {
		t.Errorf("WindowSizeMsg: got %dx%d, want 120x40", m.width, m.height)
	}
}

// ── view contents ─────────────────────────────────────────────────────────────

func TestModelV2ViewContainsExpectedContent(t *testing.T) {
	m := newTestModelV2()
	view := m.View()

	// Header brand (split into styled segments; check core parts).
	if !strings.Contains(view, "lumina") || !strings.Contains(view, ".tools") {
		t.Error("View() missing brand 'lumina.tools' in header")
	}
	// All section labels visible in the sidebar.
	for _, sec := range sections {
		if !strings.Contains(view, sec.label) {
			t.Errorf("View() missing section label %q", sec.label)
		}
	}
	// Initial view shows the Sobre panel (renderSobrePanel) because section==0 and cursor==0.
	if !strings.Contains(view, "Lumina Tools") {
		t.Error("View() missing 'Lumina Tools' from Sobre panel (section 0, cursor 0)")
	}
	// Footer hints for focusSubmenu.
	if !strings.Contains(view, "Seção") {
		t.Error("View() missing footer hint 'Seção' for focusSubmenu")
	}
}

func TestModelV2ViewUpdatesOnSectionSwitch(t *testing.T) {
	m := newTestModelV2()
	// Navigate to section 2 ("Aplicativos Linux") using Down twice in focusSubmenu.
	m2 := pressSpecial(pressSpecial(m, tea.KeyDown), tea.KeyDown)
	view := m2.View()

	if !strings.Contains(view, sections[2].label) {
		t.Errorf("after switching to section 2, View() missing %q", sections[2].label)
	}
	if !strings.Contains(view, sections[2].items[0].title) {
		t.Errorf("after switching to section 2, View() missing first item %q", sections[2].items[0].title)
	}
}

func TestModelV2ViewShowsItemCmdSecondLine(t *testing.T) {
	m := newTestModelV2()            // section 0 = Home
	m = pressSpecial(m, tea.KeyTab)  // focusContent, cursor stays 0
	m = pressSpecial(m, tea.KeyDown) // cursor → 1 = "Atualizar"
	view := m.View()
	// cursor != 0 → renderItemsList shown; "Atualizar" has cmd "lumina self-update".
	if !strings.Contains(view, "lumina self-update") {
		t.Error("View() should show cmd as second line for 'Atualizar'")
	}
}

func TestModelV2ViewShowsDescForItemWithoutCmd(t *testing.T) {
	m := newTestModelV2() // section 0 = Home
	// Navigate to "Sair" (last item in Home) in the right panel.
	m = pressSpecial(m, tea.KeyTab) // focusContent
	n := len(sections[0].items)
	for range n - 1 {
		m = pressSpecial(m, tea.KeyDown)
	}
	sair := sections[0].items[m.cursor]
	if sair.title != "Sair" {
		t.Fatalf("expected cursor on 'Sair', got %q", sair.title)
	}
	view := m.View()
	if !strings.Contains(view, sair.desc) {
		t.Errorf("View() should show desc %q as second line for 'Sair'", sair.desc)
	}
}

func TestModelV2ViewFooterChangesWithFocus(t *testing.T) {
	m := newTestModelV2()
	viewSubmenu := m.View()

	m = pressSpecial(m, tea.KeyTab) // focusContent
	viewContent := m.View()

	if !strings.Contains(viewSubmenu, "Seção") {
		t.Error("focusSubmenu footer should contain 'Seção'")
	}
	if !strings.Contains(viewContent, "Executar") {
		t.Error("focusContent footer should contain 'Executar'")
	}
}

// ── action dispatch ───────────────────────────────────────────────────────────

func TestModelV2ActionDoneMsgNilClearsMsg(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(actionDoneMsg{err: nil})
	m2 := upd.(ModelV2)
	if m2.msgKind != msgNone {
		t.Errorf("actionDoneMsg{nil}: msgKind = %v, want msgNone (success shown by the script)", m2.msgKind)
	}
	if m2.msg != "" {
		t.Errorf("actionDoneMsg{nil}: msg = %q, want empty", m2.msg)
	}
}

func TestModelV2ActionDoneMsgErrSetsError(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(actionDoneMsg{err: fmt.Errorf("falha simulada")})
	m2 := upd.(ModelV2)
	if m2.msgKind != msgError {
		t.Errorf("actionDoneMsg{err}: msgKind = %v, want msgError", m2.msgKind)
	}
	if !strings.Contains(m2.msg, "falha simulada") {
		t.Errorf("actionDoneMsg{err}: msg = %q, want to contain error text", m2.msg)
	}
}

func TestModelV2NotImplementedMsgSetsWarning(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(notImplementedMsg{})
	m2 := upd.(ModelV2)
	if m2.msgKind != msgWarning {
		t.Errorf("notImplementedMsg: msgKind = %v, want msgWarning", m2.msgKind)
	}
	if m2.msg == "" {
		t.Error("notImplementedMsg: msg should be non-empty")
	}
}

func TestModelV2KeyPressClearsMsg(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(actionDoneMsg{err: fmt.Errorf("erro simulado")})
	m = upd.(ModelV2)
	if m.msg == "" {
		t.Fatal("prerequisite: msg should be set after actionDoneMsg with error")
	}
	m = pressSpecial(m, tea.KeyDown)
	if m.msg != "" {
		t.Errorf("key press should clear msg, got %q", m.msg)
	}
	if m.msgKind != msgNone {
		t.Errorf("key press should reset msgKind to msgNone, got %v", m.msgKind)
	}
}

func TestModelV2EnterOnActNoneItem(t *testing.T) {
	m := newTestModelV2()
	// "Sobre" (cursor 0, section 0) has actNone.
	m = pressSpecial(m, tea.KeyTab) // focusContent, cursor 0 = "Sobre"
	if sections[m.section].items[m.cursor].action != actNone {
		t.Fatal("prerequisite: cursor should be on actNone item")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter on actNone item should return nil cmd")
	}
}

func TestModelV2EnterOnItemWithAction(t *testing.T) {
	m := newTestModelV2()
	// Navigate to "Atualizar" (cursor 1, actLuminaUpdate) in Home.
	m = pressSpecial(m, tea.KeyTab)  // focusContent
	m = pressSpecial(m, tea.KeyDown) // cursor → 1 = "Atualizar"
	if sections[m.section].items[m.cursor].action == actNone {
		t.Fatal("prerequisite: cursor should be on an item with a non-None action")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter on item with action should return a non-nil cmd")
	}
}

func TestModelV2StatusLineAppearsInView(t *testing.T) {
	m := newTestModelV2()
	upd, _ := m.Update(actionDoneMsg{err: fmt.Errorf("erro simulado")})
	m = upd.(ModelV2)
	view := m.View()
	if !strings.Contains(view, m.msg) {
		t.Errorf("View() should contain status message %q", m.msg)
	}
}

func TestModelV2StatusLineAbsentWhenNoMsg(t *testing.T) {
	m := newTestModelV2()
	if m.msg != "" {
		t.Fatal("prerequisite: msg must be empty on init")
	}
	view := m.View()
	if strings.Contains(view, "Concluido") {
		t.Error("View() should not contain status message when msg is empty")
	}
}

// ── connectivity ──────────────────────────────────────────────────────────────

func TestModelV2ConnectivityMsgUpdatesOnline(t *testing.T) {
	m := newTestModelV2()
	if !m.online {
		t.Fatal("prerequisite: online must be true on init")
	}
	upd, _ := m.Update(connectivityMsg{online: false})
	m = upd.(ModelV2)
	if m.online {
		t.Error("connectivityMsg{false}: online should be false")
	}
	upd, _ = m.Update(connectivityMsg{online: true})
	m = upd.(ModelV2)
	if !m.online {
		t.Error("connectivityMsg{true}: online should be true")
	}
}

func TestModelV2ConnectivityTickReturnsCmd(t *testing.T) {
	m := newTestModelV2()
	_, cmd := m.Update(connectivityTickMsg{})
	if cmd == nil {
		t.Error("connectivityTickMsg should return a non-nil cmd")
	}
}

// ── quit confirmation overlay ─────────────────────────────────────────────────

func TestModelV2QKeyOpensQuitConfirm(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 'q')
	if !m.quitConfirmOpen {
		t.Error("pressing q should open quit confirm overlay")
	}
}

func TestModelV2CtrlCQuitsImmediately(t *testing.T) {
	m := newTestModelV2()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return a quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("ctrl+c cmd produced %T, want tea.QuitMsg", msg)
	}
}

func TestModelV2QuitOverlayEnterQuits(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 'q')
	if !m.quitConfirmOpen {
		t.Fatal("prerequisite: quit overlay must be open")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter in quit overlay should return a quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Enter in quit overlay produced %T, want tea.QuitMsg", msg)
	}
}

func TestModelV2QuitOverlayEscCancels(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 'q')
	m = pressSpecial(m, tea.KeyEsc)
	if m.quitConfirmOpen {
		t.Error("Esc in quit overlay should close it")
	}
}

func TestModelV2QuitOverlayQCancels(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 'q')
	m2 := pressKey(m, 'q')
	if m2.quitConfirmOpen {
		t.Error("pressing q in quit overlay should close it")
	}
}

func TestModelV2QuitOverlayRendersInView(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 'q')
	view := m.View()
	if !strings.Contains(view, "Deseja encerrar") {
		t.Error("quit overlay view should contain quit confirmation prompt")
	}
}

// ── theme overlay ─────────────────────────────────────────────────────────────

func TestModelV2TKeyOpensThemeOverlay(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 't')
	if !m.themeOpen {
		t.Error("pressing t should open theme overlay")
	}
}

func TestModelV2ThemeOverlayCursorStartsAtCurrentTheme(t *testing.T) {
	m := newTestModelV2() // theme = "Tokyo Night"
	m = pressKey(m, 't')
	if availableThemes[m.themeCursor].Name != "Tokyo Night" {
		t.Errorf("theme cursor should point to current theme %q, got %q",
			"Tokyo Night", availableThemes[m.themeCursor].Name)
	}
}

func TestModelV2ThemeOverlayEscRestoresTheme(t *testing.T) {
	m := newTestModelV2()
	original := m.theme.Name
	m = pressKey(m, 't')
	m = pressSpecial(m, tea.KeyDown)
	m = pressSpecial(m, tea.KeyEsc)
	if m.themeOpen {
		t.Error("Esc should close theme overlay")
	}
	if m.theme.Name != original {
		t.Errorf("Esc should restore original theme %q, got %q", original, m.theme.Name)
	}
}

func TestModelV2ThemeOverlayEnterConfirmsTheme(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 't')
	m = pressSpecial(m, tea.KeyDown)
	selectedName := availableThemes[m.themeCursor].Name
	m = pressSpecial(m, tea.KeyEnter)
	if m.themeOpen {
		t.Error("Enter should close theme overlay")
	}
	if m.theme.Name != selectedName {
		t.Errorf("Enter should confirm selected theme %q, got %q", selectedName, m.theme.Name)
	}
}

func TestModelV2ThemeOverlayCursorWraps(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 't')
	for m.themeCursor != 0 {
		m = pressSpecial(m, tea.KeyUp)
	}
	m = pressSpecial(m, tea.KeyUp)
	if m.themeCursor != len(availableThemes)-1 {
		t.Errorf("theme cursor should wrap to last, got %d", m.themeCursor)
	}
}

func TestModelV2ThemeOverlayRendersInView(t *testing.T) {
	m := newTestModelV2()
	m = pressKey(m, 't')
	view := m.View()
	for _, t2 := range availableThemes {
		if !strings.Contains(view, t2.Name) {
			t.Errorf("theme overlay view missing theme %q", t2.Name)
		}
	}
}

func TestModelV2KeysBlockedByThemeOverlay(t *testing.T) {
	m := newTestModelV2()
	before := m.section
	m = pressKey(m, 't')
	// Down inside the theme overlay moves the theme cursor, not the section.
	m = pressSpecial(m, tea.KeyDown)
	if m.section != before {
		t.Error("Down key should not change section while theme overlay is open")
	}
}

func TestModelV2KeysBlockedByQuitOverlay(t *testing.T) {
	m := newTestModelV2()
	before := m.section
	m = pressKey(m, 'q')
	// Down inside the quit overlay is ignored (not mapped in updateQuitOverlay).
	m = pressSpecial(m, tea.KeyDown)
	if m.section != before {
		t.Error("Down key should not change section while quit overlay is open")
	}
}

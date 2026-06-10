package tui

import (
	"testing"
)

// TestAllEntriesHaveSingleAction verifies that every submenuEntry has exactly
// one action field (the flat-list design invariant).
func TestAllEntriesHaveSingleAction(t *testing.T) {
	for si, sec := range sections {
		for ii, entry := range sec.items {
			if entry.title == "" {
				t.Errorf("section[%d](%q) item[%d]: empty title", si, sec.label, ii)
			}
			// action field is always present (zero value actNone is valid).
			_ = entry.action
		}
	}
}

// TestItemsWithoutCmdHaveDesc verifies that any entry with an empty cmd
// provides a non-empty desc as the second-line fallback.
func TestItemsWithoutCmdHaveDesc(t *testing.T) {
	for si, sec := range sections {
		for ii, entry := range sec.items {
			if entry.cmd == "" && entry.desc == "" {
				t.Errorf(
					"section[%d](%q) item[%d](%q): both cmd and desc are empty — second line would be blank",
					si, sec.label, ii, entry.title,
				)
			}
		}
	}
}

// TestActNoneItemsAreInfoOnly verifies that items with actNone exist and
// that "Sobre" (cmd present) and "Sair" (cmd absent, desc present) satisfy
// the expected second-line rule.
func TestActNoneItemsAreInfoOnly(t *testing.T) {
	// "Sobre" is section 0, item 0 — has cmd, no action needed.
	sobre := sections[0].items[0]
	if sobre.title != "Sobre" {
		t.Fatalf("expected sections[0].items[0] to be 'Sobre', got %q", sobre.title)
	}
	if sobre.cmd == "" {
		t.Error("'Sobre' should have a non-empty cmd")
	}
	if sobre.action != actNone {
		t.Errorf("'Sobre' action = %v, want actNone", sobre.action)
	}

	// "Sair" is the last item in Home — no cmd, has desc.
	sair := sections[0].items[len(sections[0].items)-1]
	if sair.title != "Sair" {
		t.Fatalf("expected last Home item to be 'Sair', got %q", sair.title)
	}
	if sair.cmd != "" {
		t.Errorf("'Sair' should have empty cmd, got %q", sair.cmd)
	}
	if sair.desc == "" {
		t.Error("'Sair' should have a non-empty desc as second-line fallback")
	}
	if sair.action != actQuit {
		t.Errorf("'Sair' action = %v, want actQuit", sair.action)
	}
}

// TestSectionCounts verifies the expected item counts for each of the 9 sections.
func TestSectionCounts(t *testing.T) {
	want := map[string]int{
		"Home":                        6,
		"Gerenciamento Linux":          8,
		"Aplicativos Linux":            3,
		"Personalizar Linux":           5,
		"Ambiente de Desenvolvimento":  9,
		"Gerenciar Stack PHP":          6,
		"Gerenciar banco de Dados":     5,
		"Gerenciar Repositórios":       6,
		"Gerenciar Contextos IA":       2,
	}
	for _, sec := range sections {
		n, ok := want[sec.label]
		if !ok {
			t.Errorf("unexpected section label %q", sec.label)
			continue
		}
		if len(sec.items) != n {
			t.Errorf("section %q: got %d items, want %d", sec.label, len(sec.items), n)
		}
	}
}

// TestUniqueActionsPerSection verifies that no section contains two items with
// the same non-placeholder action. actPending is intentionally shared among all
// pending items (their pending bool gates dispatch, not the action value).
func TestUniqueActionsPerSection(t *testing.T) {
	for _, sec := range sections {
		seen := map[actionID]string{}
		for _, entry := range sec.items {
			if entry.action == actNone {
				continue
			}
			if prev, dup := seen[entry.action]; dup {
				t.Errorf("section %q: action %v used by both %q and %q",
					sec.label, entry.action, prev, entry.title)
			}
			seen[entry.action] = entry.title
		}
	}
}

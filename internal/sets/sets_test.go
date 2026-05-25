package sets_test

import (
	"testing"

	"github.com/kaduvelasco/lumina-tools/internal/sets"
)

func TestOf(t *testing.T) {
	s := sets.Of([]string{"a", "b", "c"})
	for _, item := range []string{"a", "b", "c"} {
		if !s[item] {
			t.Errorf("expected %q to be present", item)
		}
	}
	if s["d"] {
		t.Error("unexpected item 'd' in set")
	}
}

func TestOfEmpty(t *testing.T) {
	s := sets.Of(nil)
	if len(s) != 0 {
		t.Errorf("expected empty set, got len %d", len(s))
	}
}

func TestOfDuplicates(t *testing.T) {
	s := sets.Of([]string{"x", "x", "x"})
	if len(s) != 1 {
		t.Errorf("expected 1 entry for duplicates, got %d", len(s))
	}
	if !s["x"] {
		t.Error("expected 'x' to be present")
	}
}

func TestOfSingleItem(t *testing.T) {
	s := sets.Of([]string{"only"})
	if len(s) != 1 || !s["only"] {
		t.Errorf("unexpected set contents: %v", s)
	}
}

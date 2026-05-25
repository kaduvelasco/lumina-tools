package prompt_test

import (
	"testing"

	"github.com/kaduvelasco/lumina-tools/internal/prompt"
)

func TestParseSelection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  []int
	}{
		{"single valid", "1", 3, []int{0}},
		{"last valid", "3", 3, []int{2}},
		{"multiple valid", "1 3", 3, []int{0, 2}},
		{"all items", "1 2 3", 3, []int{0, 1, 2}},
		{"zero out of range", "0", 3, nil},
		{"above max out of range", "4", 3, nil},
		{"mixed valid and invalid", "0 2 4", 3, []int{1}},
		{"duplicates deduplicated", "2 2 2", 3, []int{1}},
		{"non-numeric tokens skipped", "abc 1 xyz", 3, []int{0}},
		{"empty input", "", 3, nil},
		{"whitespace only", "   ", 3, nil},
		{"order preserved", "3 1 2", 3, []int{2, 0, 1}},
		{"max one", "1", 1, []int{0}},
		{"max zero allows nothing", "1", 0, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := prompt.ParseSelection(tc.input, tc.max)
			if len(got) != len(tc.want) {
				t.Fatalf("ParseSelection(%q, %d) = %v, want %v", tc.input, tc.max, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("index %d: got %d, want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

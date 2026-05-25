package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadLineFrom reads a single line from r.
// Returns the text and any scanner error (nil on clean EOF with no content).
func ReadLineFrom(r io.Reader) (string, error) {
	s := bufio.NewScanner(r)
	if s.Scan() {
		return s.Text(), nil
	}
	return "", s.Err()
}

// ReadLine reads a single line from stdin.
func ReadLine() string {
	line, _ := ReadLineFrom(os.Stdin)
	return line
}

// ReadPassword prints label to w and reads a password from stdin without echo.
// Falls back to plain readline when stdin is not a terminal (e.g. pipes or tests).
func ReadPassword(w io.Writer, label string) (string, error) {
	fmt.Fprint(w, label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		p := strings.TrimSpace(ReadLine())
		fmt.Fprintln(w)
		if p == "" {
			return "", fmt.Errorf("senha nao pode ser vazia")
		}
		return p, nil
	}
	fmt.Fprintln(w)
	p := strings.TrimSpace(string(b))
	if p == "" {
		return "", fmt.Errorf("senha nao pode ser vazia")
	}
	return p, nil
}

// ParseSelection parses a space-separated list of 1-based indexes.
// Returns 0-based indexes in order entered, deduplicated, clamped to [0, max).
func ParseSelection(line string, max int) []int {
	var result []int
	seen := make(map[int]bool)
	for _, tok := range strings.Fields(line) {
		var n int
		if _, err := fmt.Sscan(tok, &n); err != nil {
			continue
		}
		if n < 1 || n > max || seen[n] {
			continue
		}
		seen[n] = true
		result = append(result, n-1)
	}
	return result
}

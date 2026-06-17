package stackconfig

import (
	"strings"
	"testing"
)

func TestBashSingleQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple path", "/home/user/workspace", `'/home/user/workspace'`},
		{"path with single quote", "/home/d'artagnan/workspace", `'/home/d'\''artagnan/workspace'`},
		{"multiple single quotes", "it's a test's path", `'it'\''s a test'\''s path'`},
		{"empty string", "", `''`},
		{"already safe", "php82", `'php82'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bashSingleQuote(tt.input); got != tt.want {
				t.Errorf("bashSingleQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildWrapperScript_Shebang(t *testing.T) {
	script := buildWrapperScript("php82", "php", "/home/user/workspace/www/html")
	if !strings.HasPrefix(script, "#!/usr/bin/env bash\n") {
		t.Errorf("script does not start with bash shebang:\n%s", script)
	}
}

func TestBuildWrapperScript_ContainerAndTool(t *testing.T) {
	tests := []struct {
		container string
		tool      string
	}{
		{"php81", "php"},
		{"php82", "phpcs"},
		{"php83", "phpunit"},
		{"php84", "composer"},
	}
	for _, tt := range tests {
		t.Run(tt.container+"/"+tt.tool, func(t *testing.T) {
			script := buildWrapperScript(tt.container, tt.tool, "/workspace/www/html")
			if !strings.Contains(script, bashSingleQuote(tt.container)) {
				t.Errorf("script does not contain container %q", tt.container)
			}
			// Tool must appear in exec lines (at least twice: -it and -i branch)
			count := strings.Count(script, " "+tt.tool+" ")
			if count < 2 {
				t.Errorf("expected tool %q at least twice in exec lines, got %d", tt.tool, count)
			}
		})
	}
}

func TestBuildWrapperScript_TTYCheckUsesBothFDs(t *testing.T) {
	script := buildWrapperScript("php82", "php", "/workspace/www/html")

	// Must check stdin (fd 0) AND stdout (fd 1) — not stdout alone.
	if !strings.Contains(script, "[ -t 0 ] && [ -t 1 ]") {
		t.Error("TTY check must test both stdin (fd 0) and stdout (fd 1)")
	}

	// Must NOT use the broken single-fd check.
	if strings.Contains(script, "if [ -t 1 ]") {
		t.Error("script must not check only stdout ([ -t 1 ]) — breaks piped stdin")
	}
}

func TestBuildWrapperScript_PathTranslation_TrailingSlash(t *testing.T) {
	script := buildWrapperScript("php82", "php", "/workspace/www/html")

	// Path prefix match must require a separator after WS_HOST to prevent
	// /workspace/www/html_extra from matching /workspace/www/html.
	if !strings.Contains(script, `"${WS_HOST}/"*`) {
		t.Error(`path match must use "${WS_HOST}/"* to avoid prefix collision`)
	}
}

func TestBuildWrapperScript_PathTranslation_ExactMatch(t *testing.T) {
	script := buildWrapperScript("php82", "php", "/workspace/www/html")

	// Exact match branch must be present so that passing the directory itself
	// (without trailing slash) is also translated.
	if !strings.Contains(script, `"${WS_HOST}"`) {
		t.Error(`script must handle exact match of WS_HOST (arg == WS_HOST)`)
	}
}

func TestBuildWrapperScript_SingleQuoteInWorkspace(t *testing.T) {
	wsHTML := "/home/d'artagnan/workspace/www/html"
	script := buildWrapperScript("php82", "php", wsHTML)

	// The generated script must be syntactically valid: single quotes properly escaped.
	escaped := `'/home/d'\''artagnan/workspace/www/html'`
	if !strings.Contains(script, escaped) {
		t.Errorf("script does not contain properly escaped workspace path\nwant: %s\nscript:\n%s", escaped, script)
	}

	// Raw unescaped single quote must not appear inside a single-quoted context.
	// A naive embedding would produce: WS_HOST='/home/d'artagnan/...' which is broken.
	if strings.Contains(script, "WS_HOST='/home/d'artagnan") {
		t.Error("script contains unescaped single quote in WS_HOST — bash syntax broken")
	}
}

func TestWriteToolWrappers_EmptyVersions(t *testing.T) {
	// Must not panic when versions is empty.
	var buf strings.Builder
	writeToolWrappers(nil, "/workspace", &buf)
	writeToolWrappers([]string{}, "/workspace", &buf)
}

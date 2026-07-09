package stackconfig

import (
	"strings"
	"testing"
)

func TestHasPHPAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		min      int
		want     bool
	}{
		{"only 8.1, needs 82", []string{"8.1"}, 82, false},
		{"8.1 and 8.2, needs 82", []string{"8.1", "8.2"}, 82, true},
		{"only 8.4, needs 82", []string{"8.4"}, 82, true},
		{"empty versions", nil, 82, false},
		{"exact match", []string{"8.2"}, 82, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasPHPAtLeast(tt.versions, tt.min); got != tt.want {
				t.Errorf("hasPHPAtLeast(%v, %d) = %v, want %v", tt.versions, tt.min, got, tt.want)
			}
		})
	}
}

func TestMoodleDefaultPHP(t *testing.T) {
	tests := []struct {
		name          string
		versions      []string
		wantContainer string
		wantSuffix    string
	}{
		{"8.2 first, matches itself", []string{"8.2", "8.3"}, "php82", "82"},
		{"8.1 first, skips to lowest compatible", []string{"8.1", "8.2"}, "php82", "82"},
		{"8.1 first, only 8.3/8.4 compatible", []string{"8.1", "8.4", "8.3"}, "php83", "83"},
		{"single compatible version", []string{"8.4"}, "php84", "84"},
		{"none compatible, falls back to versions[0]", []string{"8.1"}, "php81", "81"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContainer, gotSuffix := moodleDefaultPHP(tt.versions)
			if gotContainer != tt.wantContainer || gotSuffix != tt.wantSuffix {
				t.Errorf("moodleDefaultPHP(%v) = (%q, %q), want (%q, %q)",
					tt.versions, gotContainer, gotSuffix, tt.wantContainer, tt.wantSuffix)
			}
		})
	}
}

func TestMoodleURLPrefix(t *testing.T) {
	tests := []struct {
		name       string
		workspace  string
		moodleDir  string
		wantPrefix string
		wantOK     bool
	}{
		{"dir is html root", "/workspace", "/workspace/www/html", "", true},
		{"direct subfolder", "/workspace", "/workspace/www/html/mdle", "mdle", true},
		{"nested subfolder", "/workspace", "/workspace/www/html/mdle/sub", "mdle/sub", true},
		{"outside workspace", "/workspace", "/srv/other/mdle", "", false},
		{"sibling of html root", "/workspace", "/workspace/www/data", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotOK := moodleURLPrefix(tt.workspace, tt.moodleDir)
			if gotPrefix != tt.wantPrefix || gotOK != tt.wantOK {
				t.Errorf("moodleURLPrefix(%q, %q) = (%q, %v), want (%q, %v)",
					tt.workspace, tt.moodleDir, gotPrefix, gotOK, tt.wantPrefix, tt.wantOK)
			}
		})
	}
}

func TestMoodleURLPath(t *testing.T) {
	tests := []struct {
		name      string
		urlPrefix string
		folder    string
		want      string
	}{
		{"no prefix", "", "dev-501", "/dev-501/"},
		{"with prefix", "mdle", "dev-501", "/mdle/dev-501/"},
		{"nested prefix", "mdle/sub", "core-501", "/mdle/sub/core-501/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := moodleURLPath(tt.urlPrefix, tt.folder); got != tt.want {
				t.Errorf("moodleURLPath(%q, %q) = %q, want %q", tt.urlPrefix, tt.folder, got, tt.want)
			}
		})
	}
}

func TestBuildMoodleLocations_Empty(t *testing.T) {
	if got := buildMoodleLocations(nil, "mdle", `\.php$`, moodleDefaultDispatch); got != "" {
		t.Errorf("buildMoodleLocations(nil, ...) = %q, want empty string", got)
	}
	if got := buildMoodleLocations([]string{}, "mdle", `\.php$`, moodleDefaultDispatch); got != "" {
		t.Errorf("buildMoodleLocations([], ...) = %q, want empty string", got)
	}
}

func TestBuildMoodleLocations_SingleInstall(t *testing.T) {
	got := buildMoodleLocations([]string{"dev-501"}, "mdle", `\.php$`, moodleDefaultDispatch)

	if !strings.Contains(got, "location ^~ /mdle/dev-501/ {") {
		t.Errorf("missing ^~ prefix location for dev-501:\n%s", got)
	}
	if !strings.Contains(got, `rewrite ^/mdle/dev-501/(.*)$ /mdle/dev-501/public/$1 break;`) {
		t.Errorf("missing outer rewrite into public/ before try_files touches disk:\n%s", got)
	}
	if !strings.Contains(got, `rewrite ^/mdle/dev-501/(?!public/)(.*)$ /mdle/dev-501/public/$1 break;`) {
		t.Errorf("missing nested rewrite inside the php location — direct .php requests (install.php, index.php, ...) never reach the outer rewrite, since nginx matches the nested regex location directly:\n%s", got)
	}
	if !strings.Contains(got, "try_files $uri /mdle/dev-501/public/r.php$is_args$args;") {
		t.Errorf("missing try_files fallback to r.php:\n%s", got)
	}
	if strings.Contains(got, "try_files $uri $uri/") {
		t.Errorf("try_files should not include the $uri/ directory-match alternative (would resolve to public/index.php directly instead of the router):\n%s", got)
	}
	if !strings.Contains(got, "fastcgi_pass {{MOODLE_PHP}}:9000;") {
		t.Errorf("missing default dispatch body:\n%s", got)
	}
}

func TestBuildMoodleLocations_MultipleInstalls(t *testing.T) {
	got := buildMoodleLocations([]string{"dev-501", "core-501"}, "mdle", `[^/]\.php(/|$)`, moodleVersionedDispatch)

	if !strings.Contains(got, "location ^~ /mdle/dev-501/ {") {
		t.Errorf("missing block for dev-501:\n%s", got)
	}
	if !strings.Contains(got, "location ^~ /mdle/core-501/ {") {
		t.Errorf("missing block for core-501:\n%s", got)
	}
	if !strings.Contains(got, "$php_upstream") {
		t.Errorf("missing versioned dispatch body:\n%s", got)
	}
}

func TestBuildNginxConf_NoInstalls(t *testing.T) {
	got := buildNginxConf([]string{"8.2", "8.4"}, nil, "mdle")

	if strings.Contains(got, "{{") {
		t.Errorf("unresolved placeholder left in generated config:\n%s", got)
	}
	if !strings.Contains(got, "fastcgi_pass php82:9000;") {
		t.Errorf("missing default PHP dispatch in fixed server block:\n%s", got)
	}
	if !strings.Contains(got, "set $p_ver 82;") {
		t.Errorf("missing default PHP version fallback in versioned server block:\n%s", got)
	}
	if strings.Contains(got, "location ^~") {
		t.Errorf("no Moodle installs marked, but a ^~ location block was generated:\n%s", got)
	}
}

func TestBuildNginxConf_WithInstalls(t *testing.T) {
	got := buildNginxConf([]string{"8.2"}, []string{"dev-501", "core-501"}, "mdle")

	if strings.Contains(got, "{{") {
		t.Errorf("unresolved placeholder left in generated config:\n%s", got)
	}
	if strings.Count(got, "location ^~ /mdle/dev-501/ {") != 2 {
		t.Errorf("expected one ^~ block for dev-501 in each of the two server{} blocks:\n%s", got)
	}
	if strings.Count(got, "location ^~ /mdle/core-501/ {") != 2 {
		t.Errorf("expected one ^~ block for core-501 in each of the two server{} blocks:\n%s", got)
	}
}

func TestBuildNginxConf_MoodleFallbackDecoupledFromLegacyDefault(t *testing.T) {
	// versions[0] is 8.1 — the legacy default must stay on it, but the Moodle
	// blocks must never fall back below 8.2, regardless of selection order.
	got := buildNginxConf([]string{"8.1", "8.3"}, []string{"dev-501"}, "mdle")

	if strings.Contains(got, "{{") {
		t.Errorf("unresolved placeholder left in generated config:\n%s", got)
	}
	if !strings.Contains(got, "fastcgi_pass php81:9000;") {
		t.Errorf("legacy fixed-server dispatch should still use the first selected version (php81):\n%s", got)
	}
	if !strings.Contains(got, "set $p_ver 81;") {
		t.Errorf("legacy versioned-server fallback should still use the first selected version (81):\n%s", got)
	}
	if !strings.Contains(got, "fastcgi_pass php83:9000;") {
		t.Errorf("Moodle fixed-server dispatch should fall back to the lowest compatible version (php83):\n%s", got)
	}
	if !strings.Contains(got, "set $p_ver 83;") {
		t.Errorf("Moodle versioned-server fallback should use the lowest compatible version (83):\n%s", got)
	}
}

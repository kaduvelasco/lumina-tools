package distro

import "testing"

func TestDetectDebian(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"ubuntu", "ID=ubuntu\nID_LIKE=debian"},
		{"linuxmint", "ID=linuxmint\nID_LIKE=ubuntu"},
		{"zorin", "ID=zorin\nID_LIKE=ubuntu"},
		{"pop", "ID=pop\nID_LIKE=\"ubuntu debian\""},
		{"kali", "ID=kali\nID_LIKE=debian"},
		{"elementary", "ID=elementary\nID_LIKE=ubuntu"},
		{"neon", "ID=neon\nID_LIKE=\"ubuntu debian\""},
		{"id_like fallback", "ID=something-unknown\nID_LIKE=ubuntu"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detect(tc.content); got != Debian {
				t.Errorf("detect(%q) = %q, want %q", tc.content, got, Debian)
			}
		})
	}
}

func TestDetectFedora(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"fedora", "ID=fedora"},
		{"rhel", "ID=rhel"},
		{"rocky", "ID=rocky\nID_LIKE=\"rhel fedora\""},
		{"almalinux", "ID=almalinux\nID_LIKE=\"rhel fedora\""},
		{"nobara", "ID=nobara\nID_LIKE=fedora"},
		{"id_like fallback", "ID=something-unknown\nID_LIKE=fedora"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detect(tc.content); got != Fedora {
				t.Errorf("detect(%q) = %q, want %q", tc.content, got, Fedora)
			}
		})
	}
}

func TestDetectArch(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"arch", "ID=arch"},
		{"manjaro", "ID=manjaro\nID_LIKE=arch"},
		{"endeavouros", "ID=endeavouros\nID_LIKE=arch"},
		{"garuda", "ID=garuda\nID_LIKE=\"arch endeavouros\""},
		{"cachyos", "ID=cachyos\nID_LIKE=arch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detect(tc.content); got != Arch {
				t.Errorf("detect(%q) = %q, want %q", tc.content, got, Arch)
			}
		})
	}
}

func TestDetectUnknown(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"opensuse", "ID=opensuse-leap"},
		{"nixos", "ID=nixos"},
		{"gentoo", "ID=gentoo"},
		{"empty", ""},
		{"no id field", "NAME=Linux\nVERSION=1.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detect(tc.content); got != Unknown {
				t.Errorf("detect(%q) = %q, want %q", tc.content, got, Unknown)
			}
		})
	}
}

func TestDetectQuotedID(t *testing.T) {
	content := `ID="ubuntu"` + "\n" + `ID_LIKE="debian"`
	if got := detect(content); got != Debian {
		t.Errorf("detect with quoted ID = %q, want %q", got, Debian)
	}
}

func TestDetectIDTakesPriorityOverIDLike(t *testing.T) {
	// ID is recognized directly; ID_LIKE should not override.
	content := "ID=arch\nID_LIKE=fedora"
	if got := detect(content); got != Arch {
		t.Errorf("ID should take priority: got %q, want %q", got, Arch)
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"Ubuntu"`, "ubuntu"},
		{`ubuntu`, "ubuntu"},
		{`  Fedora  `, "fedora"},
		{`"arch"`, "arch"},
		{`""`, ""},
	}
	for _, tc := range tests {
		got := clean(tc.input)
		if got != tc.want {
			t.Errorf("clean(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestClassify(t *testing.T) {
	debianIDs := []string{"ubuntu", "debian", "linuxmint", "pop", "zorin",
		"elementary", "neon", "kali", "raspbian", "mx", "lmde", "peppermint",
		"tuxedo", "parrot"}
	for _, id := range debianIDs {
		if got := classify(id); got != Debian {
			t.Errorf("classify(%q) = %q, want %q", id, got, Debian)
		}
	}

	fedoraIDs := []string{"fedora", "rhel", "centos", "rocky", "almalinux",
		"ol", "scientific", "nobara", "ultramarine"}
	for _, id := range fedoraIDs {
		if got := classify(id); got != Fedora {
			t.Errorf("classify(%q) = %q, want %q", id, got, Fedora)
		}
	}

	archIDs := []string{"arch", "manjaro", "endeavouros", "garuda", "artix", "cachyos"}
	for _, id := range archIDs {
		if got := classify(id); got != Arch {
			t.Errorf("classify(%q) = %q, want %q", id, got, Arch)
		}
	}

	unknownIDs := []string{"opensuse", "nixos", "gentoo", "", "something-random"}
	for _, id := range unknownIDs {
		if got := classify(id); got != Unknown {
			t.Errorf("classify(%q) = %q, want %q", id, got, Unknown)
		}
	}
}

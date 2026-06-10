package megasync

import (
	"testing"
)

func TestResolvePackage(t *testing.T) {
	tests := []struct {
		id      string
		ver     string
		wantURL string
		wantRPM bool
		wantErr bool
	}{
		{
			id:      "linuxmint",
			ver:     "22.3",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb",
		},
		{
			id:      "linuxmint",
			ver:     "22.0",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb",
		},
		{
			id:      "ubuntu",
			ver:     "24.04",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb",
		},
		{
			id:      "ubuntu",
			ver:     "26.04",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_26.04/amd64/megasync-xUbuntu_26.04_amd64.deb",
		},
		{
			id:      "zorin",
			ver:     "18.1",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb",
		},
		{
			id:      "zorin",
			ver:     "18.0",
			wantURL: "https://mega.nz/linux/repo/xUbuntu_24.04/amd64/megasync-xUbuntu_24.04_amd64.deb",
		},
		{
			id:      "fedora",
			ver:     "44",
			wantURL: "https://mega.nz/linux/repo/Fedora_44/x86_64/megasync-Fedora_44.x86_64.rpm",
			wantRPM: true,
		},
		// unsupported combinations
		{id: "ubuntu", ver: "22.04", wantErr: true},
		{id: "fedora", ver: "43", wantErr: true},
		{id: "arch", ver: "", wantErr: true},
		{id: "", ver: "", wantErr: true},
	}

	for _, tc := range tests {
		pkg, err := resolvePackage(tc.id, tc.ver)
		if tc.wantErr {
			if err == nil {
				t.Errorf("resolvePackage(%q, %q): expected error, got nil", tc.id, tc.ver)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolvePackage(%q, %q): unexpected error: %v", tc.id, tc.ver, err)
			continue
		}
		if pkg.url != tc.wantURL {
			t.Errorf("resolvePackage(%q, %q).url = %q, want %q", tc.id, tc.ver, pkg.url, tc.wantURL)
		}
		if pkg.rpm != tc.wantRPM {
			t.Errorf("resolvePackage(%q, %q).rpm = %v, want %v", tc.id, tc.ver, pkg.rpm, tc.wantRPM)
		}
	}
}

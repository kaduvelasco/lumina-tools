package distro

import (
	"os"
	"strings"
	"sync"
)

// Family constants returned by Detect.
const (
	Debian  = "debian"  // Ubuntu, Mint, Zorin, Elementary, KDE Neon, Pop!_OS, …
	Fedora  = "fedora"  // Fedora, RHEL, Rocky, AlmaLinux, CentOS, …
	Arch    = "arch"    // Arch, Manjaro, EndeavourOS, Garuda, …
	Unknown = "unknown" // OpenSUSE, Gentoo, NixOS, …
)

var (
	detectOnce     sync.Once
	detectedFamily string
)

// Detect reads /etc/os-release once per process lifetime, checks ID= first and
// ID_LIKE= as fallback, and returns the normalized family (Debian, Fedora, Arch,
// or Unknown). The result is cached after the first call.
func Detect() string {
	detectOnce.Do(func() {
		data, err := os.ReadFile("/etc/os-release")
		if err != nil {
			detectedFamily = Unknown
			return
		}
		detectedFamily = detect(string(data))
	})
	return detectedFamily
}

// detect parses os-release content and returns the normalized family.
func detect(content string) string {
	var id, idLike string
	for _, line := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(line, "ID="):
			id = clean(strings.TrimPrefix(line, "ID="))
		case strings.HasPrefix(line, "ID_LIKE="):
			idLike = clean(strings.TrimPrefix(line, "ID_LIKE="))
		}
	}

	if f := classify(id); f != Unknown {
		return f
	}
	for _, like := range strings.Fields(idLike) {
		if f := classify(like); f != Unknown {
			return f
		}
	}
	return Unknown
}

func classify(id string) string {
	switch id {
	case "ubuntu", "debian", "linuxmint", "pop", "zorin",
		"elementary", "neon", "kali", "raspbian", "mx", "lmde",
		"peppermint", "tuxedo", "parrot":
		return Debian
	case "fedora", "rhel", "centos", "rocky", "almalinux",
		"ol", "scientific", "nobara", "ultramarine":
		return Fedora
	case "arch", "manjaro", "endeavouros", "garuda", "artix", "cachyos":
		return Arch
	}
	return Unknown
}

// clean removes surrounding quotes and lowercases the value.
func clean(s string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(s), `"`))
}

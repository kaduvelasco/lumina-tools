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

	rawOnce   sync.Once
	cachedID  string
	cachedVer string
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

// RawID returns the raw ID= field from /etc/os-release (e.g. "linuxmint", "ubuntu", "fedora").
// Unlike Detect, it does not normalize to a distribution family.
// Returns an empty string if /etc/os-release cannot be read or lacks the field.
func RawID() string {
	initRaw()
	return cachedID
}

// VersionID returns the VERSION_ID= field from /etc/os-release (e.g. "24.04", "44").
// Returns an empty string if /etc/os-release cannot be read or lacks the field.
func VersionID() string {
	initRaw()
	return cachedVer
}

// initRaw parses /etc/os-release once and caches the raw ID= and VERSION_ID= fields.
func initRaw() {
	rawOnce.Do(func() {
		data, err := os.ReadFile("/etc/os-release")
		if err != nil {
			return
		}
		for _, line := range strings.Split(string(data), "\n") {
			switch {
			case strings.HasPrefix(line, "ID="):
				cachedID = clean(strings.TrimPrefix(line, "ID="))
			case strings.HasPrefix(line, "VERSION_ID="):
				cachedVer = clean(strings.TrimPrefix(line, "VERSION_ID="))
			}
		}
	})
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

// NormalizedDistro maps the raw OS ID to the lumina distro token
// (mint | zorin | ubuntu | fedora). Returns "" for unsupported distros.
func NormalizedDistro() string {
	switch RawID() {
	case "linuxmint":
		return "mint"
	case "zorin":
		return "zorin"
	case "ubuntu", "kubuntu":
		return "ubuntu"
	case "fedora":
		return "fedora"
	}
	return ""
}

// DetectDE returns the normalized desktop environment token
// (cinnamon | gnome | other). Reads XDG_CURRENT_DESKTOP with DESKTOP_SESSION
// as fallback.
func DetectDE() string {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if desktop == "" {
		desktop = strings.ToLower(os.Getenv("DESKTOP_SESSION"))
	}
	switch {
	case strings.Contains(desktop, "cinnamon"):
		return "cinnamon"
	case strings.Contains(desktop, "gnome"):
		return "gnome"
	}
	return "other"
}

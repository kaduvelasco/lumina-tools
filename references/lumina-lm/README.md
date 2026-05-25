# Lumina LM

📄 Portuguese version: see [LEIAME.md](LEIAME.md)

![Version](https://img.shields.io/badge/Version-2.0.0-blue)
![Bash](https://img.shields.io/badge/Bash-5%2B-121011?logo=gnubash)
![Platform](https://img.shields.io/badge/Platform-Linux-1793D1?logo=linux)

## Description

Lumina LM is a terminal-based toolkit for recurring Linux workstation setup and maintenance tasks. It provides guided menus for post-install automation, Flatpak management, file template creation, and system update command installation.

## Features

- Post-install routines for Linux Mint 22.3, Pop!_OS 24.04 LTS (COSMIC), CachyOS, ZorinOS 18.1 (Core), ZorinOS 18.1 (Lite / XFCE), and Fedora 44
- Flatpak application installation from a numbered menu (29 available apps)
- Flatpak application uninstallation from the list of installed apps
- File template generation in the user templates directory
- System-wide installation of the `update-system` command in `/usr/local/bin`
- Friendly privilege prompts before actions that require administrator access

## Project Structure

```text
.
├── lumina-lm.sh
└── scripts/
    ├── apps/
    ├── installers/
    ├── lib/
    ├── menus/
    ├── post-install/
    ├── system/
    └── templates/
```

## Installation

Make the scripts executable:

```bash
chmod +x lumina-lm.sh scripts/lib/*.sh scripts/menus/*.sh scripts/post-install/*.sh scripts/apps/*.sh scripts/templates/*.sh scripts/system/*.sh scripts/installers/*.sh
```

## Usage

Run the main menu:

```bash
bash lumina-lm.sh
```

Main menu options:

- `1` Run post-install routines
- `2` Create user file templates
- `3` Install Flatpak applications
- `4` Uninstall installed Flatpak applications
- `5` Install the `update-system` command globally
- `0` Exit

Inside submenus, `0` returns directly to the main menu.

## Configuration

- Run the launcher as a regular user
- The project requests `sudo` only for operations that require elevated privileges
- The `update-system` command is copied to `/usr/local/bin/update-system`
- Flatpak operations require Flatpak to be available; the installer can prepare the environment when needed

## Validation

When available, validate changed shell scripts with ShellCheck:

```bash
shellcheck --severity=warning --shell=bash --exclude=SC1091 lumina-lm.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 scripts/lib/utils.sh
```

## Changelog

### v2.0.0 — 2026-05-07

- Added post-install support for Fedora 44 (dnf, RPM Fusion, multimedia groups)
- Added post-install support for ZorinOS 18.1 Lite / XFCE (xfce4-goodies, Thunar plugins, PulseAudio)
- Added post-install support for Rhino Linux / Unicorn XFCE (nala, custom distro detection)
- Expanded Flatpak catalog from 18 to 29 apps, reorganized by category
- Fixed Pop!_OS script: added `libfuse2t64` and `ntfs-3g`
- Fixed CachyOS script: removed AUR-only package `ttf-ms-fonts` from pacman list

### v2.0.0 — 2026-05-08

- Removed Rhino Linux post-install support
- Added `dnf` support to `update-system` command (Fedora 44 compatibility)
- Fixed `add-apt-repository --no-update` incompatibility on Linux Mint 22.3
- Fixed double confirmation prompt after app install, uninstall, and file template operations
- Fixed unbound variable error in `system.sh` (`temp_file` RETURN trap persisting across function scopes)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)

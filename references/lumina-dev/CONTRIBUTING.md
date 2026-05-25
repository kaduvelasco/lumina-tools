# Contributing to LuminaDev

📄 Portuguese version: see [CONTRIBUINDO.md](CONTRIBUINDO.md)

Thank you for your interest in contributing to LuminaDev! This document describes the standards and process for contributing to the project.

---

## How to Contribute

1. Fork the repository
2. Create a branch: `git checkout -b feature/my-improvement`
3. Follow the coding standards described below
4. Ensure ShellCheck passes without warnings
5. Open a Pull Request describing what was changed and why

---

## Development Environment

**Requirements:**

- Bash 4.0+
- ShellCheck (for local linting)

**Installing ShellCheck:**

```bash
sudo apt install shellcheck   # Ubuntu/Debian
sudo dnf install ShellCheck   # Fedora
sudo pacman -S shellcheck     # Arch
```

**Running the linter locally:**

```bash
find . -name "*.sh" | xargs shellcheck --severity=warning --shell=bash --exclude=SC1091
```

**Validating syntax:**

```bash
bash -n lumina-dev.sh
bash -n scripts/utils.sh
```

---

## Code Standards

All scripts must follow the conventions defined in the project style guide. Key rules:

### Script Structure

Every script follows this exact section order:

1. Shebang: `#!/usr/bin/env bash`
2. Header comment with name, description, version
3. `set -euo pipefail`
4. `readonly SCRIPT_DIR=...`
5. Dependency guard and `source` of `utils.sh`
6. Interface functions (`show_header`)
7. Helper functions
8. Business functions
9. `main()` entry point
10. `main "$@"`

### Shell Options

Always required immediately after the shebang and header:

```bash
set -euo pipefail
```

### Variables

- All script-level constants use `readonly`
- All function-local variables use `local`
- Never use implicit global variables inside functions

### Output

Use only the standardized functions from `utils.sh` for status messages:

```bash
die "error message"      # stderr + exit 1
warn "warning message"   # stderr, non-fatal
info "info message"      # stdout, informational
success "done message"   # stdout, confirmation
```

For progress messages with custom icons, use `printf '%b\n'` instead of `echo -e`:

```bash
printf '%b\n' "${C6}⚙️  Installing...${RESET}"
```

### Menus

Interactive menus use `while true` with `break` or `return 0` to exit. Never use recursion.

### Temporary Files

Every `mktemp` must be accompanied by `trap 'rm ...' EXIT`. Always do explicit cleanup and `trap - EXIT` at the end of the function:

```bash
local tmp
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
# ... work ...
rm -f "$tmp"
trap - EXIT
```

### Package Installation

Never call `apt-get`, `dnf`, or `pacman` directly in module scripts. Use the abstractions from `utils.sh`:

```bash
detect_pkg_manager   # call at start of main()
ensure_pkg "name"    # installs if not already present
```

### Idempotency

Every installation function must check whether the tool is already installed and offer to skip or reinstall:

```bash
if is_installed_cmd "tool"; then
    printf '%b\n' "${C2}✅ tool already installed.${RESET}"
    echo -ne "   Reinstall / Update? (${C3}s${RESET}/N): "
    read -r confirm
    [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
fi
```

---

## Adding a New Script

When adding a new installer script:

1. Place it in `scripts/` (CLI tools) or `ides/` (editors)
2. Source `utils.sh` using the appropriate relative path
3. Call `detect_pkg_manager` at the start of `main()`
4. Implement `require_not_root` and `require_sudo` if the script uses `sudo`
5. Add the script to the CI verification list in `.github/workflows/lint.yml`
6. Add an entry to the main menu in `lumina-dev.sh`
7. Add a corresponding removal function in `scripts/uninstall.sh`
8. Update `README.md` and `LEIAME.md` to document the new module

---

## Pull Request Process

- Keep PRs focused on a single change or feature
- Describe what was changed and why in the PR body
- Ensure ShellCheck passes: `find . -name "*.sh" | xargs shellcheck --severity=warning --shell=bash --exclude=SC1091`
- Test on at least one supported distro before submitting

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)

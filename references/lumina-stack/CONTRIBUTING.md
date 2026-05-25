# Contributing to LuminaStack

📄 Portuguese version: see [CONTRIBUINDO.md](CONTRIBUINDO.md)

Thank you for your interest in contributing to LuminaStack! This document describes the process for reporting bugs, suggesting improvements, and submitting pull requests.

---

## Reporting Bugs

Before opening an issue, verify that:

- The bug is reproducible with the latest version
- No existing issue already covers the same problem

When opening an issue, include:

- Operating system and version (e.g., Ubuntu 24.04)
- Bash version (`bash --version`)
- Docker and Docker Compose version (`docker --version`, `docker compose version`)
- Steps to reproduce the problem
- Expected vs. actual behavior
- Relevant output or error messages

---

## Suggesting Improvements

Open an issue describing:

- The problem you are trying to solve
- Your proposed solution
- Why it would benefit other users

---

## Submitting Pull Requests

### 1. Fork and branch

```bash
git clone https://github.com/kaduvelasco/lumina-stack.git
cd lumina-stack
git checkout -b feature/my-improvement
```

### 2. Follow the project coding style

All scripts must comply with the conventions documented in the project:

- Shebang `#!/usr/bin/env bash` on the first line
- Header block with `Nome do Script`, `Descrição`, `Versão`
- `set -euo pipefail` in entry-point scripts
- Load guard (`[[ -n "${LIB_LOADED:-}" ]] && return 0`) in all library files
- All function variables declared with `local`
- Use `printf` instead of `echo -e`
- Use the output functions from `lib/utils.sh` (`die`, `warn`, `info`, `success`)
- Menus use `while true` loops — no recursion
- `mktemp` always paired with `trap ... EXIT`
- No direct calls to `apt-get`, `dnf`, or `pacman` — use `ensure_pkg` from `lib/utils.sh`

### 3. Run ShellCheck before submitting

```bash
shellcheck --severity=warning --shell=bash --exclude=SC1091 install.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 clean-docker.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/utils.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/versions.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/menu.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/system.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/workspace.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/docker.sh
```

All files must pass with zero warnings before the PR is opened.

### 4. Update `lib/versions.sh` for version changes

If your change introduces or updates a component version (PHP, Nginx, MariaDB), update the constants in `lib/versions.sh`. Do not hardcode version strings directly in scripts or templates.

### 5. Open the Pull Request

Describe in the PR:

- What problem it solves
- What was changed and why
- How to test the change

---

## Code of Conduct

Be respectful and constructive. Contributions of all experience levels are welcome.

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)

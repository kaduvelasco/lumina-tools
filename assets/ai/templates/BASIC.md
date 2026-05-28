# {{AGENT_FILE}}

This file defines the rules and conventions that AI agents must follow when working in this repository.

---

## Rule Priority

When rules conflict, follow this order:

1. **User instructions** — always take precedence.
2. **This file** — applies to all tasks.
3. Existing project conventions.
4. Default AI behavior.

Never override explicit user instructions. If something is unclear, **ask before proceeding**.

---

## Language

| Context | Language |
|---|---|
| Responses to the user | Brazilian Portuguese (pt-BR) |
| Documentation (`*.md`) | English |
| Code comments | English |

Do **not** mix languages inside the same file.

---

## Agent Behavior

- Make **minimal and precise changes**.
- Modify **only files relevant to the task**.
- Respect the **existing project structure**.
- Prefer **simple and readable** solutions.
- Avoid unnecessary refactoring or large rewrites unless explicitly requested.
- If a task requires a large refactor, **ask the user before proceeding**.

---

## Git Operations

AI agents must **never** perform or simulate Git operations. Do not:

- Run `git` commands
- Generate commit messages
- Suggest commits, branches, or pull requests

Version control is handled **manually by the user**. Agents may only **create or modify files**.

---

## Dependencies

Before adding any dependency:

1. Check if the functionality exists in the standard library.
2. Prefer built-in language features.
3. If a dependency is truly required, explain why a built-in solution is insufficient.

Dependencies must be **minimal and justified**.

---

## Code Quality

Generated code must:

- Follow existing project conventions.
- Use clear and consistent naming.
- Prioritize readability over cleverness.
- Avoid unnecessary abstractions and overengineering.
- Do not remove or disable existing tests unless explicitly requested.
- If the project has a test suite, new logic should include corresponding tests.

---

## Documentation

### GitHub Documentation

Public-facing documents published to the repository root. These follow GitHub Markdown conventions and must be kept bilingual.

**Required files:**

| File | Language | Description |
|---|---|---|
| `README.md` | English | Main project documentation |
| `LEIAME.md` | Portuguese (pt-BR) | Portuguese translation of README |
| `CONTRIBUTING.md` | English | Contribution guidelines |
| `CONTRIBUINDO.md` | Portuguese (pt-BR) | Portuguese translation of CONTRIBUTING |

**Cross-linking:** every public doc must link to its counterpart:

```md
📄 Portuguese version: see LEIAME.md   <!-- README.md / CONTRIBUTING.md -->
📄 English version: see README.md      <!-- LEIAME.md / CONTRIBUINDO.md -->
```

**README structure:**

1. Project title and description
2. Badges (language version, CI, license — no decorative badges)
3. Features
4. Installation
5. Usage
6. Configuration
7. Contributing
8. License

**Signature:** every GitHub doc (`*.md` created for the repository) must end with this block, exactly once:

```md
---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
```

### General Documentation

Internal documents for planning, architecture notes, decisions, and technical reference.

**Rules:**

- Always in **English**, regardless of other language settings.
- No signature required.
- No strict structure — use whatever format suits the content.
- Suitable for: implementation plans, architecture decisions, research notes, changelogs, meeting notes.
- Not required to follow GitHub public conventions (no bilingual copies, no README structure).

**Placement:**

- `docs/` — internal technical documents not meant for GitHub display.
- Project root — documents relevant to contributors (e.g., `CHANGELOG.md`, ADRs).

**Examples:** `docs/architecture.md`, `docs/decisions.md`, `CHANGELOG.md`, `implementation-plan.md`

---

## Security

AI agents must never:

- Expose credentials, secrets, or API keys.
- Generate or hardcode sensitive information.
- Introduce insecure patterns.

If a task appears to require sensitive information, **ask the user instead of generating it**.

---

## General Principles

- Implement **only what was requested** — avoid scope creep.
- Keep solutions **simple and maintainable**.
- Preserve the existing architecture.
- When in doubt, **ask the user before proceeding**.
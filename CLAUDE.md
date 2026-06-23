# CLAUDE.md

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

- **When in doubt, always ask.** Do not assume, guess, or proceed with uncertainty — stop and ask the user first.
- Make **minimal and precise changes**.
- Modify **only files relevant to the task**.
- Prefer **editing existing files** over creating new ones; only create a file when the task explicitly requires it.
- Respect the **existing project structure**.
- Prefer **simple and readable** solutions.
- Avoid unnecessary refactoring or large rewrites; if one is truly required, **ask the user before proceeding**.

---

## Communication

- Respond **concisely**. Match response length to request complexity — a simple question gets a direct answer, not headers and sections.
- Do not add preamble before acting ("I'll now look at X and then...") — just act.
- Do not summarize what you just did at the end of a response. The user can read the diff.
- For **exploratory questions** ("what do you think about X?", "how could we approach Y?"), respond with a short recommendation and the main tradeoff — do not implement until the user confirms.

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
- Default to **no comments**; only comment the WHY when the reason is non-obvious from the code itself.
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
- Suitable for: architecture decisions, research notes, changelogs, meeting notes.
- Not required to follow GitHub public conventions (no bilingual copies, no README structure).
- Permanent — kept as long as the project lives, not deleted when a task ends.

**Placement:**

- `docs/` — internal technical documents not meant for GitHub display.
- Project root — documents relevant to contributors (e.g., `CHANGELOG.md`, ADRs).

**Examples:** `docs/architecture.md`, `docs/decisions.md`, `CHANGELOG.md`

### Analysis, Code Review & Test/Implementation Plans

Ephemeral working documents produced while doing a task: codebase analysis, code review reports, test plans, and implementation plans.

**Rules:**

- Always in **Brazilian Portuguese (pt-BR)** — this overrides the English-only rule above for this specific category.
- No signature required (not a GitHub-facing document).
- No strict structure — use whatever format suits the content.
- **Disposable:** once the analysis/review/plan is no longer useful (the work is done, outdated, or superseded), delete the file — but **always confirm with the user before deleting**.
- Not required to follow GitHub public conventions (no bilingual copies, no README structure).

**Placement:** project root or `docs/`, wherever is most convenient for the task at hand.

**Examples:** `analise-*.md`, `code-review-*.md`, `plano-teste-*.md`, `plano-implementacao-*.md`

---

## Security

AI agents must never:

- Expose credentials, secrets, or API keys.
- Generate or hardcode sensitive information.
- Introduce insecure patterns.

If a task appears to require sensitive information, **ask the user instead of generating it**.



## Subagents

Spawn subagents to isolate context, parallelize independent work,
or offload bulk mechanical tasks.

**Spawn when:**
- Tasks are independent and have no shared reasoning.
- A subtask is purely mechanical (formatting, extraction, translation).
- Context isolation would prevent contamination between concerns.

**Do not spawn when:**
- The parent needs to hold the reasoning together.
- Synthesis requires cross-task judgment.
- Spawn overhead dominates the actual work.

**Model selection — pick the least capable model that can do the job well:**

| Capability needed | Model | API string |
|---|---|---|
| Bulk mechanical, no judgment | Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| Scoped research, code tasks, in-scope synthesis | Claude Sonnet 4.6 | `claude-sonnet-4-6` |
| Planning, tradeoffs, complex reasoning | Claude Opus 4.7 | `claude-opus-4-7` (use no orquestrador, não em subagents) |

**Escalation:** Se um subagent perceber que a tarefa excede sua capacidade,
deve retornar `{ "escalate": true, "reason": "..." }` ao pai — não tentar
compensar com raciocínio além do seu modelo.

## Language-Specific Standards

@.instructions/GOLANG.md
@.instructions/BASH.md
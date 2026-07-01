# GEMINI.md

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



---

## Subagents

Spawn subagents to isolate context, parallelize independent workflows, or offload repetitive and high-volume execution tasks.

**Spawn when:**
- Tasks are strictly independent and do not require shared reasoning or centralized context.
- A subtask is primarily mechanical (e.g., formatting, data extraction, tagging, translation, or schema conversion).
- Parallel execution significantly reduces overall latency.
- Strict context isolation is required to prevent cross-contamination between distinct concerns.
- The task requires isolated, long-running tool usage, ambient monitoring, or multi-step iterative execution.

**Do not spawn when:**
- The parent agent must maintain global reasoning coherence and state tracking.
- Synthesis across multiple subtasks requires unified, centralized judgment.
- The orchestration, prompt, and token overhead exceeds the actual execution benefit.
- The task depends heavily on evolving conversational nuance, subtle user intent, or subjective context.

**Model selection — pick the least capable model that can do the job well:**

|Capability needed|Recommended Model|
|-----------------|-----------------|
|"Bulk mechanical execution, parsing, formatting, routing, and high-speed data extraction"|Gemini 3.5 Flash Low|
|"Fast general-purpose subtasks, standard transformations, and basic tool use"|Gemini 3.5 Flash Medium|
|"Tool-heavy execution, agentic workflows, multi-step orchestration, and moderate reasoning"|Gemini 3.5 Flash High|
|"Scoped research, codebase analysis, deep repository-wide synthesis, and complex coding tasks"|Gemini 3.1 Pro|
|"Complex architecture planning, deep multi-domain reasoning, and mission-critical logic"|Gemini 3.5 Pro|

If a subtask turns out to need more capability than its assigned model, the subagent must signal that to the parent — not attempt to compensate. 

**Escalation Rule**

If a subtask exceeds the reasoning tier or execution boundaries of its assigned model, the subagent must not attempt to compensate or loop endlessly. Instead, it must immediately:

1. Halt local execution and stop further escalation attempts.
2. Report the specific capability mismatch or bottleneck back to the parent agent.
3. Return all partial findings, structured logs, and clearly defined uncertainty boundaries.



## Language-Specific Standards

@.instructions/GOLANG.md
@.instructions/BASH.md
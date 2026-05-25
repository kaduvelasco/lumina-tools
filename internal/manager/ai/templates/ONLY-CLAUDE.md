---

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

| Capability needed | Model |
|---|---|
| Bulk mechanical, no judgment | Claude Haiku 4.5 |
| Scoped research, code tasks, in-scope synthesis | Claude Sonnet 4.6 |
| Planning, tradeoffs, complex reasoning | Claude Opus 4.7 |

If a subtask turns out to need more capability than its assigned model,
the subagent must signal that to the parent — not attempt to compensate.

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
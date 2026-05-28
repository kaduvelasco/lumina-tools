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


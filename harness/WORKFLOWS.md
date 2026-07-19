<!-- HARNESS:START
     version=0.34.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-19T12:15:55Z
     DO NOT EDIT — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Workflows — awsysco-go-sdk

> Operational patterns enforced by the harness for agent `awsysco-go-sdk`.
> HITL gates and must-escalate actions are **structural blocks**, not suggestions.

---

## Standard task flow

Every non-trivial task follows this sequence:

```
1. memory_search  →  check if context already exists for this task
2. [tool calls]   →  do the work
3. memory_store   →  record outcome, decisions made, or facts learned
4. audit check    →  harness records automatically; no action needed
```

---

## HITL gates (actions that block for human approval)


The following actions are blocked until `the designated approver` approves:


### create_pr

**Triggered by:** this action requires human approval
**Approval timeout:** 30 minutes (auto-deny on expiry)

**Flow:**
1. Agent calls tool → harness holds the request, issues a HITL token
2. Notification sent to `@here` via Slack DM (if configured)
3. Approver runs: `harness-ctl hitl approve <token>`
   or denies: `harness-ctl hitl deny <token>`
4. On approve → tool executes, result returned to agent
5. On deny → agent receives: `"action denied by human reviewer"`


### deploy

**Triggered by:** this action requires human approval
**Approval timeout:** 30 minutes (auto-deny on expiry)

**Flow:**
1. Agent calls tool → harness holds the request, issues a HITL token
2. Notification sent to `@here` via Slack DM (if configured)
3. Approver runs: `harness-ctl hitl approve <token>`
   or denies: `harness-ctl hitl deny <token>`
4. On approve → tool executes, result returned to agent
5. On deny → agent receives: `"action denied by human reviewer"`


### spend

**Triggered by:** this action requires human approval
**Approval timeout:** 30 minutes (auto-deny on expiry)

**Flow:**
1. Agent calls tool → harness holds the request, issues a HITL token
2. Notification sent to `@here` via Slack DM (if configured)
3. Approver runs: `harness-ctl hitl approve <token>`
   or denies: `harness-ctl hitl deny <token>`
4. On approve → tool executes, result returned to agent
5. On deny → agent receives: `"action denied by human reviewer"`



Check pending approvals at any time: `harness-ctl pending`



---

## Delegation patterns


This agent (`awsysco-go-sdk`, trust: `worker`) can delegate to:



**Delegation constraint:** you can only grant permissions you hold yourself.
Delegating `web_search` to an agent that doesn't have it in their manifest still blocks.


---

## Memory usage patterns

### When to store
- Decisions made that affect future tasks: `memory_store(key="decision-<topic>", value="...")`
- Outcomes of completed work: `memory_store(key="result-<task>", value="...")`
- Project facts discovered during work: `memory_store(key="fact-<topic>", value="...")`
- Status that another session will need: `memory_store(key="status-<feature>", value="...")`

### When to search
- Before starting any task: `memory_search(query="<task-topic>")` — avoid duplicate work
- Before making a decision: `memory_search(query="<decision-topic>")` — check prior reasoning
- When context feels missing: search before fetching externally — may already be stored

### Namespace: `awsysco-go-sdk-worker`
All memory calls write to and read from this namespace by default.
Cross-namespace reads require explicit grant — do not assume access to other agents' memory.

---

## Error recovery

| Error | Meaning | Action |
|---|---|---|
| `401 Unauthorized` | JWT expired or wrong secret | Re-register: `POST /harness/register` |
| `403 Forbidden` | Tool not in `allowed_tools` | Check manifest — do not retry |
| `429 Too Many Requests` | Budget exceeded or rate limited | Check `harness-ctl budget awsysco-go-sdk` |
| `"hitl_denied"` | Human denied the action | Stop — do not retry automatically |
| `"private and internal URLs"` | SSRF block | Use a public URL; never try localhost |
| `"path traversal not allowed"` | file_ops path escape | Use relative paths within project root |
| `"cmd must be a plain command name"` | shell: absolute path in cmd | Use plain name: `"go"` not `"/usr/bin/go"` |
| `"shell skill not permitted"` | allow_shell is false | This tool is not available to this agent |
| `"cross-namespace read"` | memory: no grant | Only query own namespace (`awsysco-go-sdk-worker`) |

---

## Session checklist

Before starting a multi-step task:

- [ ] Token is fresh (< 15 min old) — re-register if unsure
- [ ] `memory_search` run for this task topic
- [ ] Budget checked if doing many tool calls: `harness-ctl budget awsysco-go-sdk`
- [ ] Aware of must-escalate actions for this task (see above)




---

## Notes from previous version

---

<!-- Add project-specific workflow stages, approval processes, and
     business rules below. The harness block above is managed automatically. -->



<!-- HARNESS:END -->

---

<!-- Add project-specific workflow stages, approval processes, and
     business rules below. The harness block above is managed automatically. -->

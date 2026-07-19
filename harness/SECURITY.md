<!-- HARNESS:START
     version=0.34.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-19T12:15:55Z
     DO NOT EDIT — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Security — Enforced Constraints

> These are **hard limits enforced at the harness gateway layer** — not advisory rules.
> They cannot be bypassed by agent code, prompt engineering, or tool calls.
> Violations return HTTP 4xx and are recorded in the audit log.

---

## Authentication

Every request to the harness requires a valid session JWT.

- Token issued by: `POST http://127.0.0.1:7700/harness/register`
- TTL: **15 minutes** — re-register before expiry
- Format: `Authorization: Bearer <token>`
- Replay window: 5-minute grace period; tokens older than 20 min are always rejected
- Secret used to register: `HARNESS_REGISTER_SECRET` environment variable

**If you see `401 Unauthorized`:**
1. Token may have expired — re-register to get a fresh one
2. Wrong `HARNESS_REGISTER_SECRET` — check environment variable
3. JWT was issued by a different harness instance (different `HARNESS_JWT_SECRET`)

---

## Budget limits (hard terminate, not alert)

When any limit is exceeded, the harness returns **HTTP 429** and records a `budget_exceeded` audit event. The call does not execute.

| Limit | Value | Resets |
|---|---|---|
| Steps per session | 40 | On re-register |
| Tokens per session | 80000 | On re-register |
| Cost per session | $3.00 | On re-register |

Check current usage: `GET http://127.0.0.1:7700/harness/budget` (JWT required)

---

## Rate limiting

Per-agent token bucket. Default: **10 RPS, burst 20**.
Exceeded requests return **HTTP 429** immediately (no queuing).

Adjust at runtime (admin only): `POST http://127.0.0.1:7700/harness/ratelimit/set`

---

## Circuit breaker

Auto-throttle triggers when spend exceeds **3× the 7-day average** within any 15-minute window.
State: `closed` (normal) → `open` (all requests blocked, HTTP 429) → `half-open` (probe)

Reset: `harness-ctl breaker reset awsysco-go-sdk`

---

## Output scanning

Every tool result and LLM response is scanned for **222+ secret patterns** before
being returned. Matches are **redacted**, counted in metrics, and logged as `secret_violation`.

Patterns include: Anthropic API keys, AWS credentials, GitHub tokens, Slack tokens,
private keys (RSA/EC/PEM), JWT secrets, generic API key patterns.

Agents never receive raw credentials — the vault injects them per-request.

---

## SSRF protection (http_client, web_fetch)

The following targets are **always blocked**, regardless of how the URL is constructed:

| Blocked range | Examples |
|---|---|
| Loopback | `127.x.x.x`, `::1`, `localhost` |
| RFC-1918 private | `10.x.x.x`, `172.16-31.x.x`, `192.168.x.x` |
| Link-local | `169.254.x.x`, `fe80::/10` |
| Non-http(s) schemes | `file://`, `ftp://`, `smb://`, `gopher://` |

DNS rebinding is prevented: hostnames are resolved to IPs **at TCP connect time** and
the resolved IP is re-checked against the blocklist. A DNS record that points to a
private IP after the URL check is still blocked.

---

## File access restrictions (file_ops)

All paths are relative to the project root: `/Users/patrickbertsch/dev/awsysco-go-sdk`

**Always blocked (credential / key files):**
- `.env`, `.env.*`
- `*.pem`, `*.key`, `*.pfx`, `*.p12`
- `authorized_keys`, `known_hosts`
- `.aws/credentials`, `.aws/config`
- `*.json` matching Firebase service account pattern (`*-adminsdk-*.json`)
- `*.pem, *.key, .env, .env.*`

**Path traversal:** any path resolving outside the project root is blocked.
Symlinks are resolved before the boundary check (`filepath.EvalSymlinks`).

---

## MCP tool description pinning (rug-pull prevention)

External MCP tool descriptions are SHA-256 hashed on first use.
If the description changes (rug-pull / prompt injection via tool description):
- Tool is **quarantined** — calls return `"tool quarantined: description changed"`
- Admin must review and re-approve: `harness-ctl pins approve <server> <tool>`

Built-in harness tools are immune (descriptions are compiled in).

---

## Credential isolation

Agents **never** hold long-lived API keys. The harness vault injects scoped,
short-lived tokens per-request.

This agent's vault scopes: `[map[description:(vault scope — see harness vault config) scope:github ttl:30 minutes]]`

Attempting to call a vault scope not in this list returns:
`"vault: scope <scope> not granted to agent awsysco-go-sdk"`

---

## Audit log

Every event is written to an **append-only SQLite log** — no delete, no update.

Recorded events: `agent_register`, `tool_call`, `tool_error`, `llm_call`, `llm_error`,
`hitl_request`, `hitl_approved`, `hitl_denied`, `budget_exceeded`, `secret_violation`, `a2a_task`

Query: `harness-ctl audit awsysco-go-sdk`
Or: `GET http://127.0.0.1:7700/harness/audit?agent=awsysco-go-sdk`




<!-- HARNESS:END -->

<!-- HARNESS:START
     version=0.33.0
     schema=1
     updated=2026-07-19T05:36:09Z
     DO NOT EDIT — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Infrastructure — awsysco-go-sdk

---

## Harness connection

| Field | Value |
|---|---|
| Endpoint | `http://127.0.0.1:7700` |
| Liveness probe | `GET http://127.0.0.1:7700/harness/ready` (no auth) |
| Health | `GET http://127.0.0.1:7700/harness/health` (JWT required) |
| MCP tools | `POST http://127.0.0.1:7700/mcp` |
| A2A tasks | `POST http://127.0.0.1:7700/a2a/tasks` |

Harness runs as a macOS launchd daemon — auto-starts on login, restarts on crash.
Logs: `~/Library/Logs/central-harness.log`
Manage: `harness-ctl daemon status | start | stop | restart`

---

## Deployment targets



*(no deployment targets detected — add manually)*


---

## Credential vault scopes

The harness vault holds credentials. This agent (`awsysco-go-sdk`) is granted:


- **`github`** — (vault scope — see harness vault config) (expires: 30 minutes)



Request a scoped token (JWT required):
```
POST http://127.0.0.1:7700/harness/vault/token
{"scope": "<scope>"}
→ {"token": "...", "expiresIn": 1800}
```

**Do not use these environment variables directly in tool calls** — request vault tokens instead.
The harness injects credentials per-request and they expire automatically.

---

## CI / CD pipeline


| CI system | GitHub Actions |
|---|---|
| Config file | `.github/workflows/secret-scan.yml` |
| Test command | `go test ./...` |
| Build command | `go build ./...` |
| Deploy trigger | push to main / manual workflow_dispatch |

GitHub Actions workflow runs on every push and PR.


---

## Staging vs production workflow


*(staging workflow not detected — add manually if applicable)*


---

## Infrastructure files in this repo



*(none detected)*


---

## Monitoring and observability


*(no observability config detected — add manually)*


**Harness metrics** (always available):
- `GET http://127.0.0.1:7700/metrics` — Prometheus (no auth)
- `harness-ctl metrics` — human-readable format

Key counters: `harness_tool_calls_total`, `harness_rate_limit_rejections_total`,
`harness_secret_violations_total`, `harness_cache_hits_total`




---

## Notes from previous version

---

<!-- Add infrastructure decisions, runbooks, and incident notes below.
     The harness block above is managed automatically — everything below is yours. -->



<!-- HARNESS:END -->

---

<!-- Add infrastructure decisions, runbooks, and incident notes below.
     The harness block above is managed automatically — everything below is yours. -->

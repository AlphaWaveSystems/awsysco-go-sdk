<!-- HARNESS:START
     version=0.32.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-18T02:25:54Z
     DO NOT EDIT THIS BLOCK — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Harness — Active Constraints

**This file is the entry point for every task in this project — always start here.**

**Agent:** `awsysco-go-sdk` · trust: `worker` · model: `coder`
**Budget:** 40 steps · 80000 tokens · $3.00 per session
**Privacy:** local_preferred — local models preferred; cloud only on low confidence
**Memory namespace:** `awsysco-go-sdk-worker`


## Must escalate (blocks until human approves)

- `create_pr`

- `deploy`

- `spend`



## Available tools
See `harness/TOOLS.md` for full reference with parameter schemas.

- `web_search` — search the web via Brave/Google
- `web_fetch` — fetch and extract URL content
- `file_ops` — read/write files within the project root
- `memory_store` / `memory_search` — per-session key-value memory
- `code_search` — search this project's own codebase (read-only)

## Project overrides (harness.yaml)

*(no harness.yaml found — using manifest defaults)*


<!-- HARNESS:END -->

---


# Harness — AwsyscoGoSdk

**Agent:** `awsysco-go-sdk` · trust: `worker` · model: `coder`
**Project root:** `~/dev/awsysco-go-sdk`
**Remote:** `https://github.com/AlphaWaveSystems/awsysco-go-sdk`
**Stack:** `Go`

## Startup

Before working:
1. Read this file
2. `cd ~/dev/awsysco-go-sdk`
3. Run verification: `go test ./... && go vet ./...`
4. Check `git status` and `git log --oneline -10`

## Working rules

- Branch names: `feat/awsysco-go-sdk`, `fix/awsysco-go-sdk`, `chore/awsysco-go-sdk`
- Always work in a git worktree: `git worktree add .worktrees/<branch> -b <branch>`
- Stage specific files only — never `git add .`
- Commit format: `type: description` (feat/fix/chore/refactor/docs)
- PRs required for all merges — no direct commits to main/master
- Run verification before every commit

## Verification

```bash
go test ./... && go vet ./...
```

## Definition of done

- [ ] Implementation complete and verified
- [ ] Tests pass
- [ ] PR created (or commit staged if no remote)
- [ ] No regressions in adjacent features

## Guardrails

Bounded autonomy. Escalate deploys and spend to Zeus.

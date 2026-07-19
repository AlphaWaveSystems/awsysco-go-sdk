<!-- HARNESS:START
     version=0.34.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-19T12:15:55Z
     DO NOT EDIT THIS BLOCK — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Harness — This Project

**Agent `awsysco-go-sdk`** · tier `coder` · privacy `local_preferred`

Full harness configuration, budget, escalation rules, available tools, and — critically —
how to actually connect to it, all live in **`AGENTS.md`**. Read that file, not this one,
for anything harness-related.

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

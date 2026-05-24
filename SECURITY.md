# Security Policy

## Reporting Vulnerabilities

Please report security vulnerabilities to security@awsys.co. Do **not** open a public GitHub issue.

## Secret Management Rules

This repository enforces strict secret hygiene:

### What is protected
- **Pre-commit hook** (`.githooks/pre-commit`): Blocks any commit containing `awsys_` API keys, `.env` files, or common secret patterns. Activate with `git config core.hooksPath .githooks`.
- **GitHub Actions** (`.github/workflows/secret-scan.yml`): Every push and PR is scanned with [gitleaks](https://github.com/gitleaks/gitleaks).
- **GitHub secret scanning**: Enabled automatically on this public repository.
- **`.gitignore`**: All `.env*` files, credential files, and key files are blocked.

### Rules for contributors
1. **Never hardcode API keys** in source files — not even in tests.
2. **Never commit `.env` files** — copy `.env.example` to `.env.test` locally.
3. **Use environment variables** for all secrets in tests and CI.
4. **Rotate immediately** if a key is accidentally committed — assume it is compromised.

### Environment setup for tests
```bash
cp .env.example .env.test
# Edit .env.test and add your AWSYS_API_KEY
# This file is gitignored and will never be committed
```

### If a secret is accidentally committed
1. Rotate the key immediately (generate a new one in your AWSYS dashboard).
2. Contact security@awsys.co.
3. Use `git filter-repo` or BFG Repo Cleaner to purge the history.
4. Force-push the cleaned history (coordinate with the team).

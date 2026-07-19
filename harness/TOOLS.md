<!-- HARNESS:START
     version=0.33.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-19T05:36:09Z
     DO NOT EDIT — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Tools Reference

> Tools available to agent `awsysco-go-sdk`. Every call goes through the harness
> gateway (JWT auth → RBAC → execution → output scan → audit log).
> Tools not listed here will return a `forbidden` error — do not attempt to call them.

**Available to this agent:** 6 tools


---

## memory_store

Store a key-value pair in this agent's private namespace. Persists across sessions.
Namespace: `awsysco-go-sdk-worker` — other agents cannot read this unless explicitly granted.

```
Required
  key    string   identifier (e.g. "deploy-rule", "last-pr", "project-status")
  value  string   content to store (any text, including JSON)

Optional
  (none)
```

**Returns:** confirmation string — `stored key="<key>" in namespace="<namespace>"`

**Error cases:**
- Empty key or value → `"key and value are required"`
- Key length > 256 chars → `"key too long"`

**Example:**
```
memory_store(key="deploy-rule", value="staging first; production requires APPROVE gate")
```

---

## memory_search

Search this agent's namespace by keyword. Searches both keys and values.

```
Required
  query  string   keyword or phrase to search for

Optional
  namespace  string   cross-namespace read (only if grant configured in manifest)
```

**Returns:** JSON array of matching `{key, value}` pairs, or empty array if no match.

**Error cases:**
- Empty query → `"query is required"`
- Cross-namespace without grant → `"cross-namespace read requires explicit grant"`

**Example:**
```
memory_search(query="deploy")
→ [{"key":"deploy-rule","value":"staging first; production requires APPROVE gate"}]
```

---


## web_fetch

Fetch a URL and return its text content. HTML is stripped to readable text.
Private IP ranges are blocked (SSRF protection). Redirects are re-validated.

```
Required
  url    string   must start with https:// or http://

Optional
  (none)
```

**Returns:** page text content (stripped HTML), up to 32 KiB.

**Error cases:**
- Non-http/https scheme → `"only http and https URLs are allowed"`
- Private / loopback IP → `"private and internal URLs are not allowed"`
- Timeout (30 s default) → `"request timed out"`
- Non-200 status → `"HTTP <status>: <reason>"`

**Example:**
```
web_fetch(url="https://example.com/api/docs")
```

**Do not use for:**
- `file://`, `ftp://`, `smb://` — blocked by scheme allowlist
- `localhost`, `127.x.x.x`, `10.x.x.x`, `192.168.x.x` — blocked by SSRF guard
- `http://` URLs when content is sensitive — prefer `https://`



---


## web_search

Search the web using Brave Search. Returns titles, URLs, and descriptions.

```
Required
  query        string    search query

Optional
  max_results  integer   1–10, default 5
```

**Returns:** JSON array of `{title, url, description}` objects.

**Error cases:**
- Empty query → `"query is required"`
- `BRAVE_SEARCH_API_KEY` not set → tool unavailable (not in tools/list)
- API quota exceeded → `"brave search: quota exceeded"`

**Example:**
```
web_search(query="Firebase Cloud Run cold start optimization", max_results=5)
```



---



---


## file_ops

Read, write, or list files within this project's directory.
All paths are relative to the project root (`/Users/patrickbertsch/dev/awsysco-go-sdk`).

```
Required
  op    string   "read" | "write" | "list"
  path  string   relative path from project root (e.g. "src/index.ts")

Optional (write only)
  content  string  file content to write
```

**Returns:**
- `read` → file content as string
- `write` → `"written: <path> (<bytes> bytes)"`
- `list` → JSON array of relative file paths

**Error cases:**
- Path escapes project root → `"path traversal not allowed"`
- Blocked file type → `"access to .<ext> files is not allowed"` (.env, .pem, .key, .pfx, authorized_keys, AWS credentials)
- Symlink to outside root → `"symlink target is outside allowed root"`
- `write` without `content` → `"content is required for write"`
- File not found (read) → `"no such file: <path>"`

**Example — read:**
```
file_ops(op="read", path="main.go")
```

**Example — write:**
```
file_ops(op="write", path="notes/todo.md", content="# TODO\n- Fix auth bug\n")
```

**Example — list:**
```
file_ops(op="list", path="src")
```

**Never attempt:**
- `path="../../../etc/passwd"` — blocked, path traversal
- `path=".env"` — blocked, credential file
- `path=".env"` — blocked, credential file



---



---


## shell

**Not available to this agent.** `allow_shell` is `false` in the manifest.
Attempting to call `shell` will return: `"shell skill not permitted for agent awsysco-go-sdk"`



---




## code_search

Search this agent's own project codebase using lexical/BM25 full-text
search. Read-only — never modifies files, and only ever touches the one
project named in this agent's own `harness.project_root`.

```
Required
  op     string   "search" | "reindex"

For op="search"
  query  string   search terms
  limit  integer  max results (default 10, max 50)
```

**Returns:** matching `[path:start_line-end_line]` snippets, ranked by
relevance. `op="search"` indexes the project automatically the first time
it's called for this agent; `op="reindex"` forces a fresh index and should
be used sparingly (after a significant change), not before every search.

**Error cases:**
- No `harness.project_root` configured for this agent → rejected
- Empty `query` on `op="search"` → rejected

**Example:**
```
code_search(op="search", query="registerBuiltinTools")
```

**Do not use for:**
- Semantic/conceptual queries — matching is lexical (BM25 over raw tokens), not embedding-based; exact identifiers and terms work best
- Reading a known file in full — use `file_ops` instead
- Anything outside this agent's own project — there is no way to target another project's index


## Quick reference

| Tool | Required params | Returns |
|---|---|---|
| `memory_store` | `key`, `value` | confirmation string |
| `memory_search` | `query` | `[{key, value}]` array |
| `web_fetch` | `url` | page text |
| `web_search` | `query` | `[{title, url, description}]` array |

| `file_ops` | `op`, `path` | content / confirmation / path list |




| `code_search` | `op` | matching file:line snippets |

<!-- HARNESS:END -->

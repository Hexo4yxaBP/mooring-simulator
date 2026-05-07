# Access Control

*This is a client-side WASM application with no server, no user accounts, and no persistent data store. Traditional authentication/authorization does not apply. This document evaluates the relevant security boundary options for the deployment model.*

*Sources: constraints.md (C1 "must run in browser", U1 "RESOLVED: Ebiten → WASM, no server"), architecture.md (C4 Context: "No external APIs, no network calls after initial load").*

---

## Security Context

The application is a static file bundle (`index.html`, `wasm_exec.js`, `main.wasm`) served from any HTTP origin. All simulation state lives in browser memory. There is no backend, no database, no user identity, and no shared state between users.

The relevant security boundaries are:

1. **File serving** — who can access the static files
2. **Browser sandbox** — what the WASM binary can do inside the browser

---

## Option A — No Access Control (Open Static Hosting)

**Deployment flow:**  
Files served by any static HTTP server (Nginx, GitHub Pages, local `python -m http.server`). No authentication layer. Anyone with the URL can load and use the simulator.

**Pros:**
- Zero implementation complexity
- Matches the use case: single-user educational/training tool with no sensitive data
- No tokens, sessions, or login flows to build or maintain

**Cons:**
- No restriction on who can access the file if hosted publicly
- No audit trail of who ran simulations

**Audit trail:** None.

**Compromise risk:** None meaningful — there is no sensitive data, no server-side state, and no actions with side effects beyond the user's own browser tab.

**Implementation complexity:** Low (none).

---

## Option B — Static Hosting Behind HTTP Basic Auth

**Deployment flow:**  
Reverse proxy (Nginx, Caddy) sits in front of the static file server. Requests require a username/password via HTTP Basic Auth. WASM binary itself is unchanged.

**Pros:**
- Restricts access to the simulator to known users
- Trivial to configure on any reverse proxy
- No changes to Go code

**Cons:**
- Password sent base64-encoded each request (safe only over HTTPS — **HTTPS required**)
- No per-user audit trail (shared credential typical for small teams)
- Overkill for a local dev/training tool

**Audit trail:** HTTP access logs at proxy level (IP + timestamp).

**Compromise risk:** If credential is leaked, anyone can access the simulator — but there is still no server-side data to exfiltrate.

**Implementation complexity:** Low (Nginx config only, outside Go codebase).

---

## Recommendation

**Option A for MVP.**

The simulator contains no sensitive data, no user PII, no server-side state, and produces no side effects outside the user's browser tab (architecture.md C4 Context). Access control adds zero security value in this context and would add implementation overhead with no benefit.

If the project is later deployed as a shared training tool requiring access restriction, Option B (HTTP Basic Auth via reverse proxy) is the correct addition — it requires no changes to the Go codebase.

The browser's built-in WASM sandbox is the effective security boundary: the WASM binary cannot access the filesystem, make arbitrary network requests, or escape the browser tab. This is provided by the browser runtime, not by application code.

---

## WASM Security Notes (for completeness)

| Concern | Status |
|---------|--------|
| Filesystem access | Not possible from WASM (browser sandbox) |
| Network requests | Not initiated by this app (architecture.md: "no network calls after initial load") |
| Cross-origin data leakage | N/A — no external API calls |
| WASM binary tampering | Mitigated by serving over HTTPS with correct MIME type |
| `GOOS=js` binary | No CGO, no unsafe pointer use required for this app (constraints.md T1) |

**HTTPS note:** The WASM file should be served over HTTPS in any non-localhost deployment. `wasm_exec.js` and `main.wasm` must be served with correct MIME types (`application/wasm` for `.wasm`).

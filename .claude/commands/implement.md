# Phase 4 — Implementation
Perform Phase 4 contextual engineering. Goal — implement code according to the plan from Phase 3 and pass all Quality Gates.
---
## Step 1. Read the documents
Read:
- docs/Plan/implementation-tasks.md — task graph
- docs/Plan/test-strategy.md — testing strategy
- docs/Design/interface-spec.md — public interface
- docs/Design/data-schemas.md — normalization rules
- docs/Design/access-control.md — authentication decision
Build the task dependency graph. Start with tasks Depends on: none.
---
## Step 2. Implementation
For each task in dependency order:
1. Implement the code
2. Write unit test (if none — task not complete)
3. Make sure all Acceptance criteria are met
4. Mark [x] in implementation-tasks.md
Perform parallel tasks (without mutual dependencies) simultaneously.
### Code Conventions
- Input data validation — only at system boundaries (public inputs)
- External system errors → typed internal errors (do not rethrow raw)
- Credentials (tokens, passwords, keys) — never in logs, errors, responses
- Functions are small, with single responsibility
- Tests are written together with code (not after)
- Conventional commits: feat / fix / test / docs / refactor / chore
---
## Step 3. Quality Gates
All gates must pass before declaring Implementation complete.
Perform and check each item:

| Gate                  | Action                          | Threshold                  |
|-----------------------|---------------------------------|----------------------------|
| Tests pass            | run test runner                 | 100% pass                  |
| Coverage              | run coverage tool               | ≥ 80% lines                |
| Races / concurrency   | run race detector               | 0 races                    |
| Vulnerabilities       | run vuln scanner                | 0 HIGH/CRITICAL            |
| Credential leakage    | grep -rn "token\|password\|Bearer\|api_key\|secret" src/ | 0 matches in production code |
| Static analysis       | run linter + vet                | 0 errors                   |
| Payload size          | manual output check             | within spec from design    |

Use tools specified in AGENTS.md Commands section.
---
## Step 4. Security Review
After completing all tasks, create docs/Implementation/security-review.md:
# Security Review
## PASSED
- [What was checked and passed]
## WARNINGS
- [Issue] — [file:line] — [recommendation]
## CRITICAL
- [Issue] — [file:line] — [attack vector] — [fix with code example]
## SECURITY SCORE: X/10
Minimum checklist for review:
- [ ] Grep for credential leakage: token, password, Bearer, api_key, secret, -----BEGIN
- [ ] Input parameter validation on all public inputs
- [ ] External system errors do not reveal internal paths, usernames, system details
- [ ] Dependencies without critical CVE (vuln scanner)
- [ ] Logs do not contain PII and credentials
---
## Stop Conditions
Stop and notify the user if:
- API does not match Research — external system response does not match documented schema → need to update docs/Research/ and docs/Design/
- Performance outside spec (response time, payload size) → requires investigation
- HIGH/CRITICAL CVE in dependencies → fix or explicitly exclude with justification
- Architectural decision is not implementable → return to Phase 2 (/design) and update documents
---
Completion: **"Phase 4 (Implementation) completed. Security review in docs/Implementation/security-review.md."**
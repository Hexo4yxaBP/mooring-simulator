# Phase 3 — Plan
Perform Phase 3 contextual engineering. Goal — break down the implementation into atomic tasks with dependencies and completion criteria.
---
## Step 1. Read all documents from docs/Design/
The decomposition must exactly match the accepted architectural decisions. Do not introduce new components — implement what has been designed.
---
## Step 2. Create docs/Plan/implementation-tasks.md
### File Structure
# Implementation Tasks
## [Component Group 1]
[tasks]
## [Component Group 2]
[tasks]
Group by architecture layers from docs/Design/architecture.md. Typical groups:
- Project skeleton (initialization, configuration, logging)
- Transport layer / external system client
- Business logic layer
- Data normalization layer
- Implementation of each method/tool of the interface
- Tests
### Format of each task
### TASK-NNN: [verb + object]
**Depends on:** TASK-MMM, TASK-KKK (or "none")
**Acceptance criteria:**
- [ ] Specific, verifiable criterion (not "works correctly")
- [ ] Another criterion
**Definition of done:** all criteria checked + unit test exists
### Requirements
**Size:** 1–4 hours per task
**Dependencies:** explicit TASK-ID
**Criteria:** behavioral and verifiable:
  - Good: "Function X with argument Y returns Z"
  - Bad: "Works correctly", "Implemented properly"
- **[BLOCKED: reason]** — for tasks with unresolved issues
**Security tasks** are included in the same plan — not separated
### Mandatory security task types
Include explicit tasks for:
- Validation of all input parameters (types, ranges, formats)
- Prevention of credential leakage in logs, errors, responses
- Mapping of external system errors → typed internal errors
- Security review (grep-pattern check + review)
---
## Step 3. Create docs/Plan/test-strategy.md
- **Unit tests** (pure functions, transformations, validation)
- **Integration tests** (HTTP client, DB, external — mocked where appropriate, real environment used where necessary)
- How the test environment works in CI (secrets, sandbox, environment variables)
- Minimum coverage threshold (recommended ≥ 80%)
- Tools (test runner, coverage, race detector, vuln scanner)
---
## Step 4. Create docs/Plan/review-gaps.md
Check implementation-tasks.md for completeness:
# Review: Implementation Gaps
## Missing components
[Are there any components from Architecture not covered by tasks?]
## Tasks without tests
[Are there tasks without explicit test mention?]
## Cyclic dependencies
[Check: no A→B→A chains]
## Security requirements without tasks
[Are all requirements from access-control.md covered by tasks?]
## Blocked tasks
[List of [BLOCKED] tasks with path to unblocking]
---
## Step 5. Update AGENTS.md
Add or update section with:
- Technology stack (language, version, key libraries)
- Build, test, linting, vulnerability check commands
- Code conventions (validation boundary, error typing, commit conventions)
---
## Step 6. Stop Questions
After creating documents answer:
1. Are all [BLOCKED] tasks resolved or excluded with explicit justification?
2. Is the task dependency graph acyclic?
3. Is the test environment available (sandbox, mock server, test data)?
4. Is Security review included as an explicit task (not implicit intention)?
---
Completion: **"Phase 3 (Plan) completed. Run /implement to start implementation."**
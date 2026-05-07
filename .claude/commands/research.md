# Phase 1 — Research
Task: $ARGUMENTS
Perform Phase 1 contextual engineering for this task. Goal — collect facts. No architectural decisions at this stage.
---
## Step 1. Task Classification and Input Data
Define:
1. Task type:
   - Integration with external API/service
   - Adding functionality to existing codebase
   - Refactoring/rewriting existing code
   - Other (describe)
2. Input data: find in the project everything that contains facts about the system:
   - API specifications (OpenAPI/Swagger YAML/JSON, Protobuf, GraphQL schema)
   - Documentation, README, technical descriptions
   - Existing code (if the task is extension)
   - Database schemas, configs
   - Any other sources of truth
3. Phase 1 Questions: formulate a list of specific questions whose answers are needed for design.
Create `docs/Research/task-brief.md` with this information.
---
## Step 2. Creation of Research Documents
Create the `docs/Research/` directory. The set of documents depends on the task type:
### For integration with external API
**docs/Research/api-map.md** — map of relevant endpoints:
| Tag | Method | Path | Key params | Key response fields | Priority |
|-----|--------|------|-----------|-------------------|----------|
Priority: must / nice / skip. Ignore endpoints outside the task scope.
**docs/Research/auth-flow.md** — authentication:
- Supported methods (token, Basic, OAuth, API Key, cookie)
- Exact names of headers, parameters, request body fields from the specification
- Token format, TTL, refresh mechanism
- How the token is passed in subsequent requests
**docs/Research/data-model.md** — data structures:
- Schema of main objects with fields: name, type, nullable, description
- Differences between schemas of different endpoints (for example, search vs direct retrieval)
- This will become the target for normalization in Phase 2
**docs/Research/constraints.md** — constraints and features:
- Rate limits, pagination limits, maximum sizes
- Deprecations and API quirks
- Requirements for special permissions
- [UNKNOWN — verify on instance] for everything unknown
### For adding functionality to existing codebase
**docs/Research/code-map.md** — code map:
- Modules affected by the task, with key functions/methods
- Existing extension points (interfaces, hooks, plugins)
- Dependency graph in the area of changes
**docs/Research/interfaces.md** — existing contracts:
- Public interfaces and types that cannot be broken
- Tests tied to current behavior (what cannot be changed without updating tests)
- Versioning and backward compatibility
**docs/Research/constraints.md** — technical constraints:
- What cannot be broken (backward compatibility, external consumers)
- Legacy code that interferes (tech debt blocking the task)
- Platform/environment dependencies
### For refactoring
**docs/Research/current-state.md** — current state:
- Structure: modules, classes, functions in the refactoring area
- Patterns currently used
- Specific problems: duplication, etc.
**docs/Research/dependencies.md** — dependencies:
- External consumers of the code (who calls, what imports)
- Public contracts that must remain unchanged
**docs/Research/constraints.md** — what can/cannot be changed:
- Breaking changes: allowed or not
- What is obsolete and can be removed
- Version constraints
---
## Rules for Research Documents
- **Only facts.** No decisions — only what is in the sources
- **Exact names.** Fields, parameters, paths
- **Marking unknown.** What could not be established → [UNKNOWN — verify on instance]
- **Source of each fact.** File + section so the reader can verify
---
## Step 3. Stop Questions
[07.05.2026 19:50] Dmitriy P: After creating the documents answer:
1. Are there any remaining [UNKNOWN] markers that will block design in Phase 2?
2. Was all input data available, or did something have to be assumed?
3. Are there any architectural forks visible already now?
---
Completion: **"Phase 1 (Research) completed. Run /design to proceed to architecture."**
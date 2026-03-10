# Specification Guideline

## 1. Scope

This document defines where different types of specifications are written in this project.

## 2. Specification Placement Rules

| Spec Type | Location | Content |
|---|---|---|
| Feature motivation, user stories, requirements, acceptance criteria | `doc/spec/` | WHY and WHAT |
| Use case scenario (preconditions, input, output, error cases) | `doc/spec/` use case file | WHAT (high-level) |
| Domain entity invariants, business rules, state transitions | Doc comment on the type | Rules enforced by this type |
| Use case processing flow, error handling policy | Doc comment on the method | Implementation contract |
| Interface contracts (pre/postconditions, error/nil semantics) | Doc comment on the interface | Caller/implementor contract |

## 3. Boundary Between doc/spec/ and Code

- `doc/spec/` owns: WHY (motivation, user stories) and WHAT (requirements, acceptance criteria)
- Code doc comments own: rules, contracts, and behavior at the implementation level
- After implementation, `doc/spec/` use case Expected Behavior should reference code paths rather than duplicate processing steps
- Rationale: AI agents always read code they modify; doc comments are discovered automatically

## 4. AI Agent Discovery

- AI agents read source code before modifying it; doc comments in code are always seen
- AI agents do not read all files in `doc/spec/`; CLAUDE.md links and spec "Related Code" sections are the discovery path
- Therefore, implementation-level specs that must not be missed belong in code, not `doc/spec/`

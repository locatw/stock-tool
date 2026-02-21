# Specifications

Specifications live in `doc/spec/<spec>/`. Each spec directory contains:

- `<spec>.md` — Spec overview: background, purpose, user stories, acceptance criteria, requirements, design direction
- `usecase/<usecase>.md` — Individual use case scenarios for the spec

## Spec Conventions

- Background, purpose, requirements, design direction level of granularity
- Include acceptance criteria that define "done"
- Template: `_template/_template.md`

## Use Case Conventions

- One file per distinct operation scenario
- Designed for Claude to reference during implementation
- Self-contained: reader should understand the scenario without reading other use case files
- Include preconditions, input, expected behavior, output, error cases, acceptance criteria
- Reference related code layers by directory
- Template: `_template/usecase/_template.md`

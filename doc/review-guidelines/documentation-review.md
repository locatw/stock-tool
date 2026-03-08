# Documentation Review Guidelines

This file defines what to verify when reviewing changes for documentation consistency and code-doc drift.

## Scope

Apply these checks whenever a pull request modifies `.go`, `.sql`, `.yaml`, or `.md` files.

## Checks

### Contradictions Between Docs

- Flag any two documentation files that make conflicting claims about the same system behavior, configuration value, or architectural decision.
- When a newer doc supersedes an older one, verify the older doc is updated or explicitly deprecated.

### Stale Content and Code-Doc Drift

- Verify that file paths, package names, function names, and type names referenced in docs still exist in the codebase.
- When a domain entity, repository method, or use case is added or renamed, check that related docs (architecture, project overview, proposals) reflect the change.
- When a public API surface is removed, confirm that docs no longer describe it as available.

### Missing or Renamed Files Listed in CLAUDE.md or AGENTS.md

- Check that every file path listed in the Documentation Index of `AGENTS.md` and in `CLAUDE.md` exists at the stated path.
- If a doc file is renamed or moved, flag any index entries that still point to the old path.

### New Public API Without Documentation

- When a new domain entity, repository interface, or use case is introduced, verify that at least one of the following is updated: `doc/architecture.md`, `doc/project-overview.md`, or a relevant spec file under `doc/spec/`.
- A missing update is a documentation drift defect, not a style issue.

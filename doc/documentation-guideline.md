# Documentation Guideline

## Language

All documentation must be written in English.

## Format

- Use Markdown (`.md`) format
- Name files according to their purpose (e.g., `documentation-guideline.md` for documentation rules)

## Structure

Each guideline or rule document should contain:

- Clear title and purpose
- Detailed rules with examples
- Severity levels (Error/Warning) where applicable
- Rationale for each rule
- Good examples and bad examples with explanations

## Principle: Avoid Easily Outdated Content

Documentation should describe structure and conventions, not enumerate current state.
Content that changes with routine development (adding a package, endpoint, migration, etc.) must not be written inline in documentation.

### Do Not Include

- Counts: "5 tables", "3 API endpoints", "7 packages"
- Exhaustive lists of contents: Listing every package, migration file, or domain entity
- Specific version numbers: Pin versions only in actual config files (`go.mod`, `compose.yaml`, etc.), not in prose

### Do Include

- Directory paths: Where to find things (`backend/migrations/`, `backend/internal/domain/`)
- Structural conventions: How packages and layers are organized and why
- Commands and workflows: How to run, build, test, migrate
- Configuration file references: Which config files control what behavior (e.g., "`compose.yaml` for local database setup")

### Rationale

Documentation that enumerates current contents creates a maintenance burden — every routine change (adding a migration, creating a new repository, etc.) requires a documentation update.
Readers can discover current contents by looking at the filesystem.
Documentation should focus on what cannot be inferred from the code itself.

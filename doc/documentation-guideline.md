# Documentation Guideline

## Language

All documentation must be written in English.

## Format

- Use Markdown (`.md`) format
- Name files according to their purpose (e.g., `markdown.md` for Markdown rules)
- Follow the Markdown style guide in `doc/coding-guidelines/markdown.md`

## Structure

Each document should contain:

- Clear title and purpose
- Specific enough detail for readers to act on

For guideline and rule documents, also include:

- Code examples (good and bad) with explanations
- Rationale for each rule

## Principle: Top-Down Ordering

Organize sections within a document from broad to narrow. Start with high-level concepts, design policies, or overviews, then progress into implementation details and configuration references.

Readers go through a document from top to bottom. Presenting the big picture first gives them the context needed to understand the details that follow.

## Principle: Avoid Easily Outdated Content

Documentation should describe structure and conventions, not enumerate current state.
Content that changes with routine development (adding a command, creating a new endpoint, etc.) must not be written inline in documentation.

### Do Not Include

- Counts: "3 entry points", "4 .env files", "5 external services"
- Exhaustive lists of contents: Listing every command, package, or migration file
- Specific version numbers: Pin versions only in actual config files (`go.mod`, `compose.yaml`, etc.), not in prose

### Do Include

- Directory paths: Where to find things (`backend/cmd/`, `backend/internal/`)
- Structural conventions: How packages and layers are organized and why
- Commands and workflows: How to run, build, test
- Configuration file references: Which config files control what behavior (e.g., "`compose.yaml` for local infrastructure configuration")

### Rationale

Documentation that enumerates current contents creates a maintenance burden — every routine change (adding a command, creating a new endpoint, etc.) requires a documentation update.
Readers can discover current contents by looking at the filesystem.
Documentation should focus on what cannot be inferred from the code itself.

## Principle: DRY (Don't Repeat Yourself)

Avoid duplicating operational procedures and frequently changing values across multiple documents.
When a procedure is documented in another file (e.g., environment setup in `doc/requirements.md`), link to it instead of restating the steps.
When a configuration file is the source of truth for specific values (port numbers, database credentials, S3 bucket paths, etc.), point readers to the config file path instead of copying the values into prose.

### Acceptable Duplication

Duplication is acceptable when it serves readability:

- Design rationale and background — explaining why a setting exists (e.g., "the stock schema is separated from public to avoid collision with framework-managed tables") is valuable even if the same fact appears in another document
  - Readers should not have to navigate to a different file to understand the reasoning behind a design decision in the current document
- Stable technical facts — well-established information such as protocol standards, encoding formats, and widely-known conventions is unlikely to change and safe to repeat where it aids comprehension
- Brief context summaries — when referencing another document, a short summary of the topic (one to two sentences) before the link helps readers decide whether to follow it

### What to Avoid

- Duplicated step-by-step procedures — these diverge quickly when one copy is updated
- Copied lists of current values (port numbers, credentials, bucket paths) that change with routine operations — reference the config file instead
- Restating the same design explanation at the same level of detail in multiple documents — pick one document as the primary source and summarize with a link in others

### Rationale

Duplicated operational content inevitably diverges when one copy is updated but the other is not.
However, forcing readers to follow links for every piece of context harms readability.
The goal is to balance consistency (avoid drift) with comprehension (self-contained explanations).

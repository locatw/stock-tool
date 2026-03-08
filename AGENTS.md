# AGENTS.md

## Security Constraints

- NEVER read, cat, or access any `.env` file in this repository.
- NEVER read files matching: `.env`, `.env.*`, `*.pem`, `*.key`.
- Focus only on source code files: `.go`, `.md`, `.yaml`, `.sql`.

## Documentation Index

- [doc/project-overview.md](doc/project-overview.md) — What this project does and its entry points
- [doc/architecture.md](doc/architecture.md) — Clean architecture layers, domain model, database design
- [doc/coding-guidelines/go/coding.md](doc/coding-guidelines/go/coding.md) — Go coding conventions
- [doc/coding-guidelines/go/testing.md](doc/coding-guidelines/go/testing.md) — Go testing conventions
- [doc/review-guidelines/overview.md](doc/review-guidelines/overview.md) — Review guidelines index
- [doc/review-guidelines/documentation-review.md](doc/review-guidelines/documentation-review.md) — Doc consistency and code-doc drift
- [doc/review-guidelines/markdown-review.md](doc/review-guidelines/markdown-review.md) — Markdown style conformance

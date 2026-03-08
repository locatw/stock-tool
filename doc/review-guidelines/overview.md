# Review Guidelines Overview

This directory defines the review standards for this project. The guidelines cover what to verify when reviewing code changes, not how to write code.

For implementation conventions, see [doc/coding-guidelines/](../coding-guidelines/).

## Scope

- [code-review.md](code-review.md) — Bug detection and logic error checks
- [security-review.md](security-review.md) — Security vulnerability checks (OWASP Top 10)
- [test-coverage.md](test-coverage.md) — Test coverage standards
- [architecture-review.md](architecture-review.md) — Clean architecture alignment
- [documentation-review.md](documentation-review.md) — Doc consistency and code-doc drift
- [markdown-review.md](markdown-review.md) — Markdown style conformance

## Review Philosophy

- Focus on correctness first, then security, then architecture conformance.
- Flag violations with a reference to the relevant guideline so the author can look up the rationale.
- Do not suggest improvements outside the scope of the change unless they are directly related to a defect.

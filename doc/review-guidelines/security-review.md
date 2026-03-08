# Security Review Guidelines

Check for the OWASP Top 10 categories most relevant to this Go backend.

## Input Validation (A03: Injection)

- Confirm that all SQL queries use parameterized statements via GORM; no raw string interpolation.
- Verify that path parameters and query strings are validated before use.
- Check that user-supplied values are never passed directly to shell commands.

## Authentication and Authorization (A01: Broken Access Control, A07: Auth Failures)

- Confirm that protected endpoints require authentication.
- Check that resource ownership is verified before read or write operations.
- Verify that tokens and credentials are not logged.

## Sensitive Data Exposure (A02: Cryptographic Failures)

- Confirm that secrets are loaded from environment variables, not hardcoded.
- Check that `.env` files and private keys are not read or logged by code.
- Verify that passwords or API keys are not included in error messages or responses.

## Security Misconfiguration (A05)

- Check that CORS settings do not allow arbitrary origins in production paths.
- Verify that debug endpoints or verbose error details are not exposed in production code paths.

## Vulnerable Dependencies (A06)

- Note any new third-party packages added; flag if they are unknown or unmaintained.

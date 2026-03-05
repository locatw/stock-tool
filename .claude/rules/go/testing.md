---
paths:
  - "**/*_test.go"
---

# Go Testing Rules

IMPORTANT: Before writing or modifying Go test code, you MUST read `doc/coding-guidelines/go/testing.md` and follow all rules defined there.

Key rules:
- Use `testify/suite` for test organization
- Table-driven subtests for parameterized testing
- Mock structs use `Mock` suffix (e.g., `UserRepositoryMock`)
- Use `go-cmp` (`cmp.Diff`) for struct comparison, not individual field assertions
- `s.True(cmp.Equal(expected, actual), cmp.Diff(expected, actual))`
- Usecase tests: real DB/S3 via dockertest; only mock external APIs
- Skip tests for trivially simple functions exercised by other tests
- Handler tests: `mock.Anything` only for `context.Context`; all other arguments must be specific values

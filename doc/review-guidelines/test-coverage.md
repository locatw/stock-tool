# Test Coverage Guidelines

Check that tests meet the standards defined in [doc/coding-guidelines/go/testing.md](../coding-guidelines/go/testing.md).

## Test Suite Structure

- Struct unit tests must use `testify/suite`.
- Multiple cases for the same function must use table-driven subtests inside a single test method.
- Separate test methods are acceptable only when setup or assertions differ significantly between cases.

## Mock Naming

- Mock structs must use the `Mock` suffix (e.g., `UserRepositoryMock`).

## Usecase Test Strategy

- Usecase tests must use real infrastructure (PostgreSQL, SeaweedFS) via `dockertest`.
- Only external third-party services (e.g., J-Quants API) may be mocked.
- Confirm that pure mock-only usecase tests are flagged for replacement.

## Struct Comparison

- Struct comparisons must use `go-cmp` (`cmp.Diff`) in a single assertion.
- All fields in the expected struct must be set; omitting a field means accepting any value.

## Distractor Data

- Any test that filters, scopes, or targets by an identifier must include at least one distractor record.
- The distractor must share the same structure but belong to a different scope.
- Verify that assertions would fail if the WHERE clause or filter were removed.

## Trivial Functions

Functions that do not require dedicated tests:

- `NewXXX()` constructors that only assign arguments to fields and set timestamps.
- Getter methods that return a single private field.

Functions that do require tests:

- Helpers with conversion logic or non-trivial defaults.
- Constructors with conditional branching.

## Handler Tests

- Handler tests must use `testify/mock` to mock the usecase layer.
- All non-context arguments must use specific expected values, not `mock.Anything`.

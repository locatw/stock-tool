# Go Testing Guidelines

## 1. Testing

### Test Suite

Use `testify/suite` for struct unit tests:

```go
type FooTestSuite struct {
    suite.Suite
}

func TestFoo(t *testing.T) {
    suite.Run(t, new(FooTestSuite))
}

func (s *FooTestSuite) TestBar() {
    s.Equal(expected, actual)
}
```

Each test suite must target a single type.
Name the suite after the type (e.g., `FooTestSuite` for type `Foo`).
Name test methods after the function or method under test, omitting the type name prefix
(e.g., `TestNew` inside `FooTestSuite`, not `TestNewFoo`).

### Parameterized Tests

When testing a single function or method with multiple cases, use a single test method with table-driven subtests:

```go
func (s *FooTestSuite) TestBar() {
    type TestCase struct {
        name     string
        input    int
        expected int
    }
    tests := []TestCase{
        {"positive", 1, 2},
        {"zero", 0, 1},
        {"negative", -1, 0},
    }
    for _, tt := range tests {
        s.Run(tt.name, func() {
            s.Equal(tt.expected, Bar(tt.input))
        })
    }
}
```

Separate test methods are acceptable when the setup or assertions differ significantly between cases (e.g., different mock configurations in usecase tests).

### Mock Naming

Mock structs should use the `Mock` suffix (e.g., `UserRepositoryMock`).

### Usecase Tests

Usecase tests should be integration tests that use real infrastructure (DB, S3) via `dockertest`, and only mock external third-party services (e.g., J-Quants API):

| Dependency | Approach | Reason |
|---|---|---|
| Database | Real (PostgreSQL via dockertest) | Catches query bugs, schema mismatches |
| Object storage | Real (SeaweedFS via dockertest) | Catches upload/key issues |
| External APIs | Mock | Cannot call in tests, rate limits, costs |

Rationale: catches integration bugs that pure mock tests miss (wrong column names, type conversion errors, S3 key format issues) and makes tests resilient to internal refactoring.

### When to Skip Tests

Skip writing tests for functions or methods where both of the following apply:

- The implementation is trivially simple (direct field assignment, no branching logic)
- The function is exercised frequently by other tests (e.g., constructors used in integration tests, getters called in assertions)

Examples of functions that do not need dedicated tests:

- `NewXXX()` constructors that only assign arguments to fields and set timestamps
- Getter methods that return a single private field

Examples of functions that do need tests:

- Helper methods with conversion logic (e.g., `StaleTimeout()` converting minutes to `time.Duration`)
- Constructors with conditional logic or non-trivial defaults

### Struct Comparison

Use `go-cmp` (`cmp.Diff`) to compare structs in a single assertion instead of asserting individual fields. This detects missing field mappings in conversion functions and ensures no field is silently ignored.

- Set all fields in the expected struct — omitting a field means accepting any value for it
- Use `s.True(cmp.Equal(expected, actual), cmp.Diff(expected, actual))` within testify/suite tests — the diff string becomes the failure message

```go
// Good — all fields verified in one assertion; diff shown on failure
expected := api.GetDataSource200JSONResponse{
    Id: id1, Name: "src", Enabled: true, Timezone: "UTC",
    Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
}
actual := resp.(api.GetDataSource200JSONResponse)
s.True(cmp.Equal(expected, actual), cmp.Diff(expected, actual))

// Bad — individual field checks miss omitted fields
s.Equal("src", resp.Name)
s.Equal(true, resp.Enabled)
```

When `cmp.Option` values are needed (e.g. `cmpopts.IgnoreFields`), pass them to both calls:

```go
opts := []cmp.Option{cmpopts.IgnoreFields(Foo{}, "UpdatedAt")}
s.True(cmp.Equal(expected, actual, opts...), cmp.Diff(expected, actual, opts...))
```

### Distractor Data

Tests must include data that a buggy implementation would incorrectly return — distractor data. A test that passes with both a correct and a broken implementation proves nothing.

When testing any operation that filters, scopes, or targets by an identifier, create at least one additional record that shares the same structure but belongs to a different scope. Then assert that only the expected records are returned / affected.

```go
// Good — distractor source proves the WHERE clause works
src1 := createSource("src1")
src2 := createSource("src2")
createDataType(src1.ID, "dt1")
createDataType(src2.ID, "dt-other") // distractor

result, _ := repo.ListBySourceID(ctx, src1.ID)
s.Len(result, 1) // fails if WHERE is missing

// Bad — no distractor; passes even if ListBySourceID ignores sourceID
src := createSource("src")
createDataType(src.ID, "dt1")

result, _ := repo.ListBySourceID(ctx, src.ID)
s.Len(result, 1) // passes with SELECT * FROM data_types (no WHERE)
```

When to add distractors:

- Filter / list queries — another parent with its own children
- Find by unique key — another record with a different key, to confirm the lookup does not just return the first row
- Delete / update by ID — another record that must survive the operation unchanged; verify it still exists afterward

## 2. Handler Tests

Handler tests use `testify/mock` to mock the usecase layer. The primary goal is to verify that the handler correctly converts API request objects into usecase request objects and maps usecase responses back to API response objects.

- `context.Context` is passed through unchanged, so use `mock.Anything` for it
- All other arguments must use specific expected values, not `mock.Anything`, to verify conversion logic

```go
// Good — specific expected request verifies handler→usecase conversion
expectedReq := &usecase.CreateDataSourceRequest{
    Name: "src", Enabled: true, Timezone: "UTC", Settings: map[string]any{},
}
s.ucMock.On("Create", mock.Anything, expectedReq).Return(...)

// Bad — mock.Anything hides mapping bugs (e.g. ID not set, wrong field)
s.ucMock.On("Create", mock.Anything, mock.Anything).Return(...)
```

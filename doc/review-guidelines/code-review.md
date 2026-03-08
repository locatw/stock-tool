# Code Review Guidelines

Focus on correctness and adherence to project coding conventions.

## Bug and Logic Errors

- Verify that error return values are always checked and propagated.
- Check that context is passed through correctly and not discarded.
- Confirm that nil pointer dereferences are not possible on expected code paths.
- Look for off-by-one errors in slice indexing and range loops.
- Verify that timezone handling only occurs at infrastructure boundaries, not in upper layers.

## Go Coding Convention Violations

Check for violations of [doc/coding-guidelines/go/coding.md](../coding-guidelines/go/coding.md):

- Comments that restate what the function name already says (remove them).
- Manual slice loops where `samber/lo` should be used (except DB operations).
- Interfaces defined in the repository package (must be defined by the consumer).
- Lines exceeding 120 columns.
- Multi-line struct literals or function calls with multiple elements on the same line.
- Unnecessary intermediate variables before a return statement.
- `=` instead of `:=` when declaring an error in a single-line if statement.
- Redundant GORM tags that match default behavior.
- Repository methods missing `context.Context` as the first parameter.
- Repository methods that call `WithContext()` is missing for GORM operations.
- Methods or functions added that have no current callers (YAGNI violation).

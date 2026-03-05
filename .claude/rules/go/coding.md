---
paths:
  - "**/*.go"
---

# Go Coding Rules

IMPORTANT: Before writing or modifying Go code, you MUST read `doc/coding-guidelines/go/coding.md` and follow all rules defined there.

The most critical rules (always follow even without reading the full document):

- Use `samber/lo` for slice transformations (map, filter); keep standard for loops only for DB operations
- Return function results directly — do not store in a variable just to return it
- Use `:=` (not `=`) in single-line `if err` checks to scope the variable
- Line length max 120 columns; break multi-param signatures across lines
- When a struct literal or function call spans multiple lines, each element on its own line
- Delete comments that merely restate the type name or constructor purpose
- Do not define interfaces in the repository package — let consumers define them

GORM models:
- Remove tags that match GORM defaults (primaryKey, column name, not null)
- Singular struct names (ExtractedDataFile, not ExtractedDataFiles)
- Pointer slices for one-to-many: `Children []*Child`
- No parent reference in child structs
- `ToEntity()` method on DB models; private `toDBModel()` for reverse

Repository:
- No transaction handling in repository methods
- Use "Repository" not "Repo"
- All methods accept `context.Context` as first parameter
- Always use `WithContext()` for GORM operations

Timezone:
- Convert only at the boundary where needed, not in upper layers

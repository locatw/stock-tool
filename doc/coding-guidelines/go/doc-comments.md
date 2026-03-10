# Go Doc Comment Conventions

## 1. Scope

This document defines what to write in Go doc comments. For the overall strategy of where specifications belong (doc/spec/ vs code), see [specification-guideline.md](../../specification-guideline.md).

## 2. Relationship with Existing Comment Rules

- [coding.md](coding.md) section 2 forbids trivial restating comments
- This document defines what constitutes a valuable doc comment
- Both rules apply: delete trivial comments, write meaningful ones

## 3. Domain Entity Doc Comments

What to write: business purpose, invariants, state transition rules.

What NOT to write: field-by-field descriptions, table/column mapping.

Example using `DataSource`:

```go
// DataSource represents an external data provider from which stock data
// is ingested. Timezone must be a valid IANA location; NewDataSource
// validates on creation and Update re-validates on mutation.
type DataSource struct {
```

Example using `ExtractTaskExecution`:

```go
// ExtractTaskExecution tracks a single run of data extraction.
// Status transitions: running -> succeeded (via Succeed) or
// running -> failed (via Fail). Terminal status must not change.
type ExtractTaskExecution struct {
```

## 4. Use Case Doc Comments

What to write: one-sentence summary, processing flow (numbered list), error handling policy, return value semantics (especially nil-return conventions).

The one-sentence summary states the business purpose (what the caller gets), not implementation steps. Do not include verbs like "validates", "persists", or "queries" in the summary.

What NOT to write: constructor restatement, full spec duplication from `doc/spec/`, specific validation field names or error trigger conditions (e.g., "if the timezone is invalid"). Listing specific conditions creates maintenance burden when validation rules change. Instead, write only the error type name and semantics (e.g., ValidationError on invalid input, (nil, nil) when not found).

Example using `Extract`:

```go
// Extract fetches raw data from a source API and stores it in S3.
//
// Processing flow:
//  1. Find or create ExtractTask for (source, dataType, timing)
//  2. Create a running ExtractTaskExecution
//  3. Fetch raw data from the source API
//  4. Upload raw data to S3
//  5. Record S3 key in ExtractedDataS3
//  6. Mark execution as succeeded
//
// On failure at steps 3-5, the execution is marked as failed before
// returning the error.
//
// See doc/spec/data-ingestion/usecase/ingest-data.md for requirements.
func (uc *ExtractTaskUseCase) Extract(ctx context.Context, req *ExtractTaskRequest) (*ExtractTaskResponse, error) {
```

Example using `Create`:

```go
// Create creates a new data source. Returns a ValidationError on invalid input.
func (uc *DataSourceUseCase) Create(ctx context.Context, req *CreateDataSourceRequest) (*DataSourceResponse, error) {
```

Example using `Get`:

```go
// Get retrieves a data source by ID. Returns (nil, nil) when not found.
func (uc *DataSourceUseCase) Get(ctx context.Context, id uuid.UUID) (*DataSourceResponse, error) {
```

## 5. Interface Doc Comments

What to write: purpose of the interface, method-level contracts (preconditions, postconditions, nil/error semantics).

What NOT to write: restating the method signature in prose.

Example using `DataSourceRepository`:

```go
// DataSourceRepository provides persistence for DataSource entities.
type DataSourceRepository interface {
    // Create persists a new DataSource. Returns the created entity with
    // server-assigned fields.
    Create(ctx context.Context, src *ingestion.DataSource) (*ingestion.DataSource, error)

    // FindByID returns the DataSource with the given ID, or (nil, nil) if not found.
    FindByID(ctx context.Context, id uuid.UUID) (*ingestion.DataSource, error)
}
```

## 6. Other Exported Symbols

For exported symbols not covered by sections 3-5 (standalone functions, constants, request/response structs, error variables, etc.), write a doc comment only when the information cannot be read from the name or the source code.

Example -- the output format is a convention not derivable from the implementation's callers:

```go
// GenerateS3Key generates an S3 object key following the landing layer path convention:
// landing/{source}/{data_type}/{yyyy}/{mm}/{dd}/{timestamp}_{uuid}.{ext}
func GenerateS3Key(source string, dataType string, executionTime time.Time, ext string) string {
```

Skip doc comments when the name is self-explanatory (e.g., `NewExtractTaskUseCase`, `CreateDataSourceRequest`, `ErrNotAuthorized`).

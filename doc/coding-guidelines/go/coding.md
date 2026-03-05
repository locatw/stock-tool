# Go Coding Guidelines

## 1. Code Structure

- Keep related type definitions together
- Place `NewXXX()` functions immediately after their struct definitions

## 2. Comments

Delete comments that don't provide additional context:

- Comments explaining table mapping (clear from type name)
  ```go
  // ExtractTask represents a record in the stock.extract_tasks table  // unnecessary
  type ExtractTask struct {
  ```

- Comments explaining constructor functions (clear from naming)
  ```go
  // NewExtractTask creates a new ExtractTask instance  // unnecessary
  func NewExtractTask(source, dataType, status string) *ExtractTask {
  ```

## 3. Slice Operations

Use `samber/lo` for slice transformations:

```go
// Good
files := lo.Map(t.Files, func(f *File, _ int) *domain.File {
    return f.ToEntity()
})

// Bad
files := make([]*domain.File, len(t.Files))
for i, f := range t.Files {
    files[i] = f.ToEntity()
}
```

Keep standard for loops for database operations:

```go
// Good - Keep explicit control for DB operations
for _, file := range files {
    if err := db.Create(file).Error; err != nil {
        return err
    }
}
```

## 4. Interface Definition

- Do not define interfaces in the repository package
- Let the consumer define interfaces according to their needs
- Repository package should only provide concrete implementations

## 5. Code Formatting

- Line length should not exceed 120 columns
- Break method signatures with multiple parameters into multiple lines:
  ```go
  // Good
  func (r *ExtractTaskRepository) Create(
      ctx context.Context,
      task *ExtractTask,
      files []*ExtractedDataFile,
      s3Files []*ExtractedDataS3,
  ) error {

  // Bad
  func (r *ExtractTaskRepository) Create(ctx context.Context, task *ExtractTask, files []*ExtractedDataFile, s3Files []*ExtractedDataS3) error {
  ```
- When a struct literal or function call spans multiple lines, each element must occupy its own line:
  ```go
  // Good
  foo := Foo{
      FieldA: 1,
      FieldB: 2,
      FieldC: 3,
  }

  bar(
      arg1,
      arg2,
      arg3,
  )

  foo := Foo{FieldA: 1, FieldB: 2, FieldC: 3}  // single line is fine when it fits

  bar(arg1, arg2, arg3)  // single line is fine when it fits

  // Bad — multiple elements on the same line in a multi-line construct
  foo := Foo{
      FieldA: 1, FieldB: 2,
      FieldC: 3,
  }

  bar(
      arg1, arg2,
      arg3,
  )
  ```

## 6. Return Statements

Return the result of a function call directly instead of storing it in a variable only to return it:

```go
// Good
func (c *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    return NewAPICommand(db, port).Run(cmd.Context())
}

// Bad — unnecessary variable and conditional
func (c *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    if err := NewAPICommand(db, port).Run(cmd.Context()); err != nil {
        return err
    }
    return nil
}
```

This also applies to non-error values:

```go
// Good
func (r *Repository) Find(ctx context.Context, id int) (*Entity, error) {
    return r.findByID(ctx, id)
}

// Bad
func (r *Repository) Find(ctx context.Context, id int) (*Entity, error) {
    entity, err := r.findByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return entity, nil
}
```

## 7. Error Checking

Use `:=` instead of `=` when checking errors in a single line if statement:

```go
// Good
if err := sqlDB.Ping(); err != nil {
    return err
}

// Bad
if err = sqlDB.Ping(); err != nil {
    return err
}
```

Rationale: `:=` makes it clear that `err` is a new variable scoped to the if block.

## 8. GORM Models

### Configuration Tags

Remove tags that match GORM's default behavior:

- `gorm:"primaryKey"` — ID fields are primary keys by default
- `gorm:"column:field_name"` — Column names are auto-generated from field names
- `gorm:"not null"` — Non-pointer fields are not null by default
- `TableName()` method — Not needed when default naming convention is sufficient

Only include tags that override default behavior:

```go
CreatedAt time.Time `gorm:"autoCreateTime:false"`
UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
```

### Model Structure

- Use exported fields (capitalize first letter)
- Keep struct names in singular form even when table names are plural:
  ```go
  // Good
  type ExtractedDataFile struct {}  // maps to extracted_data_files table

  // Bad
  type ExtractedDataFiles struct {}
  ```

### Relations

- Use pointer slices for one-to-many relations:
  ```go
  // Good
  type Parent struct {
      Children []*Child
  }

  // Bad
  type Parent struct {
      Children []Child
  }
  ```

- Do not include parent reference in child objects:
  ```go
  // Good
  type Child struct {
      ParentID int
  }

  // Bad
  type Child struct {
      ParentID int
      Parent   *Parent  // unnecessary reference
  }
  ```

- One-to-many relations should be optional (parent can exist without children)
- Let GORM handle foreign key naming and default table/column naming

### Entity Conversion

Implement `ToEntity()` method on DB models (DB → domain):

```go
func (m *DBModel) ToEntity() *domain.Entity {
    return domain.NewEntityDirectly(
        m.ID,
        m.Name,
    )
}
```

Use private conversion functions for domain → DB:

```go
func toDBModel(e *domain.Entity) *DBModel {
    return &DBModel{
        ID:   e.ID(),
        Name: e.Name(),
    }
}
```

## 9. Repository Rules

- Do not handle transactions within repository methods; transaction control belongs in the upper layer
- Use "Repository" instead of "Repo" for type names
- All repository methods must accept `context.Context` as their first parameter
- Always use `WithContext()` for GORM operations:
  ```go
  // Good
  func (r *ExtractTaskRepository) Create(ctx context.Context, task *ExtractTask) error {
      return r.db.WithContext(ctx).Create(task).Error
  }

  // Bad
  func (r *ExtractTaskRepository) Create(task *ExtractTask) error {
      return r.db.Create(task).Error
  }
  ```

## 10. Timezone Handling

Perform timezone conversions only at the boundary where they are required (e.g., S3 key generation, database persistence), not in upper layers such as usecases:

```go
// Good — usecase passes time.Now() as-is; GenerateS3Key converts to UTC internally
now := time.Now()
s3Key := extract.GenerateS3Key(source, dataType, now, ext)

// Bad — usecase converts to UTC even though GenerateS3Key handles it
now := time.Now().UTC()
s3Key := extract.GenerateS3Key(source, dataType, now, ext)
```

Each component that requires a specific timezone is responsible for converting it internally. Rationale: keeps upper layers free from infrastructure concerns and avoids redundant conversions.

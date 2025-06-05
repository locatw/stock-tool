# Backend Implementation Rules

## 1. Code Structure

- Keep related type definitions together
- Place NewXXX() functions immediately after their struct definitions

## 2. Remove Self-evident Comments

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

## 3. Utilize samber/lo for Slice Operations

- Use lo.Map for slice transformations
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

- Keep standard for loops for database operations
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
- Break method signatures with multiple parameters into multiple lines
- Example:
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

## 6. Error Checking

- Use `:=` instead of `=` when checking errors in a single line if statement
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

- This rule applies to all single-line error checks where the error variable is both assigned and checked
- The `:=` operator makes it clear that we're creating a new error variable in the if statement's scope

# Database Implementation Rules

## 1. GORM Model Rules

### 1.1 Configuration Tags

Remove tags that match GORM's default behavior:

- `gorm:"primaryKey"` - ID fields are primary keys by default
- `gorm:"column:field_name"` - Column names are auto-generated from field names
- `gorm:"not null"` - Non-pointer fields are not null by default
- TableName() method - Not needed when default naming convention is sufficient

Only include tags that override default behavior:

- Disable auto-update for created_at/updated_at
  ```go
  CreatedAt time.Time `gorm:"autoCreateTime:false"`
  UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
  ```

### 1.2 Model Structure

- Use exported fields (capitalize first letter)

## 2. GORM Relation Rules

### 2.1 Model Naming Convention

- Keep struct names in singular form even when table names are plural
  ```go
  // Good
  type ExtractedDataFile struct {}  // maps to extracted_data_files table
  
  // Bad
  type ExtractedDataFiles struct {}
  ```

### 2.2 Use Pointers for Related Objects

- Use pointer slices for one-to-many relations
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

- Do not include parent reference in child objects
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

### 2.2 Table Relations

- One-to-many relations should be optional
  - Parent can exist without children
  - Children reference parent only by ID
  - No need for direct parent object reference in children

- Use appropriate GORM defaults
  - Let GORM handle foreign key naming
  - Use default table and column naming when possible

### 2.3 Entity Conversion

- Implement ToEntity() method for DB models
  ```go
  func (m *DBModel) ToEntity() *domain.Entity {
      return domain.NewEntityDirectly(
          m.ID,
          m.Name,
          // ...other fields
      )
  }
  ```

- Use private conversion functions for domain to DB model
  ```go
  func toDBModel(e *domain.Entity) *DBModel {
      return &DBModel{
          ID:   e.ID(),
          Name: e.Name(),
          // ...other fields
      }
  }
  ```

## 3. Repository Rules

### 3.1 Transaction Control

- Do not handle transactions within repository methods
- Transaction control should be managed by the upper layer
- Repository methods should focus on basic database operations

### 3.2 Naming Conventions

- Use "Repository" instead of "Repo" for type names
- Example:
  ```go
  // Good
  type ExtractTaskRepository struct {
      db *gorm.DB
  }

  // Bad
  type ExtractTaskRepo struct {
      db *gorm.DB
  }
  ```

### 3.3 Context Usage

- All repository methods must accept context.Context as their first parameter
- Always use WithContext() for GORM operations
- Example:
  ```go
  // Good
  func (r *ExtractTaskRepository) Create(ctx context.Context, task *ExtractTask) error {
      return r.db.WithContext(ctx).Create(task).Error
  }

  // Bad
  func (r *ExtractTaskRepository) Create(task *ExtractTask) error {
      return r.db.Create(task).Error
  }

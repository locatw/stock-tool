# OpenAPI Schema Conventions

## 1. Every property must have a description

State what the field represents — its business meaning, not its type.

Do not restate the type, format, or constraints already expressed by keywords such as
`minLength`, `minimum`, `format`, or `enum`.

Good:

```yaml
timezone:
  description: IANA timezone name used to interpret scheduled times.
  type: string
```

Bad (restates the type keyword):

```yaml
timezone:
  description: A string representing the timezone.
  type: string
```

## 2. Every scalar property must have an example

Use a realistic value that a consumer could copy into a request or recognize in a response.

Good:

```yaml
timezone:
  description: IANA timezone name used to interpret scheduled times.
  type: string
  example: "Asia/Tokyo"
```

Bad (placeholder value):

```yaml
timezone:
  description: IANA timezone name used to interpret scheduled times.
  type: string
  example: "string"
```

## 3. Key ordering

### 3.1 Schema-level key order

At the schema root (top level of a YAML file), order keys as:
`description` → `type` → `required` → `properties` → `additionalProperties` → `example`.

Good:

```yaml
description: Defines when ingestion runs for a data type.
type: object
required:
  - type
  - times
properties:
  ...
additionalProperties: false
example:
  type: daily
  times:
    - "09:00"
```

Bad (`description` placed after other keys, `additionalProperties` before `properties`):

```yaml
type: object
additionalProperties: false
required:
  - type
properties:
  ...
description: Defines when ingestion runs for a data type.
```

### 3.2 Property-level key order

Place `description` first within each property block, followed by `type`, then all remaining type
keywords (`format`, `enum`, `minimum`, `minLength`, `items`, `additionalProperties`, etc.), then
`example` last.

Good:

```yaml
staleTimeoutMinutes:
  description: Minutes after the scheduled time before a run is considered stale.
  type: integer
  minimum: 0
  example: 30
```

Bad (`description` placed after type keywords):

```yaml
staleTimeoutMinutes:
  type: integer
  minimum: 0
  description: Minutes after the scheduled time before a run is considered stale.
  example: 30
```

## 4. Schema-level description for $ref targets

OpenAPI 3.0 forbids sibling keywords alongside `$ref`.
Attach `description` to the referenced schema itself, not the reference site.

Add a schema-level `description` whenever the schema is referenced via `$ref` from other schemas,
so that documentation is visible wherever the schema is used.

Good (description on the referenced schema):

```yaml
# Schedule.yaml
description: Defines when ingestion runs for a data type.
type: object
properties:
  ...
```

Bad (sibling keyword at the reference site — invalid in OpenAPI 3.0):

```yaml
# DataType.yaml
schedule:
  $ref: './Schedule.yaml'
  description: The schedule for this data type.  # not allowed
```

## 5. Schema-level example

Add a schema-level `example` only when property-level examples are insufficient to convey the
overall object structure.

Prefer property-level examples first.
Add a top-level `example` only when the relationship between properties matters and cannot be
inferred from individual property examples alone.

## 6. What NOT to write

- Do not restate the type (`type: string` — description should not say "a string").
- Do not restate constraints already expressed by keywords (`minimum: 0` — description should
  not say "must be non-negative").
- Do not restate the format (`format: uuid` — description should not say "a UUID string").

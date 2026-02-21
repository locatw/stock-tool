# Data Lineage Design

## Overview

Lineage tracking for stock-tool's lakehouse (landing → bronze → silver → gold).
Recommends batch-level processing execution records over custom per-record data IDs.

For zone definitions, see [data-persistence-architecture.md](data-persistence-architecture.md).

## Industry-Standard Lineage Tracking

| Mechanism | Description | Adoption cost |
|---|---|---|
| Iceberg snapshots | Snapshot ID + parent per commit; change history traversable automatically | Zero (built-in) |
| Iceberg manifests | File-level metadata: originating snapshot ID, operation type, timestamps | Zero (built-in) |
| Iceberg V3 row-level IDs | `_row_id` + `_last_updated_sequence_number` as built-in columns | Zero (V3) |
| Orchestration tracking | Airflow/Dagster record task dependencies and data movement | Low–medium |
| OpenLineage | Open standard for lineage metadata (run, job, dataset entities) | Medium |

## Why Custom Data IDs Are Not Suitable

- Granularity mismatch: Landing (files) → Bronze (rows) → Silver (merged rows) → Gold (aggregations)
  - Cross-zone relationships are many-to-many at every boundary
  - Per-record IDs require junction tables for each zone pair
- Write overhead: custom `_lineage_id` column on every row increases Parquet size and write latency
- Maintenance cost: 4 zone-specific entity sets, 6 relationship types, each with repository/migration/usecase logic — excessive for a small-team project
- Industry misalignment: row-level lineage IDs are for regulatory environments (HIPAA, financial auditing), not stock analysis

## Recommended Approach: Batch-Level Processing Execution Records

Record lineage at the processing execution level: each batch operation logs inputs and outputs.

### Conceptual model

```text
ProcessingExecution
  - source_zone: 'landing' | 'bronze' | 'silver'
  - target_zone: 'bronze' | 'silver' | 'gold'
  - source_identifier: input identifier (e.g., 'jquants/daily_quotes')
  - target_identifier: output identifier (e.g., 'jquants_daily_quotes')
  - input_snapshot_ref: input data reference (S3 key pattern or Iceberg snapshot ID)
  - output_snapshot_id: output Iceberg snapshot ID
  - status, started_at, finished_at, records_read, records_written, etc.
```

### Lineage traversal example

"What data produced silver daily_quotes snapshot #18?"

1. Query `processing_executions` → output snapshot #18 came from bronze snapshot #42
2. Query `processing_executions` → bronze snapshot #42 came from `landing/jquants/daily_quotes/2025/06/15/*.json`

Full chain traversable without custom IDs on the data itself.

### Alignment with existing codebase patterns

- Domain hierarchy: Extends `ExtractTask` → `ExtractTaskExecution` → `ExtractedDataS3` pattern to zone-to-zone transformations
- Immutable entities: `New*()` / `New*Directly()` convention
- GORM repository: `ToEntity()` + private `to*()` functions
- PostgreSQL `stock` schema: Same schema, migrations, indexing

## Gold Zone Lineage

Automated (Go): Same `ProcessingExecution` pattern as bronze/silver.

Interactive (Python/Marimo): Convention-based Iceberg table properties:

```python
table.update_properties({
    'lineage.source_tables': 'silver.daily_quotes,silver.brand_master',
    'lineage.notebook': 'analysis/technical_indicators.py',
    'lineage.produced_at': datetime.now().isoformat()
})
```

OpenLineage events can automate this if MLflow is introduced later.

## Phased Adoption

| Phase | Content | Timing |
|---|---|---|
| Phase 1 | Add `processing_executions` table; record during landing→bronze | Bronze implementation |
| Phase 2 | Combine Iceberg snapshot metadata + PostgreSQL records for cross-zone queries | Silver implementation |
| Phase 3 | Introduce OpenLineage (only if project scales to require it) | Project expansion |

## Decision

Do not introduce custom data IDs. Use batch-level processing execution records + Iceberg's built-in snapshot tracking.

Rationale: one new table (follows existing patterns), no extra Iceberg columns, sidesteps granularity mismatch, aligns with industry practice, fully achieves lineage goal.

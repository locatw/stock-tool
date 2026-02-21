# Data Acquisition Requirements

## Overview

General requirements for landing-layer data acquisition, applicable to any source.
Source-specific docs (e.g., [J-Quants](jquants-data-acquisition.md)) extend these with source-specific constraints.

Current state: only brand data from J-Quants API, manual CLI, no scheduling/gap detection/per-data-type config.

For landing layer conventions and storage paths, see [data-persistence-architecture.md](data-persistence-architecture.md).

## Scope

In scope: Landing layer file storage, historical backfill and gap detection,
extensible per-data-source DB configuration

Out of scope: Bronze/Silver/Gold processing, scheduler implementation (k8s CronJob manifests), source-specific API details (covered by source-specific docs)

## Daily File Storage

### Target Date in Storage Path

S3 path must use the target date (business date the data represents), not extraction timestamp.
Example: stock price data for 2025-06-02 → `2025/06/02/` directory regardless of when extraction ran.

Required changes to current implementation:

- `GenerateS3Key` in `backend/internal/domain/extract/extract.go` — receives `executionTime`, must use business date instead
- `ExtractTaskUseCase.Extract` in `backend/internal/usecase/task/extract.go` — passes `time.Now()`, must use business date instead

### Re-Run Strategy

When the same (source, data_type, target_date) is extracted multiple times:

| Strategy | Behavior | Pros | Cons |
|---|---|---|---|
| Append | New file per run (UUID/timestamp suffix) | Full audit trail | Downstream dedup complexity, storage growth |
| Overwrite | Single file per target date | Simple downstream, predictable storage | No intermediate versions |

Decision deferred. Either works without schema changes (`extracted_data_s3s` already tracks all S3 keys per execution).

### Landing Zone Storage Format

Store responses byte-for-byte in original format — no transformation, filtering,
or restructuring (see [Landing layer](data-persistence-architecture.md#landing-raw-files)).
All parsing, schema mapping, deduplication belongs to Bronze layer and later.
Rationale: reprocessing possible without re-fetch; parsing bugs cannot cause irreversible data loss.

## Acquisition Metadata

Every acquired data file must have associated metadata for origin identification and quality verification.

### Data Identity

| Item | Description | Currently Tracked |
|---|---|---|
| Source | Data source identifier (e.g., `jquants`) | Yes — `ExtractTask.source` |
| Data type | Data type name (e.g., `daily_quotes`) | Yes — `ExtractTask.dataType` |
| Target date | Business date the data represents | Yes — `ExtractTaskExecution.targetDateTime` |
| Acquired at | Timestamp when fetched | Yes — `ExtractTaskExecution.startedAt` |
| Data format | Format of stored file (JSON, CSV, etc.) | No — implicit in S3 key extension |

### Data Quality

| Item | Description | Currently Tracked |
|---|---|---|
| File size (bytes) | Size of the stored file | No |
| Checksum (SHA-256) | Hash of the stored file content | No |

### Notes

- Items may already exist via joins (`ExtractTask` → `ExtractTaskExecution` → `ExtractedDataS3`); redundant storage vs joins is a design decision
- Record count excluded — requires parsing raw data, conflicts with landing raw-preservation
  - File size as proxy
  - Record count deferred to Bronze
- Storage mechanism (DB columns, S3 metadata, sidecar files) is a design decision

## Historical Backfill

Detect dates with missing data and fetch automatically.

### Gap Detection

For each (source, data_type), compare:

- Expected dates: dates that should have data (determined by source calendar, e.g., TSE business days)
- Succeeded dates: dates with succeeded `ExtractTaskExecution` record

Expected date without succeeded execution = gap requiring backfill.
Calendar source is data-source-specific (see each source's requirements doc).

### Historical Limit

Maximum lookback period per source (e.g., subscription plan restrictions).
Configurable via DB configuration ([see Configuration](#extensible-per-data-source-configuration)).

### Execution Records for Backfill

Each backfilled date produces its own `ExtractTaskExecution` with correct `target_date_time`.
Backfill executions are indistinguishable from regular executions — no separate status/flag.
Data-type-specific backfill behavior: see [D4](#data-type-level-items).

## Concurrent Execution Control

### Basic Policy

Maximize parallelism.
Different (source, data_type, target_date) combinations execute concurrently without restriction.
Mutual exclusion applies only within the same (source, data_type, target_date) to prevent duplicate writes for identical data.
Source-level rate limits (S2) are the primary concurrency constraint.

### Exclusion Scope

Per (source, data_type, target_date). Implications:

- Different target dates for the same data type can execute in parallel
- Backfill and daily extraction can run concurrently for different target dates
- Only same-date duplicate runs are blocked

### Conflict Behavior

When an execution for the same (source, data_type, target_date) is already in progress, new requests skip immediately.
Rationale: next scheduler cycle or backfill sweep picks up the date.

### Stale Execution Detection

Executions exceeding a configurable timeout without completion are considered stale.
Stale executions release their exclusion so subsequent runs can proceed.

## Per-Source and Per-Data-Type Requirement Items

Each source-specific doc must address these items at two levels.
Defaults may be overridden or inherited by omission.

### Source-Level Items

| # | Item | Required | Description |
|---|---|---|---|
| S1 | Timezone | Required | Reference timezone for update times, calendars, target dates |
| S2 | Rate limits and mitigation | Recommended | API rate limits and strategy (wait-and-retry, spreading, concurrency cap) |
| S3 | Max concurrent executions | Recommended | Maximum parallel executions across all data types for this source (rate-limit driven) |

### Data-Type-Level Items

| # | Item | Required | Description |
|---|---|---|---|
| D1 | Update time | Required | Time of day data becomes available (array if multiple windows) |
| D2 | Update frequency | Required | `daily`, `weekly`, `irregular`, etc. |
| D3 | Processing range per execution | Required | Max dates per single execution |
| D4 | Backfill behavior | Required | Rate-limit considerations, fetch order, partial-failure, re-execution granularity |
| D5 | Backfill target | Recommended | Subject to gap detection? Default: `true` |
| D6 | Re-run strategy | Recommended | `overwrite` or `append`. Default: per [Re-Run Strategy](#re-run-strategy) decision |
| D7 | Retry policy | Recommended | Retries, backoff, error categories. Defaults: 3 retries, exponential, retry on 429/5xx/timeout |
| D8 | Empty response handling | Recommended | Zero records = success or failure? Default: `success` |
| D9 | Dependencies | Optional | Data types that must be fetched first |
| D10 | Stale execution timeout | Recommended | Time before a running execution is considered stale. Default: source-level setting |

## Extensible Per-Data-Source Configuration

Schema must support new config items without DB migrations.

### Configuration Hierarchy

Source level — historical limit, enabled flag, source-specific settings

Data type level — update time(s) (native TZ), update frequency, backfill enabled, enabled flag, data-type-specific settings

Runtime settings read by system. Checklist items map to both DB config (D1, D2,
D5) and documentation-only items (D4, D9).

### Relationship to Existing Tables

Separate from `extract_tasks`/`extract_task_executions` (execution state).
Configuration = what/when; execution tables = what happened.

### Design Constraints

- Credentials stay in env vars — only operational config in DB
- Config changes take effect without deploy or restart
- Schema accommodates new data sources without migrations

## Constraints

- Existing patterns: immutable domain entities, `stock` schema in PostgreSQL,
  `samber/do` for DI
- Forward compatibility: schema accommodates new sources without migrations
- Timezone handling: timing values stored in source market's native timezone (e.g., JST)

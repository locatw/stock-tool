# Data Acquisition

## Background

stock-tool provides a foundation for investment decisions by fetching and accumulating investment-related data from external sources.
Market data such as stock prices and brand information is updated daily, so a mechanism for continuous, gap-free acquisition is essential.
As data sources and data types will grow over time, a general-purpose acquisition framework is needed that can absorb per-source constraints such as rate limits, update timing, and trading calendars.

## Purpose

Ensure external data arrives in the landing zone at predictable, business-date-keyed S3 paths,
remains historically complete through automated gap detection and backfill, and can be extended
to new sources without code changes.

## User Stories

- As a system operator, I want files stored under the business date path so downstream processing has predictable S3 keys
- As a system operator, I want gap detection and automatic backfill so the dataset stays complete
- As a system operator, I want concurrent extractions controlled so duplicate writes are prevented
- As a developer, I want per-source configuration in the DB so new sources need no code changes

## Requirements

### FR-1: Target Date in Storage Path

S3 path uses the target date (business date the data represents), not extraction timestamp.
Example: stock price data for 2025-06-02 stores under `2025/06/02/` regardless of when extraction ran.

### FR-2: Raw File Preservation

Files stored byte-for-byte in original format without transformation, filtering, or restructuring.
Rationale: reprocessing possible without re-fetch; parsing bugs cannot cause irreversible data loss.

### FR-3: Acquisition Metadata

Every acquired data file must have associated metadata for origin identification and quality verification.

#### Data Identity

| Item | Description | Currently Tracked |
|---|---|---|
| Source | Data source identifier (e.g., `jquants`) | Yes — `ExtractTask.source` |
| Data type | Data type name (e.g., `daily_quotes`) | Yes — `ExtractTask.dataType` |
| Target date | Business date the data represents | Yes — `ExtractTaskExecution.targetDateTime` |
| Acquired at | Timestamp when fetched | Yes — `ExtractTaskExecution.startedAt` |
| Data format | Format of stored file (JSON, CSV, etc.) | No — implicit in S3 key extension |

#### Data Quality

| Item | Description | Currently Tracked |
|---|---|---|
| File size (bytes) | Size of the stored file | No |
| Checksum (SHA-256) | Hash of the stored file content | No |

#### Design Notes

- Items may already exist via joins (`ExtractTask` -> `ExtractTaskExecution` -> `ExtractedDataS3`); redundant storage vs joins is a design decision
- Record count excluded — requires parsing raw data, conflicts with landing raw-preservation
  - File size serves as proxy
  - Record count deferred to Bronze
- Storage mechanism (DB columns, S3 metadata, sidecar files) is a design decision — deferred
- Data format tracking deferred — currently implicit in S3 key extension
- File size and checksum tracking deferred (see [Out of Scope](#out-of-scope))

### FR-4: Gap Detection

Compare expected calendar dates with succeeded `ExtractTaskExecution` dates for each (source, data_type).
Expected date without succeeded execution = gap requiring backfill.
Calendar source is data-source-specific (see each source's requirements doc).

### FR-5: Backfill Execution Records

Each backfilled date has its own `ExtractTaskExecution` with correct `target_date_time`.
Backfill executions are indistinguishable from regular executions — no separate status/flag.

### FR-6: Concurrent Execution

Different (source, data_type, target_date) triplets execute concurrently without restriction.
Source-level rate limits are the primary concurrency constraint.

### FR-7: Duplicate Skip

Same (source, data_type, target_date) already in progress skips immediately; no new execution created.
Rationale: next scheduler cycle or backfill sweep picks up the date.

### FR-8: Stale Execution Recovery

Execution exceeding a configurable stale timeout releases its exclusion so subsequent runs can proceed.

### FR-9: DB-Driven Configuration

Source and data-type configuration stored in DB; changes take effect without deploy or restart.

### FR-10: Re-Run Strategy

When the same (source, data_type, target_date) is extracted multiple times:

| Strategy | Behavior | Pros | Cons |
|---|---|---|---|
| Append | New file per run (UUID/timestamp suffix) | Full audit trail | Downstream dedup complexity, storage growth |
| Overwrite | Single file per target date | Simple downstream, predictable storage | No intermediate versions |

Decision deferred. Either works without schema changes (`extracted_data_s3s` already tracks all S3 keys per execution).

### FR-11: Per-Source and Per-Data-Type Configuration Items

Each source-specific doc must address these items at two levels.

#### Source-Level Items

| # | Item | Required | Description |
|---|---|---|---|
| S1 | Timezone | Required | Reference timezone for update times, calendars, target dates |
| S2 | Rate limits and mitigation | Recommended | API rate limits and strategy (wait-and-retry, spreading, concurrency cap) |
| S3 | Max concurrent executions | Recommended | Maximum parallel executions across all data types for this source |

#### Data-Type-Level Items

| # | Item | Required | Description |
|---|---|---|---|
| D1 | Update time | Required | Time of day data becomes available (array if multiple windows) |
| D2 | Update frequency | Required | `daily`, `weekly`, `irregular`, etc. |
| D3 | Processing range per execution | Required | Max dates per single execution |
| D4 | Backfill behavior | Required | Rate-limit considerations, fetch order, partial-failure, re-execution granularity |
| D5 | Backfill target | Recommended | Subject to gap detection? Default: `true` |
| D6 | Re-run strategy | Recommended | `overwrite` or `append`. Default: per FR-10 decision |
| D7 | Retry policy | Recommended | Retries, backoff, error categories. Defaults: 3 retries, exponential, retry on 429/5xx/timeout |
| D8 | Empty response handling | Recommended | Zero records = success or failure? Default: `success` |
| D9 | Dependencies | Optional | Data types that must be fetched first |
| D10 | Stale execution timeout | Recommended | Time before a running execution is considered stale. Default: source-level setting |

- NFR-1: Config schema supports new sources without DB migrations
- NFR-2: Credentials remain in env vars; DB holds only operational config
- NFR-3: Timing values stored in source's native timezone

### Configuration Hierarchy

Source level — historical limit, enabled flag, source-specific settings

Data type level — update time(s) (native TZ), update frequency, backfill enabled, enabled flag, data-type-specific settings

Separate from `extract_tasks`/`extract_task_executions` (execution state).
Configuration = what/when; execution tables = what happened.

## Acceptance Criteria

- [ ] S3 path for business date 2025-06-02 contains `2025/06/02/` regardless of run time
- [ ] Stored file content is byte-identical to API response
- [ ] `ExtractTaskExecution.target_date_time` reflects business date, not extraction time
- [ ] Gap detection returns dates with no succeeded execution within historical limit
- [ ] Concurrent runs for different target dates proceed without blocking each other
- [ ] Duplicate run for same (source, data_type, target_date) skips immediately
- [ ] Execution exceeding stale_timeout releases lock; subsequent run proceeds
- [ ] New source configuration added without schema migration

## Design Direction

- Fix `GenerateS3Key` (`backend/internal/domain/extract/extract.go`) to accept business date
- Fix `ExtractTaskUseCase.Extract` (`backend/internal/usecase/task/extract.go`) to pass business date
- Add DB-level mutual exclusion per (source, data_type, target_date)
- Gap detection queries succeeded executions vs source calendar
- Configuration table with JSONB extension column for source- and data-type-level settings

## Use Cases

- [extract-data](usecase/extract-data.md) — Fetch and store one (source, data_type, target_date)

Use cases for gap detection, backfill, duplicate skip, and stale recovery are deferred.

## Constraints

- Existing patterns: immutable domain entities, `stock` schema in PostgreSQL, `samber/do` for DI
- Forward compatibility: schema accommodates new sources without migrations
- Timezone handling: timing values stored in source market's native timezone (e.g., JST)

## Out of Scope

- Bronze/Silver/Gold layer processing -- belongs to downstream pipeline stages
- Scheduler implementation (k8s CronJob manifests) -- infrastructure concern, separate from acquisition logic
- Source-specific API details -- covered by source-specific docs (e.g., [J-Quants](data-sources/jquants.md))
- File checksum and size tracking -- deferred; storage mechanism undecided (see [FR-3 Design Notes](#design-notes))
- Data format explicit tracking -- deferred; currently implicit in S3 key extension
- Re-run strategy decision (append vs overwrite) -- deferred; either works without schema changes (see [FR-10](#fr-10-re-run-strategy))

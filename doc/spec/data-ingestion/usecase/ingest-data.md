# Ingest Data

## Summary

Fetch data for one (source, data_type, target_date) from the source API and store it
byte-for-byte in the landing zone S3 path keyed by business date.

## Preconditions

- Configuration exists for (source, data_type) in DB
- No non-stale in-progress execution for (source, data_type, target_date)
- Source API accessible with valid credentials

## Input

| Parameter | Type | Required | Description |
|---|---|---|---|
| source | string | Yes | Data source identifier (e.g., `jquants`) |
| data_type | string | Yes | Data type name (e.g., `daily_quotes`) |
| target_date | date | Yes | Business date the data represents |

## Expected Behavior

<!-- Entity names below (ExtractTaskExecution, ExtractedDataS3) will be renamed to IngestTaskExecution, IngestedDataS3 during re-implementation -->

1. Check for non-stale in-progress execution for (source, data_type, target_date); skip if found
2. Create ExtractTaskExecution with status=in_progress and target_date_time=target_date
3. Call source API for target_date
4. Write response byte-for-byte to S3 at `<source>/<data_type>/<YYYY>/<MM>/<DD>/<filename>`
5. Record S3 key in ExtractedDataS3
6. Update ExtractTaskExecution status=succeeded

## Output

### Success

- ExtractTaskExecution status=succeeded
- S3 object created with byte-identical content to API response

### Side Effects

- ExtractTaskExecution record created with target_date_time=target_date
- ExtractedDataS3 record created referencing the S3 key

## Error Cases

| Condition | Expected Behavior | Error Type |
|---|---|---|
| Same (source, data_type, target_date) in progress | Skip; return immediately | Conflict |
| API returns 429 | Retry per D7 policy (3 retries, exponential backoff) | Transient |
| API returns 5xx | Retry per D7 policy | Transient |
| API returns 4xx (non-429) | Mark execution failed; no retry | Permanent |
| S3 write fails | Retry; mark execution failed if retries exhausted | Transient |

## Acceptance Criteria

- [ ] S3 path uses target_date, not extraction timestamp
- [ ] S3 object content matches API response byte-for-byte
- [ ] ExtractTaskExecution.target_date_time equals the business date
- [ ] Execution status transitions: in_progress → succeeded or failed

## Related Code

<!-- Paths below reflect the current codebase; will be renamed during re-implementation (domain/extract/ → domain/ingest/, usecase/task/extract.go → usecase/task/ingest.go) -->
- Domain: `backend/internal/domain/extract/`
- Use case: `backend/internal/usecase/task/`
- Repository: `backend/internal/infra/repository/`
- Storage: `backend/internal/infra/storage/`

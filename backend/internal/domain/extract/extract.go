package extract

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSucceeded ExecutionStatus = "succeeded"
	ExecutionStatusFailed    ExecutionStatus = "failed"
)

// ExtractTask defines what to extract from a source: the combination of
// source, data type, and timing that identifies a repeatable extraction job.
type ExtractTask struct {
	id        int
	source    string
	dataType  string
	timing    string
	createdAt time.Time
	updatedAt time.Time

	executions []*ExtractTaskExecution
}

func NewExtractTask(source string, dataType string, timing string) *ExtractTask {
	now := time.Now()
	return &ExtractTask{
		source:     source,
		dataType:   dataType,
		timing:     timing,
		createdAt:  now,
		updatedAt:  now,
		executions: []*ExtractTaskExecution{},
	}
}

func NewExtractTaskDirectly(
	id int,
	source string,
	dataType string,
	timing string,
	createdAt time.Time,
	updatedAt time.Time,
	executions []*ExtractTaskExecution,
) *ExtractTask {
	return &ExtractTask{
		id:         id,
		source:     source,
		dataType:   dataType,
		timing:     timing,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		executions: executions,
	}
}

func (t *ExtractTask) AddExecution(exec *ExtractTaskExecution) {
	t.executions = append(t.executions, exec)
}

func (t *ExtractTask) ID() int {
	return t.id
}

func (t *ExtractTask) Source() string {
	return t.source
}

func (t *ExtractTask) DataType() string {
	return t.dataType
}

func (t *ExtractTask) Timing() string {
	return t.timing
}

func (t *ExtractTask) CreatedAt() time.Time {
	return t.createdAt
}

func (t *ExtractTask) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *ExtractTask) Executions() []*ExtractTaskExecution {
	return t.executions
}

// ExtractTaskExecution tracks a single run of data extraction.
// Status transitions: running -> succeeded (via Succeed) or
// running -> failed (via Fail). Terminal status must not change.
type ExtractTaskExecution struct {
	id             int
	targetDateTime time.Time
	status         ExecutionStatus
	errorInfo      *string
	startedAt      *time.Time
	finishedAt     *time.Time
	createdAt      time.Time
	updatedAt      time.Time

	s3Files []*ExtractedDataS3
}

func NewRunningExecution(targetDateTime time.Time) *ExtractTaskExecution {
	now := time.Now()
	return &ExtractTaskExecution{
		targetDateTime: targetDateTime,
		status:         ExecutionStatusRunning,
		errorInfo:      nil,
		startedAt:      &now,
		finishedAt:     nil,
		createdAt:      now,
		updatedAt:      now,
		s3Files:        []*ExtractedDataS3{},
	}
}

func NewExtractTaskExecutionDirectly(
	id int,
	targetDateTime time.Time,
	status ExecutionStatus,
	errorInfo *string,
	startedAt *time.Time,
	finishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	s3Files []*ExtractedDataS3,
) *ExtractTaskExecution {
	return &ExtractTaskExecution{
		id:             id,
		targetDateTime: targetDateTime,
		status:         status,
		errorInfo:      errorInfo,
		startedAt:      startedAt,
		finishedAt:     finishedAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		s3Files:        s3Files,
	}
}

func (t *ExtractTaskExecution) Succeed() {
	now := time.Now()
	t.status = ExecutionStatusSucceeded
	t.finishedAt = &now
	t.updatedAt = now
}

func (t *ExtractTaskExecution) Fail(errorInfo string) {
	now := time.Now()
	t.status = ExecutionStatusFailed
	t.errorInfo = &errorInfo
	t.finishedAt = &now
	t.updatedAt = now
}

func (t *ExtractTaskExecution) AddS3File(file *ExtractedDataS3) {
	t.s3Files = append(t.s3Files, file)
}

func (t *ExtractTaskExecution) ID() int {
	return t.id
}

func (t *ExtractTaskExecution) TargetDateTime() time.Time {
	return t.targetDateTime
}

func (t *ExtractTaskExecution) Status() ExecutionStatus {
	return t.status
}

func (t *ExtractTaskExecution) ErrorInfo() *string {
	return t.errorInfo
}

func (t *ExtractTaskExecution) StartedAt() *time.Time {
	return t.startedAt
}

func (t *ExtractTaskExecution) FinishedAt() *time.Time {
	return t.finishedAt
}

func (t *ExtractTaskExecution) CreatedAt() time.Time {
	return t.createdAt
}

func (t *ExtractTaskExecution) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *ExtractTaskExecution) S3Files() []*ExtractedDataS3 {
	return t.s3Files
}

// ExtractedDataS3 records the S3 object key of data produced by an extraction run.
type ExtractedDataS3 struct {
	id        int
	key       string
	createdAt time.Time
	updatedAt time.Time
}

func NewExtractedDataS3(key string) *ExtractedDataS3 {
	now := time.Now()
	return &ExtractedDataS3{
		key:       key,
		createdAt: now,
		updatedAt: now,
	}
}

func NewExtractedDataS3Directly(
	id int,
	key string,
	createdAt time.Time,
	updatedAt time.Time,
) *ExtractedDataS3 {
	return &ExtractedDataS3{
		id:        id,
		key:       key,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (s *ExtractedDataS3) ID() int {
	return s.id
}

func (s *ExtractedDataS3) Key() string {
	return s.key
}

func (s *ExtractedDataS3) CreatedAt() time.Time {
	return s.createdAt
}

func (s *ExtractedDataS3) UpdatedAt() time.Time {
	return s.updatedAt
}

// GenerateS3Key generates an S3 object key following the landing layer path convention:
// landing/{source}/{data_type}/{yyyy}/{mm}/{dd}/{timestamp}_{uuid}.{ext}
func GenerateS3Key(source string, dataType string, executionTime time.Time, ext string) string {
	utc := executionTime.UTC()
	timestamp := utc.Format("20060102T150405Z")
	uuidSuffix := uuid.New().String()[:8]

	return fmt.Sprintf(
		"landing/%s/%s/%04d/%02d/%02d/%s_%s.%s",
		source,
		dataType,
		utc.Year(),
		utc.Month(),
		utc.Day(),
		timestamp,
		uuidSuffix,
		ext,
	)
}

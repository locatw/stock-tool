package extract

import "time"

type ExtractTask struct {
	id         int
	source     string
	dataType   string
	status     string
	errorInfo  string
	startedAt  *time.Time
	finishedAt *time.Time
	createdAt  time.Time
	updatedAt  time.Time
	s3Files    []*ExtractedDataS3
}

func NewExtractTask(source, dataType, status string) *ExtractTask {
	now := time.Now()
	return &ExtractTask{
		source:    source,
		dataType:  dataType,
		status:    status,
		createdAt: now,
		updatedAt: now,
		s3Files:   []*ExtractedDataS3{},
	}
}

func NewExtractTaskDirectly(
	id int,
	source string,
	dataType string,
	status string,
	errorInfo string,
	startedAt *time.Time,
	finishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	s3Files []*ExtractedDataS3,
) *ExtractTask {
	return &ExtractTask{
		id:         id,
		source:     source,
		dataType:   dataType,
		status:     status,
		errorInfo:  errorInfo,
		startedAt:  startedAt,
		finishedAt: finishedAt,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		s3Files:    s3Files,
	}
}

func (t *ExtractTask) AddS3File(file *ExtractedDataS3) {
	t.s3Files = append(t.s3Files, file)
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

func (t *ExtractTask) Status() string {
	return t.status
}

func (t *ExtractTask) ErrorInfo() string {
	return t.errorInfo
}

func (t *ExtractTask) StartedAt() *time.Time {
	return t.startedAt
}

func (t *ExtractTask) FinishedAt() *time.Time {
	return t.finishedAt
}

func (t *ExtractTask) CreatedAt() time.Time {
	return t.createdAt
}

func (t *ExtractTask) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *ExtractTask) S3Files() []*ExtractedDataS3 {
	return t.s3Files
}

type ExtractedDataS3 struct {
	id             int
	extractTaskID  int
	targetDateTime time.Time
	bucket         string
	key            string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewExtractedDataS3(targetDateTime time.Time, bucket, key string) *ExtractedDataS3 {
	now := time.Now()
	return &ExtractedDataS3{
		targetDateTime: targetDateTime,
		bucket:         bucket,
		key:            key,
		createdAt:      now,
		updatedAt:      now,
	}
}

func NewExtractedDataS3Directly(
	id int,
	extractTaskID int,
	targetDateTime time.Time,
	bucket string,
	key string,
	createdAt time.Time,
	updatedAt time.Time,
) *ExtractedDataS3 {
	return &ExtractedDataS3{
		id:             id,
		extractTaskID:  extractTaskID,
		targetDateTime: targetDateTime,
		bucket:         bucket,
		key:            key,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (s *ExtractedDataS3) ID() int {
	return s.id
}

func (s *ExtractedDataS3) ExtractTaskID() int {
	return s.extractTaskID
}

func (s *ExtractedDataS3) TargetDateTime() time.Time {
	return s.targetDateTime
}

func (s *ExtractedDataS3) Bucket() string {
	return s.bucket
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

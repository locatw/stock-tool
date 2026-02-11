package extract

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ExtractTestSuite struct {
	suite.Suite
}

func TestExtract(t *testing.T) {
	suite.Run(t, new(ExtractTestSuite))
}

func (s *ExtractTestSuite) TestGenerateS3Key() {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	type TestCase struct {
		name          string
		source        string
		dataType      string
		executionTime time.Time
		ext           string
		wantPattern   string
	}
	tests := []TestCase{
		{
			name:          "generates path with UTC time",
			source:        "jquants",
			dataType:      "brand",
			executionTime: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
			ext:           "json",
			wantPattern:   `^landing/jquants/brand/2025/06/01/20250601T120000Z_[0-9a-f]{8}\.json$`,
		},
		{
			name:          "converts non-UTC time to UTC",
			source:        "jquants",
			dataType:      "brand",
			executionTime: time.Date(2025, 6, 2, 3, 0, 0, 0, jst), // 2025-06-01 18:00:00 UTC
			ext:           "json",
			wantPattern:   `^landing/jquants/brand/2025/06/01/20250601T180000Z_[0-9a-f]{8}\.json$`,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			key := GenerateS3Key(tt.source, tt.dataType, tt.executionTime, tt.ext)
			s.Regexp(regexp.MustCompile(tt.wantPattern), key)
		})
	}

	s.Run("different calls produce different keys", func() {
		executionTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
		key1 := GenerateS3Key("jquants", "brand", executionTime, "json")
		key2 := GenerateS3Key("jquants", "brand", executionTime, "json")
		s.NotEqual(key1, key2)
	})
}

func (s *ExtractTestSuite) TestNewRunningExecution() {
	targetDateTime := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	exec := NewRunningExecution(targetDateTime)

	s.Equal(ExecutionStatusRunning, exec.Status())
	s.Equal(targetDateTime, exec.TargetDateTime())
	s.NotNil(exec.StartedAt())
	s.Nil(exec.FinishedAt())
	s.Nil(exec.ErrorInfo())
	s.Empty(exec.S3Files())
}

func (s *ExtractTestSuite) TestSucceed() {
	exec := NewRunningExecution(time.Now())

	exec.Succeed()

	s.Equal(ExecutionStatusSucceeded, exec.Status())
	s.NotNil(exec.FinishedAt())
	s.Nil(exec.ErrorInfo())
}

func (s *ExtractTestSuite) TestFail() {
	exec := NewRunningExecution(time.Now())

	exec.Fail("connection timeout")

	s.Equal(ExecutionStatusFailed, exec.Status())
	s.NotNil(exec.FinishedAt())
	s.NotNil(exec.ErrorInfo())
	s.Equal("connection timeout", *exec.ErrorInfo())
}

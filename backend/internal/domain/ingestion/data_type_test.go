package ingestion

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type DataTypeTestSuite struct {
	suite.Suite
}

func TestDataType(t *testing.T) {
	suite.Run(t, new(DataTypeTestSuite))
}

func (s *DataTypeTestSuite) TestStaleTimeout() {
	dt := NewDataType(uuid.Nil, "test", true, "daily", nil, false, 30, nil)

	s.Equal(30*time.Minute, dt.StaleTimeout())
}

func (s *DataTypeTestSuite) TestUpdate() {
	dt := NewDataType(uuid.Nil, "original", true, "daily", []string{"18:00"}, true, 30, map[string]any{"k": "v"})

	dt.Update("renamed", false, "weekly", []string{"09:00"}, false, 60, map[string]any{})

	s.Equal("renamed", dt.Name())
	s.False(dt.Enabled())
	s.Equal("weekly", dt.UpdateFrequency())
	s.Equal([]string{"09:00"}, dt.UpdateTimes())
	s.False(dt.BackfillEnabled())
	s.Equal(60, dt.StaleTimeoutMinutes())
	s.Empty(dt.Settings())
	s.True(dt.UpdatedAt().After(dt.CreatedAt()))
}

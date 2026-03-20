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

func (s *DataTypeTestSuite) mustDailySchedule(times ...TimeOfDay) Schedule {
	sched, err := NewDailySchedule(times)
	s.Require().NoError(err)
	return sched
}

func (s *DataTypeTestSuite) TestStaleTimeout() {
	dt := NewDataType(uuid.Nil, "test", true, s.mustDailySchedule("09:00"), false, 30, nil)

	s.Equal(30*time.Minute, dt.StaleTimeout())
}

func (s *DataTypeTestSuite) TestUpdate() {
	dt := NewDataType(
		uuid.Nil, "original", true,
		s.mustDailySchedule("18:00"),
		true, 30, map[string]any{"k": "v"},
	)

	dt.Update(
		"renamed", false,
		s.mustDailySchedule("09:00", "15:00"),
		false, 60, map[string]any{},
	)

	s.Equal("renamed", dt.Name())
	s.False(dt.Enabled())
	s.Equal(ScheduleTypeDaily, dt.Schedule().Type())
	s.Equal([]TimeOfDay{"09:00", "15:00"}, dt.Schedule().Times())
	s.False(dt.BackfillEnabled())
	s.Equal(60, dt.StaleTimeoutMinutes())
	s.Empty(dt.Settings())
	s.True(dt.UpdatedAt().After(dt.CreatedAt()))
}

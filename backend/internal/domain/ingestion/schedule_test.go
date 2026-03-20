package ingestion

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TimeOfDayTestSuite struct {
	suite.Suite
}

func TestTimeOfDay(t *testing.T) {
	suite.Run(t, new(TimeOfDayTestSuite))
}

func (s *TimeOfDayTestSuite) TestNew() {
	type testCase struct {
		name    string
		input   string
		wantErr bool
	}
	tests := []testCase{
		{name: "lower bound", input: "00:00"},
		{name: "typical time", input: "09:30"},
		{name: "upper bound", input: "23:59"},
		{name: "hour exceeds 23", input: "24:00", wantErr: true},
		{name: "missing leading zero", input: "9:30", wantErr: true},
		{name: "minute exceeds 59", input: "09:60", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
		{name: "non-numeric", input: "abc", wantErr: true},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tod, err := NewTimeOfDay(tc.input)
			if tc.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)
			s.Equal(TimeOfDay(tc.input), tod)
		})
	}
}

type DailyScheduleTestSuite struct {
	suite.Suite
}

func TestDailySchedule(t *testing.T) {
	suite.Run(t, new(DailyScheduleTestSuite))
}

func (s *DailyScheduleTestSuite) TestNew() {
	type testCase struct {
		name      string
		times     []TimeOfDay
		wantTimes []TimeOfDay
		wantErr   bool
	}
	tests := []testCase{
		{name: "single time", times: []TimeOfDay{"09:00"}, wantTimes: []TimeOfDay{"09:00"}},
		{name: "multiple times", times: []TimeOfDay{"09:00", "15:00"}, wantTimes: []TimeOfDay{"09:00", "15:00"}},
		{name: "nil slice", times: nil, wantErr: true},
		{name: "empty slice", times: []TimeOfDay{}, wantErr: true},
		{name: "duplicate times", times: []TimeOfDay{"09:00", "09:00"}, wantTimes: []TimeOfDay{"09:00"}},
		{name: "unsorted times", times: []TimeOfDay{"15:00", "09:00"}, wantTimes: []TimeOfDay{"09:00", "15:00"}},
		{name: "duplicate and unsorted", times: []TimeOfDay{"15:00", "09:00", "09:00"}, wantTimes: []TimeOfDay{"09:00", "15:00"}},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			sched, err := NewDailySchedule(tc.times)
			if tc.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)
			s.Equal(ScheduleTypeDaily, sched.Type())
			s.Equal(tc.wantTimes, sched.Times())
		})
	}
}

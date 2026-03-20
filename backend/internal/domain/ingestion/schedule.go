package ingestion

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/samber/lo"
)

// ScheduleType identifies how a DataType is scheduled for updates.
type ScheduleType string

const (
	ScheduleTypeDaily ScheduleType = "daily"
)

// TimeOfDay is a validated HH:MM time string.
type TimeOfDay string

var timeOfDayRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// NewTimeOfDay creates a TimeOfDay from s, returning an error if s is not a valid HH:MM string.
func NewTimeOfDay(s string) (TimeOfDay, error) {
	if !timeOfDayRe.MatchString(s) {
		return "", fmt.Errorf("invalid time of day: %s", s)
	}
	return TimeOfDay(s), nil
}

// Schedule holds the update timing configuration for a DataType.
// It specifies a list of daily execution times (HH:MM).
type Schedule struct {
	scheduleType ScheduleType
	times        []TimeOfDay
}

// NewDailySchedule creates a Schedule that runs at the given times every day.
// Returns an error if times is empty.
func NewDailySchedule(times []TimeOfDay) (Schedule, error) {
	if len(times) == 0 {
		return Schedule{}, fmt.Errorf("times must not be empty")
	}
	times = lo.Uniq(times)
	slices.Sort(times)
	return Schedule{
		scheduleType: ScheduleTypeDaily,
		times:        times,
	}, nil
}

func (s Schedule) Type() ScheduleType { return s.scheduleType }
func (s Schedule) Times() []TimeOfDay { return s.times }

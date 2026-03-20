package clock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ClockTestSuite struct {
	suite.Suite
}

func TestClock(t *testing.T) {
	suite.Run(t, new(ClockTestSuite))
}

func (s *ClockTestSuite) TestNow_Default() {
	before := time.Now()
	got := Now(context.Background())
	after := time.Now()

	s.False(got.Before(before))
	s.False(got.After(after))
}

func (s *ClockTestSuite) TestWithFixedTime() {
	fixed := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	ctx := WithFixedTime(context.Background(), fixed)

	s.Equal(fixed, Now(ctx))
	s.Equal(fixed, Now(ctx))
}

func (s *ClockTestSuite) TestWithSequentialTimes() {
	t1 := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 2, 12, 0, 0, 0, time.UTC)
	ctx := WithSequentialTimes(context.Background(), t1, t2)

	s.Equal(t1, Now(ctx))
	s.Equal(t2, Now(ctx))
}

func (s *ClockTestSuite) TestWithSequentialTimes_PanicsWhenExhausted() {
	t1 := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	ctx := WithSequentialTimes(context.Background(), t1)

	Now(ctx) // consume the only time
	s.Panics(func() { Now(ctx) })
}

func (s *ClockTestSuite) TestWithGenerator() {
	called := 0
	expected := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	ctx := WithGenerator(context.Background(), func() time.Time {
		called++
		return expected
	})

	s.Equal(expected, Now(ctx))
	s.Equal(1, called)
}

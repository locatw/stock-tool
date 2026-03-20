package idp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type IDPTestSuite struct {
	suite.Suite
}

func TestIDP(t *testing.T) {
	suite.Run(t, new(IDPTestSuite))
}

func (s *IDPTestSuite) TestNewV7_Default() {
	id := NewV7(context.Background())

	s.NotEqual(uuid.Nil, id)
	s.Equal(uuid.Version(7), id.Version())
}

func (s *IDPTestSuite) TestWithFixedID() {
	fixed := uuid.MustParse("01961a3d-0000-7000-8000-000000000001")
	ctx := WithFixedID(context.Background(), fixed)

	s.Equal(fixed, NewV7(ctx))
	s.Equal(fixed, NewV7(ctx))
}

func (s *IDPTestSuite) TestWithSequentialIDs() {
	id1 := uuid.MustParse("01961a3d-0000-7000-8000-000000000001")
	id2 := uuid.MustParse("01961a3d-0000-7000-8000-000000000002")
	ctx := WithSequentialIDs(context.Background(), id1, id2)

	s.Equal(id1, NewV7(ctx))
	s.Equal(id2, NewV7(ctx))
}

func (s *IDPTestSuite) TestWithSequentialIDs_PanicsWhenExhausted() {
	id1 := uuid.MustParse("01961a3d-0000-7000-8000-000000000001")
	ctx := WithSequentialIDs(context.Background(), id1)

	NewV7(ctx) // consume the only ID
	s.Panics(func() { NewV7(ctx) })
}

func (s *IDPTestSuite) TestWithGenerator() {
	called := 0
	expected := uuid.MustParse("01961a3d-0000-7000-8000-000000000099")
	ctx := WithGenerator(context.Background(), func() uuid.UUID {
		called++
		return expected
	})

	s.Equal(expected, NewV7(ctx))
	s.Equal(1, called)
}

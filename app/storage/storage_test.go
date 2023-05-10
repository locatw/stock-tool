package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateScan(t *testing.T) {
	input := "2023-01-02"

	actual := Date{}
	err := actual.Scan(input)

	assert.Nil(t, err)
	assert.Equal(t, Date{Year: 2023, Month: 1, Day: 2}, actual)
}

func TestDateValue(t *testing.T) {
	date := Date{Year: 2023, Month: 1, Day: 2}

	actual, err := date.Value()

	assert.Nil(t, err)
	assert.Equal(t, "2023-01-02", actual)
}

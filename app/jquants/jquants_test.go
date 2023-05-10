package jquants

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateUnmarshal(t *testing.T) {
	input := "\"2023-01-02\""

	var actual Date
	err := json.Unmarshal([]byte(input), &actual)

	assert.Nil(t, err)
	assert.Equal(t, Date{Year: 2023, Month: 1, Day: 2}, actual)
}

func TestDateMarshal(t *testing.T) {
	date := Date{Year: 2023, Month: 1, Day: 2}

	actual, err := json.Marshal(date)

	assert.Nil(t, err)
	assert.Equal(t, []byte("\"2023-01-02\""), actual)
}

package jquants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func Test_newErrorResponseBody(t *testing.T) {
	message := "error message"
	contents := fmt.Sprintf("{\"message\": \"%s\"}", message)

	body := ioutil.NopCloser(bytes.NewReader([]byte(contents)))
	defer body.Close()

	resp := &http.Response{Body: body}

	result, err := newErrorResponseBody(resp)

	assert.Nil(t, err)
	assert.Equal(t, message, result.Message)
}

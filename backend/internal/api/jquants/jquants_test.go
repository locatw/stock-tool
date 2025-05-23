package jquants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	contents := fmt.Sprintf(`{"message": "%s"}`, message)

	body := io.NopCloser(bytes.NewReader([]byte(contents)))
	defer body.Close()

	resp := &http.Response{Body: body}

	result, err := newErrorResponseBody(resp)

	assert.Nil(t, err)
	assert.Equal(t, message, result.Message)
}

type httpClientMock struct {
	mock.Mock
}

func (m *httpClientMock) Do(request *http.Request) (*http.Response, error) {
	result := m.Called(request)

	return result.Get(0).(*http.Response), result.Error(1)
}

func Test_API_AuthUser(t *testing.T) {
	token := "token"
	errorMessage := "Unexpected error. Please try again later."

	params := []struct {
		respStatusCode   int
		respBodyContents string
		respBody         interface{}
	}{
		{
			200,
			fmt.Sprintf(`{"refreshToken": "%s"}`, token),
			AuthUserResponseBody{RefreshToken: token},
		},
		{
			500,
			fmt.Sprintf(`{"message":"%s"}`, errorMessage),
			ErrorResponseBody{Message: errorMessage},
		},
	}

	for _, param := range params {
		req := AuthUserRequest{
			MailAddress: "user1@mail.test",
			Password:    "password",
		}

		rawResp := makeResponse(param.respStatusCode, param.respBodyContents)

		httpClientMock := new(httpClientMock)
		httpClientMock.On(
			"Do",
			mock.MatchedBy(requestMatcher{
				ExpectedMethod: http.MethodPost,
				ExpectedURL:    fmt.Sprintf("%s/token/auth_user", baseUrl),
				ExpectedHeader: map[string][]string{
					"Content-Type": {
						"application/json",
					},
				},
				ExpectedBodyContents: toStringPointer(
					fmt.Sprintf(`{"mailaddress":"%s","password":"%s"}`, req.MailAddress, req.Password),
				),
			}.ToFunc()),
		).Return(rawResp, nil)

		api := &API{httpClient: httpClientMock}

		resp, err := api.AuthUser(req)

		assert.Nil(t, err)
		assert.Equal(t, rawResp.StatusCode, resp.StatusCode())
		assert.Equal(t, param.respBody, resp.Body)
	}
}

func Test_API_RefreshToken(t *testing.T) {
	token := "id-token"
	errorMessage := "Unexpected error. Please try again later."

	params := []struct {
		respStatusCode   int
		respBodyContents string
		respBody         interface{}
	}{
		{
			200,
			fmt.Sprintf(`{"idToken": "%s"}`, token),
			RefreshTokenResponseBody{IDToken: token},
		},
		{
			500,
			fmt.Sprintf(`{"message":"%s"}`, errorMessage),
			ErrorResponseBody{Message: errorMessage},
		},
	}

	for _, param := range params {
		req := RefreshTokenRequest{
			RefreshToken: "refresh-token",
		}

		rawResp := makeResponse(param.respStatusCode, param.respBodyContents)

		httpClientMock := new(httpClientMock)
		httpClientMock.On(
			"Do",
			mock.MatchedBy(requestMatcher{
				ExpectedMethod:       http.MethodPost,
				ExpectedURL:          fmt.Sprintf("%s/token/auth_refresh?refreshtoken=%s", baseUrl, req.RefreshToken),
				ExpectedHeader:       map[string][]string{},
				ExpectedBodyContents: nil,
			}.ToFunc()),
		).Return(rawResp, nil)

		api := &API{httpClient: httpClientMock}

		resp, err := api.RefreshToken(req)

		assert.Nil(t, err)
		assert.Equal(t, rawResp.StatusCode, resp.StatusCode())
		assert.Equal(t, param.respBody, resp.Body)
	}
}

func makeResponse(statusCode int, bodyContents string) *http.Response {
	respBody := io.NopCloser(bytes.NewReader([]byte(bodyContents)))

	return &http.Response{
		Body:       respBody,
		StatusCode: statusCode,
	}
}

type requestMatcher struct {
	ExpectedMethod       string
	ExpectedURL          string
	ExpectedHeader       http.Header
	ExpectedBodyContents *string
}

func (m requestMatcher) Matches(request *http.Request) bool {
	if request.Method != http.MethodPost {
		return false
	}

	if request.URL.String() != m.ExpectedURL {
		return false
	}

	if len(request.Header) != len(m.ExpectedHeader) {
		return false
	}

	for expectedKey, expectedValues := range m.ExpectedHeader {
		values, ok := request.Header[expectedKey]
		if !ok {
			return false
		}

		if len(values) != len(expectedValues) {
			return false
		}

		for _, expectedValue := range expectedValues {
			found := false
			for _, value := range values {
				if value == expectedValue {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	if m.ExpectedBodyContents != nil {
		bodyData, err := io.ReadAll(request.Body)
		if err != nil {
			return false
		}

		if string(bodyData) != *m.ExpectedBodyContents {
			return false
		}
	} else {
		if request.Body != nil {
			return false
		}
	}

	return true
}

func (m requestMatcher) ToFunc() func(*http.Request) bool {
	return func(request *http.Request) bool {
		return m.Matches(request)
	}
}

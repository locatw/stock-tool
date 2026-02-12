package jquants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
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

func Test_API_AuthUser_Success(t *testing.T) {
	token := "token"

	req := AuthUserRequest{
		MailAddress: "user1@mail.test",
		Password:    "password",
	}

	rawResp := makeResponse(200, fmt.Sprintf(`{"refreshToken": "%s"}`, token))
	defer rawResp.Body.Close()

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
	assert.Equal(t, 200, resp.StatusCode())
	assert.Equal(t, AuthUserResponseBody{RefreshToken: token}, resp.Body)
	assert.NotEmpty(t, resp.RawBody)
}

func Test_API_AuthUser_Error(t *testing.T) {
	errorMessage := "Unexpected error. Please try again later."

	req := AuthUserRequest{
		MailAddress: "user1@mail.test",
		Password:    "password",
	}

	rawResp := makeResponse(500, fmt.Sprintf(`{"message":"%s"}`, errorMessage))
	defer rawResp.Body.Close()

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

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, errorMessage, err.Error())
}

func Test_API_RefreshToken_Success(t *testing.T) {
	token := "id-token"

	req := RefreshTokenRequest{
		RefreshToken: "refresh-token",
	}

	rawResp := makeResponse(200, fmt.Sprintf(`{"idToken": "%s"}`, token))
	defer rawResp.Body.Close()

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
	assert.Equal(t, 200, resp.StatusCode())
	assert.Equal(t, RefreshTokenResponseBody{IDToken: token}, resp.Body)
	assert.NotEmpty(t, resp.RawBody)
}

func Test_API_RefreshToken_Error(t *testing.T) {
	errorMessage := "Unexpected error. Please try again later."

	req := RefreshTokenRequest{
		RefreshToken: "refresh-token",
	}

	rawResp := makeResponse(500, fmt.Sprintf(`{"message":"%s"}`, errorMessage))
	defer rawResp.Body.Close()

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

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, errorMessage, err.Error())
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
			found := slices.Contains(values, expectedValue)
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

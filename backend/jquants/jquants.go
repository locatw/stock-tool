package jquants

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

const (
	baseUrl          = "https://api.jquants.com/v1"
	dateStringFormat = "%04d-%02d-%02d"
)

var (
	NotAuthorizedError = errors.New("Not authorized.")
	timezone_jst       = time.FixedZone("Asia/Tokyo", 9*60*60)
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func NewDateFromString(value string) (Date, error) {
	date := Date{}
	_, err := fmt.Sscanf(value, dateStringFormat, &date.Year, &date.Month, &date.Day)
	if err != nil {
		return Date{}, err
	}

	return date, nil
}

func (d *Date) Format() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	format := fmt.Sprintf("\"%s\"", dateStringFormat)
	_, err := fmt.Sscanf(string(data), format, &d.Year, &d.Month, &d.Day)
	if err != nil {
		return err
	}

	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", d.Format())), nil
}

type Response struct {
	Body interface{}

	rawRequest  *http.Request
	rawResponse *http.Response
}

func (r *Response) StatusCode() int {
	return r.rawResponse.StatusCode
}

func newResponse(request *http.Request, response *http.Response, responseBody interface{}) *Response {
	return &Response{
		Body:        responseBody,
		rawRequest:  request,
		rawResponse: response,
	}
}

type ErrorResponseBody struct {
	Message string `json:"message"`
}

func (r *ErrorResponseBody) Error() string {
	return r.Message
}

func newErrorResponseBody(resp *http.Response) (*ErrorResponseBody, error) {
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var body ErrorResponseBody
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

type AuthUserRequest struct {
	MailAddress string `json:"mailaddress"`
	Password    string `json:"password"`
}

type AuthUserResponseBody struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenResponseBody struct {
	IDToken string `json:"idToken"`
}

type ListBrandRequest struct {
	Code *string `json:"code"`
	Date *Date   `json:"date"`
}

type ListBrandResponseBody struct {
	Brands []BrandInfo `json:"info"`
}

type BrandInfo struct {
	Date               Date   `json:"Date"`
	Code               string `json:"Code"`
	CompanyName        string `json:"CompanyName"`
	CompanyNameEnglish string `json:"CompanyNameEnglish"`
	Sector17Code       string `json:"Sector17Code"`
	Sector17CodeName   string `json:"Sector17CodeName"`
	Sector33Code       string `json:"Sector33Code"`
	Sector33CodeName   string `json:"Sector33CodeName"`
	ScaleCategory      string `json:"ScaleCategory"`
	MarketCode         string `json:"MarketCode"`
	MarketCodeName     string `json:"MarketCodeName"`
}

type GetDailyQuoteRequest struct {
	Code          *string `json:"code"`
	Date          *Date   `json:"date"`
	From          *Date   `json:"from"`
	To            *Date   `json:"to"`
	PaginationKey *string `json:"pagination_key"`
}

type GetDailyQuoteResponseBody struct {
	DailyQuotes   []DailyQuote `json:"daily_quotes"`
	PaginationKey *string      `json:"pagination_key"`
}

type DailyQuote struct {
	Date             Date            `json:"Date"`
	Code             string          `json:"Code"`
	Open             decimal.Decimal `json:"Open"`
	High             decimal.Decimal `json:"High"`
	Low              decimal.Decimal `json:"Low"`
	Close            decimal.Decimal `json:"Close"`
	Volume           decimal.Decimal `json:"Volume"`
	TurnoverValue    decimal.Decimal `json:"TurnoverValue"`
	AdjustmentFactor decimal.Decimal `json:"AdjustmentFactor"`
	AdjustmentOpen   decimal.Decimal `json:"AdjustmentOpen"`
	AdjustmentHigh   decimal.Decimal `json:"AdjustmentHigh"`
	AdjustmentLow    decimal.Decimal `json:"AdjustmentLow"`
	AdjustmentClose  decimal.Decimal `json:"AdjustmentClose"`
	AdjustmentVolume decimal.Decimal `json:"AdjustmentVolume"`
}

func NewGetDailyQuoteRequestByCode(code string) GetDailyQuoteRequest {
	return GetDailyQuoteRequest{
		Code: toStringPointer(code),
		Date: nil,
		From: nil,
		To:   nil,
	}
}

func NewGetDailyQuoteRequestByCodeAndPeriod(code string, from Date, to Date) GetDailyQuoteRequest {
	return GetDailyQuoteRequest{
		Code: toStringPointer(code),
		Date: nil,
		From: toDatePointer(from),
		To:   toDatePointer(to),
	}
}

func NewGetDailyQuoteRequestByDate(date Date) GetDailyQuoteRequest {
	return GetDailyQuoteRequest{
		Code: nil,
		Date: toDatePointer(date),
		From: nil,
		To:   nil,
	}
}

func (r *GetDailyQuoteRequest) WithPaginationKey(value string) *GetDailyQuoteRequest {
	r.PaginationKey = toStringPointer(value)

	return r
}

type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type API struct {
	httpClient httpClient
}

func NewAPI() *API {
	return &API{httpClient: &http.Client{}}
}

func (c *API) AuthUser(request AuthUserRequest) (*Response, error) {
	req, err := newRequestBuilder(http.MethodPost, "token/auth_user").
		withJSONBody(request).
		build()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var body AuthUserResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	} else {
		var body ErrorResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	}
}

func (c *API) RefreshToken(request RefreshTokenRequest) (*Response, error) {
	req, err := newRequestBuilder(http.MethodPost, "token/auth_refresh").
		addQueryParameter("refreshtoken", request.RefreshToken).
		build()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var body RefreshTokenResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	} else {
		var body ErrorResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	}
}

func (c *API) ListBrand(idToken string, request ListBrandRequest) (*Response, error) {
	params := url.Values{}
	if request.Code != nil {
		params.Add("code", *request.Code)
	}
	if request.Date != nil {
		params.Add("date", request.Date.Format())
	}

	req, err := newRequestBuilder(http.MethodGet, "listed/info").
		withAuthorizationHeader(idToken).
		addQueryParameters(params).
		build()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var body ListBrandResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	} else {
		var body ErrorResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	}
}

func (c *API) GetDailyQuotes(idToken string, request GetDailyQuoteRequest) (*Response, error) {
	params := url.Values{}
	if request.Code != nil {
		params.Add("code", *request.Code)
	}
	if request.Date != nil {
		params.Add("date", request.Date.Format())
	}
	if request.From != nil {
		params.Add("from", request.From.Format())
	}
	if request.To != nil {
		params.Add("to", request.To.Format())
	}
	if request.PaginationKey != nil {
		params.Add("pagination_key", *request.PaginationKey)
	}

	req, err := newRequestBuilder(http.MethodGet, "prices/daily_quotes").
		withAuthorizationHeader(idToken).
		addQueryParameters(params).
		build()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var body GetDailyQuoteResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	} else {
		var body ErrorResponseBody
		err = makeResponseBody(resp, &body)
		if err != nil {
			return nil, err
		}

		return newResponse(req, resp, body), nil
	}
}

func makeResponseBody(resp *http.Response, body interface{}) error {
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyData, body)
}

type requestBuilder struct {
	method      string
	path        string
	query       url.Values
	header      http.Header
	requestBody *[]byte
	err         error
}

func newRequestBuilder(method string, path string) *requestBuilder {
	return &requestBuilder{
		method: method,
		path:   path,
		query:  url.Values{},
		header: http.Header{},
	}
}

func (b *requestBuilder) withAuthorizationHeader(value string) *requestBuilder {
	if b.err != nil {
		return b
	}

	b.header.Add("Authorization", fmt.Sprintf("Bearer %s", value))

	return b
}

func (b *requestBuilder) addQueryParameter(key string, value string) *requestBuilder {
	if b.err != nil {
		return b
	}

	b.query.Add(key, value)

	return b
}

func (b *requestBuilder) addQueryParameters(params url.Values) *requestBuilder {
	if b.err != nil {
		return b
	}

	for key, values := range params {
		for _, value := range values {
			b.addQueryParameter(key, value)
		}
	}

	return b
}

func (b *requestBuilder) withJSONBody(body interface{}) *requestBuilder {
	if b.err != nil {
		return b
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		b.err = err
		return b
	}

	b.header.Add("Content-Type", "application/json")
	b.requestBody = &reqBody

	return b
}

func (b *requestBuilder) build() (*http.Request, error) {
	if b.err != nil {
		return nil, b.err
	}

	u, err := b.makeUrl()
	if err != nil {
		return nil, err
	}

	body := b.makeBody()

	return b.makeRequest(u, body)
}

func (b *requestBuilder) makeUrl() (*url.URL, error) {
	u, err := url.Parse(fmt.Sprintf("%s/%s", baseUrl, b.path))
	if err != nil {
		return nil, err
	}

	u.RawQuery = b.query.Encode()

	return u, nil
}

func (b *requestBuilder) makeBody() io.Reader {
	if b.requestBody == nil {
		return nil
	}

	return bytes.NewBuffer(*b.requestBody)
}

func (b *requestBuilder) makeRequest(u *url.URL, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(b.method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for header, values := range b.header {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}

	return req, nil
}

func toStringPointer(value string) *string {
	p := new(string)
	*p = value

	return p
}

func toDatePointer(value Date) *Date {
	p := new(Date)
	*p = value

	return p
}

type Client struct {
	api      API
	authInfo authInfo
}

type authInfo struct {
	MailAddress  string
	Password     string
	RefreshToken *string
	IDToken      *string
}

func (ai *authInfo) ResetRefreshToken(value string) {
	ai.RefreshToken = toStringPointer(value)
}

func (ai *authInfo) ResetIDToken(value string) {
	ai.IDToken = toStringPointer(value)
}

func NewClient(mailAddress string, password string) *Client {
	return &Client{
		api: *NewAPI(),
		authInfo: authInfo{
			MailAddress:  mailAddress,
			Password:     password,
			RefreshToken: nil,
			IDToken:      nil,
		},
	}
}

func (c *Client) Login() error {
	authUserResp, err := c.authUser()
	if err != nil {
		return err
	}

	switch body := authUserResp.Body.(type) {
	case ErrorResponseBody:
		return errors.New(body.Message)
	}

	refreshTokenResp, err := c.refreshToken()
	if err != nil {
		return err
	}

	switch body := refreshTokenResp.Body.(type) {
	case ErrorResponseBody:
		return errors.New(body.Message)
	}

	return nil
}

func (c *Client) ListBrands(request ListBrandRequest) (*Response, error) {
	if !c.IsAuthorized() {
		return nil, NotAuthorizedError
	}

	resp, err := c.withRefreshToken(func() (*Response, error) {
		return c.api.ListBrand(*c.authInfo.IDToken, request)
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) GetDailyQuotes(request GetDailyQuoteRequest) (*Response, error) {
	if !c.IsAuthorized() {
		return nil, NotAuthorizedError
	}

	resp, err := c.withRefreshToken(func() (*Response, error) {
		return c.api.GetDailyQuotes(*c.authInfo.IDToken, request)
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) IsAuthorized() bool {
	return c.authInfo.RefreshToken != nil && c.authInfo.IDToken != nil
}

func (c *Client) authUser() (*Response, error) {
	resp, err := c.api.AuthUser(
		AuthUserRequest{
			MailAddress: c.authInfo.MailAddress,
			Password:    c.authInfo.Password,
		},
	)
	if err != nil {
		return nil, err
	}

	switch body := resp.Body.(type) {
	case AuthUserResponseBody:
		c.authInfo.ResetRefreshToken(body.RefreshToken)
	}

	return resp, nil
}

func (c *Client) refreshToken() (*Response, error) {
	resp, err := c.api.RefreshToken(
		RefreshTokenRequest{
			RefreshToken: *c.authInfo.RefreshToken,
		},
	)
	if err != nil {
		return nil, err
	}

	switch body := resp.Body.(type) {
	case RefreshTokenResponseBody:
		c.authInfo.ResetIDToken(body.IDToken)
	}

	return resp, nil
}

func (c *Client) withRefreshToken(requestFunc func() (*Response, error)) (*Response, error) {
	for {
		resp, err := requestFunc()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() == 401 {
			err = c.reAuth()
			if err != nil {
				return nil, err
			}

			continue
		}

		return resp, nil
	}
}

func (c *Client) reAuth() error {
	refreshTokenResp, err := c.refreshToken()
	if err != nil {
		return err
	}

	switch body := refreshTokenResp.Body.(type) {
	case AuthUserResponseBody:
		return nil
	case ErrorResponseBody:
		// TODO: investigate the HTTP status code when the refresh token has expired.
		if refreshTokenResp.StatusCode() != 401 {
			message := "Failed to refresh token: " + body.Message
			return errors.New(message)
		}
	default:
		message := fmt.Sprintf("Unknown response body: %+v", body)
		return errors.New(message)
	}

	authUserResp, err := c.authUser()
	if err != nil {
		return err
	}

	switch body := authUserResp.Body.(type) {
	case AuthUserResponseBody:
		return nil
	case ErrorResponseBody:
		message := "Failed to authenticate user: " + body.Message
		return errors.New(message)
	default:
		message := fmt.Sprintf("Unknown response body: %+v", body)
		return errors.New(message)
	}
}

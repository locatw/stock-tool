package jquants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

const (
	baseUrl          = "https://api.jquants.com/v1"
	dateStringFormat = "%04d-%02d-%02d"
)

var (
	timezone_jst = time.FixedZone("Asia/Tokyo", 9*60*60)
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

type AuthUserRequest struct {
	MailAddress string `json:"mailaddress"`
	Password    string `json:"password"`
}

type AuthUserResponse struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenResponse struct {
	IDToken string `json:"idToken"`
}

type ListBrandRequest struct {
	Code *string `json:"code"`
	Date *Date   `json:"date"`
}

type ListBrandResponse struct {
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

type GetDailyQuoteResponse struct {
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

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{}}
}

func (c *Client) AuthUser(request AuthUserRequest) (AuthUserResponse, error) {
	req, err := newRequestBuilder(http.MethodPost, "token/auth_user").
		withJSONBody(request).
		build()
	if err != nil {
		return AuthUserResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AuthUserResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return AuthUserResponse{}, err
	}

	var result AuthUserResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return AuthUserResponse{}, err
	}

	return result, nil
}

func (c *Client) RefreshToken(request RefreshTokenRequest) (RefreshTokenResponse, error) {
	req, err := newRequestBuilder(http.MethodPost, "token/auth_refresh").
		addQueryParameter("refreshtoken", request.RefreshToken).
		build()
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	var result RefreshTokenResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	return result, nil
}

func (c *Client) ListBrand(idToken string, request ListBrandRequest) (ListBrandResponse, error) {
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
		return ListBrandResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ListBrandResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ListBrandResponse{}, err
	}

	var result ListBrandResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return ListBrandResponse{}, err
	}

	return result, nil
}

func (c *Client) GetDailyQuotes(idToken string, request GetDailyQuoteRequest) (GetDailyQuoteResponse, error) {
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
		return GetDailyQuoteResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetDailyQuoteResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GetDailyQuoteResponse{}, err
	}

	if request.PaginationKey == nil {
		ioutil.WriteFile("debug-daily-quotes.json", respBody, 0775)
	} else {
		ioutil.WriteFile(fmt.Sprintf("debug-daily-quotes_%s.json", *request.PaginationKey), respBody, 0775)
	}

	var result GetDailyQuoteResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return GetDailyQuoteResponse{}, err
	}

	return result, nil
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

func LoadBrands() (ListBrandResponse, error) {
	jsonFile, err := os.Open("brands.json")
	if err != nil {
		return ListBrandResponse{}, err
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return ListBrandResponse{}, err
	}

	var brands ListBrandResponse
	if err := json.Unmarshal(jsonBytes, &brands); err != nil {
		return ListBrandResponse{}, err
	}

	return brands, nil
}

func LoadPrices() (GetDailyQuoteResponse, error) {
	jsonFile, err := os.Open("daily-quotes.json")
	if err != nil {
		return GetDailyQuoteResponse{}, err
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return GetDailyQuoteResponse{}, err
	}

	var prices GetDailyQuoteResponse
	if err := json.Unmarshal(jsonBytes, &prices); err != nil {
		return GetDailyQuoteResponse{}, err
	}

	return prices, nil
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

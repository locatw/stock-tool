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
)

const (
	baseUrl = "https://api.jquants.com/v1"
)

var (
	timezone_jst = time.FixedZone("Asia/Tokyo", 9*60*60)
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func (d *Date) Format() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	_, err := fmt.Sscanf(string(data), "\"%04d-%02d-%02d\"", &d.Year, &d.Month, &d.Day)
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

func (r ListBrandRequest) WithCode(code string) ListBrandRequest {
	value := new(string)
	*value = code
	r.Code = value

	return r
}

func (r ListBrandRequest) WithDate(date Date) ListBrandRequest {
	value := new(Date)
	*value = date
	r.Date = value

	return r
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

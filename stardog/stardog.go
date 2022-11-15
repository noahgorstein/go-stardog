package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	defaultBaseURL   = "http://localhost:5820/"
	defaultUserAgent = "stardog-go"
)

// Client manages communications with the Stardog API
type Client struct {
	client    *http.Client
	UserAgent string
	Username  string
	password  string
	BaseURL   *url.URL

	common service

	//Services for talking to different parts of the Stardog API
	Security    *SecurityService
	StoredQuery *StoredQueryService
	ServerAdmin *ServerAdminService
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	fmt.Println(baseURL.String())
	fmt.Println(baseURL.Path)
	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: defaultUserAgent}
	c.common.client = c
	c.Security = (*SecurityService)(&c.common)
	c.StoredQuery = (*StoredQueryService)(&c.common)
	c.ServerAdmin = (*ServerAdminService)(&c.common)
	return c
}

type service struct {
	client *Client
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BasicAuthTransport is an http.RoundTripper that authenticates all requests
// using HTTP Basic Authentication with the provided username and password.
type BasicAuthTransport struct {
	Username string
	Password string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := setCredentialsAsHeaders(req, t.Username, t.Password)
	return t.transport().RoundTrip(req2)
}

// Client returns an *http.Client that makes requests that are authenticated
// using HTTP Basic Authentication.
func (t *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

func setCredentialsAsHeaders(req *http.Request, username, password string) *http.Request {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	convertedRequest := new(http.Request)
	*convertedRequest = *req
	convertedRequest.Header = make(http.Header, len(req.Header))

	for k, s := range req.Header {
		convertedRequest.Header[k] = append([]string(nil), s...)
	}
	convertedRequest.SetBasicAuth(username, password)
	return convertedRequest
}

type BearerAuthTransport struct {
	Bearer    string
	Transport http.RoundTripper
}

func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := setBearerAuthHeaders(req, t.Bearer)
	return t.transport().RoundTrip(req2)
}

func (t *BearerAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

func (t *BearerAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func setBearerAuthHeaders(req *http.Request, bearer string) *http.Request {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	convertedRequest := new(http.Request)
	*convertedRequest = *req
	convertedRequest.Header = make(http.Header, len(req.Header))

	for k, s := range req.Header {
		convertedRequest.Header[k] = append([]string(nil), s...)
	}
	convertedRequest.Header.Set("Authorization", fmt.Sprintf("bearer %s", bearer))
	return convertedRequest
}

func (c *Client) NewRequest(method, urlStr, contentType string, acceptType string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	if acceptType != "" {
		req.Header.Set("Accept", "application/json")
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Response is a Stardog API response. This wraps the standard http.Response
type Response struct {
	*http.Response
}

func (client *Client) sendRequest(request *http.Request, v interface{}) error {
	request.Header.Add("Accept", "application/json")
	request.SetBasicAuth(client.Username, client.password)

	response, err := client.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed) {

		var errorResponse errorResponse
		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return errors.New(errorResponse.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", response.StatusCode)
	}

	fullResponse := v
	if err = json.NewDecoder(response.Body).Decode(&fullResponse); err != nil {
		// response body is empty
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// BareDo sends an API request and lets you handle the api response. If an error
// or API Error occurs, the error will contain more information. Otherwise you
// are supposed to read and close the response's Body.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is
// canceled or times out, ctx.Err() will be returned.
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	r := newResponse(resp)
	err = CheckResponse(resp)
	return r, err
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer interface,
// the raw response body will be written to v, without attempting to first
// decode it. If v is nil, and no error hapens, the response is returned as is.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it
// is canceled or times out, ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// parseBoolResponse determines the boolean result from a GitHub API response.
// Several GitHub API methods return boolean responses indicated by the HTTP
// status code in the response (true indicated by a 204, false indicated by a
// 404). This helper function will determine that result and hide the 404
// error if present. Any other error will be returned through as-is.
func parseBoolResponse(err error) (bool, error) {
	if err == nil {
		return true, nil
	}

	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode == http.StatusNotFound {
		// Simply false. In this one case, we do not pass the error through.
		return false, nil
	}

	// some other real error occurred
	return false, err
}

/*
An ErrorResponse reports one or more errors caused by an API request.

Stardog API docs: https://stardog-union.github.io/http-docs/#section/Error-Codes
*/
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("[%v] [%d] [%v]",
		r.Response.Request.Method,
		r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range.
// API error responses are expected to have response
// body, and a JSON response body that maps to ErrorResponse.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err == nil && data != nil {
		fmt.Println(fmt.Sprintf("%v", string(data)))
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

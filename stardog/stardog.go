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
	DefaultServerURL = "http://localhost:5820/"
	defaultUserAgent = "stardog-go"
)

var errNonNilContext = errors.New("context must be non-nil")

// Client manages communications with the Stardog API
type Client struct {
	client    *http.Client
	UserAgent string
	Username  string
	BaseURL   *url.URL

	common service

	//Services for talking to different parts of the Stardog API
	DatabaseAdmin *DatabaseAdminService
	Security      *SecurityService
	ServerAdmin   *ServerAdminService
	Transaction   *TransactionService
}

// Client returns the http.Client used by this Stardog client.
func (c *Client) Client() *http.Client {
	clientCopy := *c.client
	return &clientCopy
}

type service struct {
	client *Client
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

// BearerAuthTransport is an http.RoundTripper that authenticates all requests
// using Bearer Authentication with the provided bearer token.
type BearerAuthTransport struct {
	BearerToken string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

type requestHeaderOptions struct {
	ContentType string
	Accept      string
}

// NewClient returns a new Stardog API client. If a nil httpClient is provided, a new http.Client will be used.
// To make authenticated API calls, provide an http.Client that will perform the authentication for you.
func NewClient(serverURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	serverEndpoint, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(serverEndpoint.Path, "/") {
		serverEndpoint.Path += "/"
	}

	c := &Client{client: httpClient, BaseURL: serverEndpoint, UserAgent: defaultUserAgent}
	c.common.client = c
	c.DatabaseAdmin = (*DatabaseAdminService)(&c.common)
	c.Security = (*SecurityService)(&c.common)
	c.ServerAdmin = (*ServerAdminService)(&c.common)
	c.Transaction = (*TransactionService)(&c.common)
	return c, nil
}

func (c *Client) NewMultipartFormDataRequest(method string, urlStr string, headerOpts *requestHeaderOptions, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		//revive:disable-next-line:error-strings
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if body != nil && headerOpts != nil {
		if strings.Contains(headerOpts.ContentType, "multipart/form-data") {
			buf, ok := body.(*bytes.Buffer)
			if ok {
				reader := strings.NewReader(buf.String())
				req, err := http.NewRequest(method, u.String(), reader)
				req.Header.Set("Content-Type", headerOpts.ContentType)
				return req, err
			}
		}
	}
	return nil, errors.New("Missing 'Content-Type multipart/form-data' header")
}

func (c *Client) NewRequest(method string, urlStr string, headerOpts *requestHeaderOptions, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		//revive:disable-next-line:error-strings
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

	if body != nil && headerOpts != nil {
		req.Header.Set("Content-Type", headerOpts.ContentType)
	}

	if headerOpts != nil {
		if headerOpts.Accept != "" {
			req.Header.Set("Accept", headerOpts.Accept)
		}
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
	if ctx == nil {
		return nil, errNonNilContext
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if e, ok := err.(*url.Error); ok {
			return nil, e
		}
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

	if u.RawQuery != "" {
		u.RawQuery = u.RawQuery + "&" + qs.Encode()
	} else {
		u.RawQuery = qs.Encode()
	}
	return u.String(), nil
}

// parseBoolResponse determines the boolean result from a Stardog API response.
// Some Stardog API methods return boolean responses indicated by the HTTP
// status code in the response. Any error will be returned through as-is.
func parseBoolResponse(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode >= http.StatusBadRequest {
		return false, err
	}
	return false, err
}

// compareHTTPResponse returns whether two http.Response objects are equal or not.
// Currently, only StatusCode is checked. This function is used when implementing the
// Is(error) bool interface for the custom error types in this package.
func compareHTTPResponse(r1, r2 *http.Response) bool {
	if r1 == nil && r2 == nil {
		return true
	}

	if r1 != nil && r2 != nil {
		return r1.StatusCode == r2.StatusCode
	}
	return false
}

/*
An ErrorResponse reports an error caused by an API request.

Stardog API docs: https://stardog-union.github.io/http-docs/#section/Error-Codes
*/
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
	Code     string         `json:"code"`    // Stardog error code
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("[%v] [%v] | [%v] [%v]",
		r.Response.Request.Method,
		r.Response.Status, r.Message, r.Code)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range.
// API error responses are expected to have response
// body, and a JSON response body that maps to ErrorResponse.
func CheckResponse(r *http.Response) error {
	//revive:disable-next-line:add-constant
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			return errors.New(string(data))
		}
	}
	return errorResponse
}

// Is returns whether the provided error equals this error.
func (r *ErrorResponse) Is(target error) bool {
	v, ok := target.(*ErrorResponse)
	if !ok {
		return false
	}

	if r.Message != v.Message || (r.Code != v.Code) ||
		!compareHTTPResponse(r.Response, v.Response) {
		return false
	}

	return true
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

// RoundTrip implements the RoundTripper interface.
func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := setBearerAuthHeaders(req, t.BearerToken)
	return t.transport().RoundTrip(req2)
}

func (t *BearerAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// Client returns an *http.Client that makes requests that are authenticated
// using Bearer Authentication.
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

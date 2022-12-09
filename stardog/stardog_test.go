package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// baseURLPath is a non-empty Client.BaseURL path to use during tests,
// to ensure relative URLs are used for all endpoints.
const baseURLPath = "/stardog-testing"

// setup sets up a test HTTP server along with a stardog.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the Stardog client being tested and is
	// configured to use test server.
	client, _ = NewClient(DefaultServerURL, nil)
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = url

	return client, mux, server.URL, server.Close
}

func TestNewClient(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)

	if got, want := c.BaseURL.String(), DefaultServerURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, defaultUserAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}

	c2, _ := NewClient(DefaultServerURL, nil)
	if c.client == c2.client {
		t.Error("NewClient returned same http.Clients, but they should differ")
	}
}

func TestNewClient_trailingSlashServerURL(t *testing.T) {
	serverURL := "http://localhost:5821"
	c, _ := NewClient(serverURL, nil)

	if got, want := c.BaseURL.String(), fmt.Sprintf("%v/", serverURL); got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
}

func TestNewClient_invalidServerURL(t *testing.T) {
	invalidServerURL := "%%%"
	_, err := NewClient(invalidServerURL, nil)
	if err == nil {
		t.Errorf("NewClient returned no error")
	}

}

func TestClient(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)
	c2 := c.Client()
	if c.client == c2 {
		t.Error("Client returned same http.Client, but should be different")
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

// Test function under NewRequest failure and then s.client.Do failure.
// Method f should be a regular call that would normally succeed, but
// should return an error when NewRequest or s.client.Do fails.
func testNewRequestAndDoFailure(t *testing.T, methodName string, client *Client, f func() (*Response, error)) {
	t.Helper()
	if methodName == "" {
		t.Error("testNewRequestAndDoFailure: must supply method methodName")
	}

	// invalid BaseURL (i.e. one without a trailing slash)
	// this will make NewRequest fail
	client.BaseURL.Path = ""
	resp, err := f()
	if resp != nil {
		t.Errorf("client.BaseURL.Path='' %v resp = %#v, want nil", methodName, resp)
	}
	if err == nil {
		t.Errorf("client.BaseURL.Path='' %v err = nil, want error", methodName)
	}

	client.BaseURL.Path = baseURLPath + "/"
	resp, err = f()
	if err == nil {
		t.Errorf("%v err = nil, want error", methodName)
	}
}

// Test how bad options are handled. Method f under test should
// return an error.
func testBadOptions(t *testing.T, methodName string, f func() error) {
	t.Helper()
	if methodName == "" {
		t.Error("testBadOptions: must supply method methodName")
	}
	if err := f(); err == nil {
		t.Errorf("bad options %v err = nil, want error", methodName)
	}
}

func TestNewRequest(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)

	inURL, outURL := "/foo", DefaultServerURL+"foo"
	inBody, outBody := &isEnabledResponse{Enabled: true}, `{"enabled":true}`+"\n"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := c.NewRequest("GET", inURL, &headerOpts, inBody)

	// test that relative URL was expanded
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	// test that body was JSON encoded
	body, _ := io.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%v) Body is %v, want %v", inBody, got, want)
	}

	userAgent := req.Header.Get("User-Agent")

	// test that default user-agent is attached to the request
	if got, want := userAgent, c.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestNewRequest_invalidJSON(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)

	type T struct {
		A map[interface{}]interface{}
	}
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
		Accept:      mediaTypeApplicationJSON,
	}
	_, err := c.NewRequest("GET", ".", &headerOpts, &T{})

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected a JSON error; got %#v.", err)
	}
}

func testURLParseError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	_, err := c.NewRequest("GET", ":", &headerOpts, nil)
	testURLParseError(t, err)
}

func TestNewRequest_badMethod(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	if _, err := c.NewRequest("FAKE\nMETHOD", ".", &headerOpts, nil); err == nil {
		t.Fatal("NewRequest returned nil; expected error")
	}
}

func TestBareDo_returnsOpenBody(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	expectedBody := "Hello from the other side !"

	mux.HandleFunc("/test-url", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, expectedBody)
	})

	ctx := context.Background()
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := client.NewRequest("GET", "test-url", &headerOpts, nil)
	if err != nil {
		t.Fatalf("client.NewRequest returned error: %v", err)
	}

	resp, err := client.BareDo(ctx, req)
	if err != nil {
		t.Fatalf("client.BareDo returned error: %v", err)
	}

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	if string(got) != expectedBody {
		t.Fatalf("Expected %q, got %q", expectedBody, string(got))
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("resp.Body.Close() returned error: %v", err)
	}

}

func TestBareDo_URLError(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	ctx := context.Background()
  req, err := client.NewRequest("GET", "htp://hello", nil, nil)
	if err != nil {
		t.Fatalf("client.NewRequest returned error: %v", err)
	}
	_, err = client.BareDo(ctx, req)
	_, isURLError := err.(*url.Error)

	if !isURLError {
		t.Fatalf("was not a url.Error, got %T", err)
	}
}

func TestAddOptions_QueryValues(t *testing.T) {
	if _, err := addOptions("url", ""); err == nil {
		t.Error("addOptions err = nil, want error")
	}
	var nilPtr *bytes.Buffer
	_, err := addOptions("url", nilPtr)
	if err != nil {
		t.Errorf("addOptions err = %v, want nil", err)
	}
}

// Test whether the marshaling of v produces JSON that corresponds
// to the want string.
func testJSONMarshal(t *testing.T, v interface{}, want string) {
	t.Helper()
	// Unmarshal the wanted JSON, to verify its correctness, and marshal it back
	// to sort the keys.
	u := reflect.New(reflect.TypeOf(v)).Interface()
	if err := json.Unmarshal([]byte(want), &u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v: %v", want, err)
	}
	w, err := json.Marshal(u)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", u)
	}

	// Marshal the target value.
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", v)
	}

	if string(w) != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}
}

func TestErrorResponse_Marshal(t *testing.T) {
	testJSONMarshal(t, &ErrorResponse{}, "{}")

	u := &ErrorResponse{
		Message: "unknown",
		Code:    "000012",
	}

	want := `{
		"message": "unknown",
    "code": "000012"
	}`

	testJSONMarshal(t, u, want)
}

func TestCheckResponse(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"message":"m", "code": "1234"}`)),
	}
	err := CheckResponse(res).(*ErrorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	want := &ErrorResponse{
		Response: res,
		Message:  "m",
		Code:     "1234",
	}
	if !errors.Is(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

func TestSetCredentialsAsHeaders(t *testing.T) {
	req := new(http.Request)
	username, password := "admin", "admin"
	modifiedRequest := setCredentialsAsHeaders(req, username, password)

	actualUsername, actualPassword, ok := modifiedRequest.BasicAuth()
	if !ok {
		t.Errorf("request does not contain basic credentials")
	}

	if actualUsername != username {
		t.Errorf("id is %s, want %s", actualUsername, username)
	}

	if actualPassword != password {
		t.Errorf("secret is %s, want %s", actualPassword, password)
	}
}

func TestBasicAuthTransport(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	username, password := "username", "password"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			t.Errorf("request does not contain basic auth credentials")
		}
		if u != username {
			t.Errorf("request contained basic auth username %q, want %q", u, username)
		}
		if p != password {
			t.Errorf("request contained basic auth password %q, want %q", p, password)
		}
	})

	tp := &BasicAuthTransport{
		Username: username,
		Password: password,
	}
	basicAuthClient, _ := NewClient(DefaultServerURL, tp.Client())
	basicAuthClient.BaseURL = client.BaseURL
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := basicAuthClient.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	basicAuthClient.Do(ctx, req, nil)
}

func TestBasicAuthTransport_transport(t *testing.T) {
	// default transport
	tp := &BasicAuthTransport{}
	if tp.transport() != http.DefaultTransport {
		t.Errorf("Expected http.DefaultTransport to be used.")
	}

	// custom transport
	tp = &BasicAuthTransport{
		Transport: &http.Transport{},
	}
	if tp.transport() == http.DefaultTransport {
		t.Errorf("Expected custom transport to be used.")
	}
}

func TestSetBearerAuthHeaders(t *testing.T) {
	req := new(http.Request)
	token := "12345"
	modifiedRequest := setBearerAuthHeaders(req, token)

	authHeader := modifiedRequest.Header.Get("Authorization")

	if !strings.Contains(authHeader, fmt.Sprintf("bearer %s", token)) {
		t.Errorf("Authorization header does not contain: bearer %s", token)
	}
}

func TestBearerAuthTransport(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	token := "12345"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.Contains(authHeader, fmt.Sprintf("bearer %s", token)) {
			t.Errorf("Authorization header does not contain: bearer %s", token)
		}
	})

	tp := &BearerAuthTransport{
		BearerToken: "12345",
	}
	bearerAuthClient, _ := NewClient(DefaultServerURL, tp.Client())
	bearerAuthClient.BaseURL = client.BaseURL
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := bearerAuthClient.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	bearerAuthClient.Do(ctx, req, nil)
}

func TestBearerAuthTransport_transport(t *testing.T) {
	// default transport
	tp := &BearerAuthTransport{}
	if tp.transport() != http.DefaultTransport {
		t.Errorf("Expected http.DefaultTransport to be used.")
	}

	// custom transport
	tp = &BearerAuthTransport{
		Transport: &http.Transport{},
	}
	if tp.transport() == http.DefaultTransport {
		t.Errorf("Expected custom transport to be used.")
	}
}

func TestParseBooleanResponse_true(t *testing.T) {
	result, err := parseBoolResponse(nil)
	if err != nil {
		t.Errorf("parseBoolResponse returned error: %+v", err)
	}

	if want := true; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestParseBooleanResponse_error(t *testing.T) {
	v := &ErrorResponse{Response: &http.Response{StatusCode: http.StatusBadRequest}}
	result, err := parseBoolResponse(v)

	if err == nil {
		t.Errorf("Expected error to be returned.")
	}

	if want := false; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestCompareHttpResponse(t *testing.T) {
	testcases := map[string]struct {
		h1       *http.Response
		h2       *http.Response
		expected bool
	}{
		"both are nil": {
			expected: true,
		},
		"both are non nil - same StatusCode": {
			expected: true,
			h1:       &http.Response{StatusCode: 200},
			h2:       &http.Response{StatusCode: 200},
		},
		"both are non nil - different StatusCode": {
			expected: false,
			h1:       &http.Response{StatusCode: 200},
			h2:       &http.Response{StatusCode: 404},
		},
		"one is nil, other is not": {
			expected: false,
			h2:       &http.Response{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			v := compareHTTPResponse(tc.h1, tc.h2)
			if tc.expected != v {
				t.Errorf("Expected %t, got %t for (%#v, %#v)", tc.expected, v, tc.h1, tc.h2)
			}
		})
	}
}

func TestErrorResponse_Is(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{},
		Message:  "m",
		Code:     "1",
	}
	testcases := map[string]struct {
		wantSame   bool
		otherError error
	}{
		"errors are same": {
			wantSame: true,
			otherError: &ErrorResponse{
				Response: &http.Response{},
				Message:  "m",
				Code:     "1",
			},
		},
		"errors have different values - Message": {
			wantSame: false,
			otherError: &ErrorResponse{
				Response: &http.Response{},
				Message:  "m1",
				Code:     "1",
			},
		},
		"errors have different values - Code": {
			wantSame: false,
			otherError: &ErrorResponse{
				Response: &http.Response{},
				Message:  "m",
				Code:     "2",
			},
		},
		"errors have different values - Response is nil": {
			wantSame: false,
			otherError: &ErrorResponse{
				Message: "m",
			},
		},
		"errors have different types": {
			wantSame:   false,
			otherError: errors.New("Stardog"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if tc.wantSame != err.Is(tc.otherError) {
				t.Errorf("Error = %#v, want %#v", err, tc.otherError)
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}
}

func TestNewRequest_emptyBody(t *testing.T) {
	c, _ := NewClient(DefaultServerURL, nil)
	var i interface{}
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := c.NewRequest("GET", "some-url", &headerOpts, i)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Fatalf("constructed request contains a non-nil Body")
	}
}
func TestDo(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"A":"a"}`)
	})

	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	body := new(foo)
	ctx := context.Background()
	client.Do(ctx, req, body)

	want := &foo{"a"}
	if !cmp.Equal(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_nilContext(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	_, err := client.Do(nil, req, nil)

	if !errors.Is(err, errNonNilContext) {
		t.Errorf("Expected context must be non-nil error")
	}
}

func TestDo_httpError(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	resp, err := client.Do(ctx, req, nil)

	if err == nil {
		t.Fatal("Expected HTTP 400 error, got no error.")
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected HTTP 400 error, got %d status code.", resp.StatusCode)
	}
}

func TestDo_noContent(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	var body json.RawMessage

	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	_, err := client.Do(ctx, req, &body)
	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}
}

func TestDo_invalidJsonBodyError(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	invalidJsonBody := json.RawMessage(`{"percent": 100%}`)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(invalidJsonBody)
	})

	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	_, err := client.Do(ctx, req, &invalidJsonBody)
	if err == nil {
		t.Fatalf("Do returned nil error.")
	}
}

func TestDo_contextCancelledError(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	headerOpts := requestHeaderOptions{
		Accept: mediaTypePlainText,
	}
	req, _ := client.NewRequest("GET", ".", &headerOpts, nil)
	ctx := context.Background()
	ctxWithDeadline, cancelContext := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	cancelContext()
	_, err := client.Do(ctxWithDeadline, req, nil)

	want := context.Canceled
	if !errors.Is(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

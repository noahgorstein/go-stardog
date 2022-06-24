package stardog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	StardogDefaultURL = "http://localhost:5820"
)

// Client manages communications with the Stardog API
type Client struct {
	HTTPClient *http.Client
	Username   string
	password   string
	BaseURL    string

	common service

	//Services for talking to different parts of the Stardog API
	Security    *SecurityService
	StoredQuery *StoredQueryService
	ServerAdmin *ServerAdminService
}

//  NewClient returns a new Stardog API client.
func NewClient(baseURL, username, password string) *Client {

	c := &Client{
		Username:   username,
		password:   password,
		BaseURL:    StardogDefaultURL,
		HTTPClient: &http.Client{},
	}
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

func (client *Client) sendRequest(request *http.Request, v interface{}) error {
	request.Header.Add("Content-type", "application/json")
	request.Header.Add("Accept", "application/json")
	request.SetBasicAuth(client.Username, client.password)

	response, err := client.HTTPClient.Do(request)
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

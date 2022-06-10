package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func NewStoredQuery(name string, description string, database string,
	query string, creator string, reasoning bool, shared bool) *StoredQuery {
	return &StoredQuery{
		Name:        name,
		Description: description,
		Database:    database,
		Query:       query,
		Creator:     creator,
		Reasoning:   reasoning,
		Shared:      shared,
	}
}

// Represents a stored query
type StoredQuery struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Database    string `json:"database"`
	Query       string `json:"query"`
	Shared      bool   `json:"shared"`
	Reasoning   bool   `json:"reasoning"`
	Creator     string `json:"creator"`
}

// GetStoredQueries represents the response from the list stored queries endpoint
type GetStoredQueries struct {
	Queries *[]StoredQuery `json:"queries"`
}

// GetStoredQueries lists the stored queries that are accessible to the authenticated client
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Queries/operation/listStoredQueries
func (client *Client) GetStoredQueries(ctx context.Context) (*GetStoredQueries, error) {

	url := fmt.Sprintf("%s/admin/queries/stored", client.BaseURL)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var storedQueries GetStoredQueries
	if err := client.sendRequest(request, &storedQueries); err != nil {
		return nil, err
	}

	return &storedQueries, nil
}

// Add stored query, overwriting if a query with that name already exists
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Queries/operation/updateStoredQuery
func (client *Client) UpdateStoredQuery(ctx context.Context, sq StoredQuery) (bool, error) {
	url := fmt.Sprintf("%s/admin/queries/stored", client.BaseURL)

	requestBody, _ := json.Marshal(sq)
	payloadBuf := bytes.NewBuffer(requestBody)

	request, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}

	var v struct{}
	if err := client.sendRequest(request, &v); err != nil {
		return false, err
	}

	return true, nil
}

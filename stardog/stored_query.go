package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type StoredQueryService service

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

// getStoredQueriesResponse represents the response from the list stored queries endpoint
type getStoredQueriesResponse struct {
	StoredQueries *[]StoredQuery `json:"queries"`
}

// GetStoredQueries lists the stored queries that are accessible to the authenticated client
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Queries/operation/listStoredQueries
func (s *StoredQueryService) GetStoredQueries(ctx context.Context) (*[]StoredQuery, error) {

	url := fmt.Sprintf("%s/admin/queries/stored", s.client.BaseURL)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var storedQueriesResponse getStoredQueriesResponse
	if err := s.client.sendRequest(request, &storedQueriesResponse); err != nil {
		return nil, err
	}

	return storedQueriesResponse.StoredQueries, nil
}

// Add stored query, overwriting if a query with that name already exists
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Queries/operation/updateStoredQuery
func (s *StoredQueryService) CreateOrUpdateStoredQuery(ctx context.Context, sq StoredQuery) (bool, error) {
	url := fmt.Sprintf("%s/admin/queries/stored", s.client.BaseURL)

	requestBody, _ := json.Marshal(sq)
	payloadBuf := bytes.NewBuffer(requestBody)

	request, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}

	request.Header.Add("Content-type", "application/json")

	var v struct{}
	if err := s.client.sendRequest(request, &v); err != nil {
		return false, err
	}

	return true, nil
}

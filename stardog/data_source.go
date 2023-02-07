package stardog

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// DataSourceService handles communication with the [data source] related methods of the Stardog API.
//
// [data source]: https://docs.stardog.com/virtual-graphs/data-sources/
type DataSourceService service

type DataSource struct {
	// Name of data source
	Name string `json:"entityName"`
	// Whether the data source can be shared amongst virtual graphs
	Shareable bool `json:"sharable"`
	// Whether the data source is available or not
	Available bool `json:"available"`
}

// RefreshDataSourceOptions are optional parameters to the [DataSourceService.RefreshMetadata] method
type RefreshDataSourceOptions struct {
	// Optional table to refresh. Example formats (case-sensitive): catalog.schema.table, schema.table, table
	Table string `json:"name,omitempty"`
}

// response for ListNames
type listDataSourceNamesResponse struct {
	DataSources []string `json:"data_sources"`
}

// response for List
type listDataSourcesResponse struct {
	DataSources []DataSource `json:"data_sources"`
}

// response for Options
type dataSourceOptionsResponse struct {
	Options map[string]interface{} `json:"options"`
}

// request for Add
type addDataSourceRequest struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}

// request for Update
type updateDataSourceRequest struct {
	Options map[string]interface{} `json:"options"`
}

// ListNames returns the names of all data sources registered in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/listDataSources
func (s *DataSourceService) ListNames(ctx context.Context) ([]string, *Response, error) {
	u := "admin/data_sources"
	headerOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listDataSourcesResponse listDataSourceNamesResponse
	resp, err := s.client.Do(ctx, req, &listDataSourcesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listDataSourcesResponse.DataSources, resp, nil
}

// List returns the all DataSources registered in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/dataSourceInfos
func (s *DataSourceService) List(ctx context.Context) ([]DataSource, *Response, error) {
	u := "admin/data_sources/list"
	headerOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listDataSourcesResponse listDataSourcesResponse
	resp, err := s.client.Do(ctx, req, &listDataSourcesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listDataSourcesResponse.DataSources, resp, nil
}

// IsAvailable checks if a given data data source is available
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/availableDataSource
func (s *DataSourceService) IsAvailable(ctx context.Context, datasource string) (*bool, *Response, error) {
	u := fmt.Sprintf("admin/data_sources/%s/available", datasource)
	headerOpts := &requestHeaderOptions{
		Accept: mediaTypePlainText,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	resp, err := s.client.Do(ctx, req, &buf)
	if err != nil {
		return nil, resp, err
	}
	resultAsBool, err := strconv.ParseBool(buf.String())
	if err != nil {
		return nil, resp, err
	}
	return &resultAsBool, resp, err
}

// Options returns the all set options for the given data source
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/getDataSourceOptions
func (s *DataSourceService) Options(ctx context.Context, datasource string) (map[string]interface{}, *Response, error) {
	u := fmt.Sprintf("admin/data_sources/%s/options", datasource)
	headerOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var dataSourceOptionsResponse dataSourceOptionsResponse
	resp, err := s.client.Do(ctx, req, &dataSourceOptionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return dataSourceOptionsResponse.Options, resp, nil
}

// Add adds a new data source to the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/addDataSource
func (s *DataSourceService) Add(ctx context.Context, name string, opts map[string]interface{}) (*Response, error) {
	u := "admin/data_sources"
	headerOpts := &requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := &addDataSourceRequest{
		Name:    name,
		Options: opts,
	}
	req, err := s.client.NewRequest(http.MethodPost, u, headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Update updates an existing data source.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/updateDataSource
func (s *DataSourceService) Update(ctx context.Context, datasource string, opts map[string]interface{}) (*Response, error) {
	u := fmt.Sprintf("admin/data_sources/%s", datasource)
	headerOpts := &requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := &updateDataSourceRequest{
		Options: opts,
	}
	req, err := s.client.NewRequest(http.MethodPut, u, headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// RefreshMetadata clear the saved metadata for a
// Data Source and reload all its dependent Virtual Graphs with fresh metadata.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Data-Sources/operation/refreshMetadata
func (s *DataSourceService) RefreshMetadata(ctx context.Context, datasource string, opts *RefreshDataSourceOptions) (*Response, error) {
	u := fmt.Sprintf("admin/data_sources/%s/refresh_metadata", datasource)
	headerOpts := &requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}

  // Stardog expect to be sent at a minimum an empty JSON object if 
  // no table is specified in the opts
	var body interface{} = make(map[string]interface{})
	if opts != nil {
		body = opts
	} 
	req, err := s.client.NewRequest(http.MethodPost, u, headerOpts, body)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

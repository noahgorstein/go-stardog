package stardog

import (
	"context"
	"net/http"
)

// VirtualGraphService handles communication with the [virtual graph] related methods of the Stardog API.
//
// [virtual graph]: https://docs.stardog.com/virtual-graphs/
type VirtualGraphService service

type VirtualGraph struct {
	// name of the virtual graph (virtual:// prefix omitted)
	Name string `json:"name"`
	// the datasource associated with the virtual graph
	DataSource string `json:"data_source"`
	// the database associated with the virtual graph
	Database string `json:"database"`
	// whether this virtual graph is available to query
	Available bool `json:"available"`
}

type listVirtualGraphNamesResponse struct {
	VirtualGraphs []string `json:"virtual_graphs"`
}

type listVirtualGraphsResponse struct {
	VirtualGraphs []VirtualGraph `json:"virtual_graphs"`
}

// ListNames will return the names of all virtual graphs registered in the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Virtual-Graphs/operation/listVGs
func (s *VirtualGraphService) ListNames(ctx context.Context) ([]string, *Response, error) {
	u := "admin/virtual_graphs"
	requestHeaderOptions := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, requestHeaderOptions, nil)
	if err != nil {
		return nil, nil, err
	}
	var listVirtualGraphNamesResponse listVirtualGraphNamesResponse
	resp, err := s.client.Do(ctx, req, &listVirtualGraphNamesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listVirtualGraphNamesResponse.VirtualGraphs, resp, nil
}

// List will return the VirtualGraph(s) registered in the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Virtual-Graphs/operation/virtualGraphInfos
func (s *VirtualGraphService) List(ctx context.Context) ([]VirtualGraph, *Response, error) {
	u := "admin/virtual_graphs/list"
	requestHeaderOptions := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, requestHeaderOptions, nil)
	if err != nil {
		return nil, nil, err
	}
	var listVirtualGraphsResponse listVirtualGraphsResponse
	resp, err := s.client.Do(ctx, req, &listVirtualGraphsResponse)
	if err != nil {
		return nil, resp, err
	}
	return listVirtualGraphsResponse.VirtualGraphs, resp, nil
}

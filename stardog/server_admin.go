package stardog

import (
	"context"
	"fmt"
	"net/http"
)

// ServerAdminService provides access to the server admin related functions in the Stardog API.
type ServerAdminService service

// ProcessProgress represents a Process's progress
type ProcessProgress struct {
	Max     int    `json:"max"`
	Current int    `json:"current"`
	Stage   string `json:"stage"`
}

// Process represent a Stardog server process
type Process struct {
	Type      string          `json:"type"`
	KernelID  string          `json:"kernelId"`
	ID        string          `json:"id"`
	Db        string          `json:"db"`
	User      string          `json:"user"`
	StartTime int64           `json:"startTime"`
	Status    string          `json:"status"`
	Progress  ProcessProgress `json:"progress"`
}

// IsAlive returns whether the server is accepting traffic or not.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Server-Admin/operation/aliveCheck
func (s *ServerAdminService) IsAlive(ctx context.Context) (*bool, *Response, error) {
	url := "admin/alive"
	request, err := s.client.NewRequest(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.Do(ctx, request, nil)
	isAlive, err := parseBoolResponse(err)
	return &isAlive, resp, err
}

// GetProcesses returns all server processes.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Monitoring/operation/listProcesses
func (s *ServerAdminService) GetProcesses(ctx context.Context) (*[]Process, *Response, error) {
	url := "admin/processes"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getProcessesResponse []Process
	resp, err := s.client.Do(ctx, request, &getProcessesResponse)
	if err != nil {
		return nil, resp, err
	}

	return &getProcessesResponse, resp, nil
}

// GetProcess returns details for a server process.
func (s *ServerAdminService) GetProcess(ctx context.Context, processID string) (*Process, *Response, error) {
	url := fmt.Sprintf("admin/processes/%s", processID)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var ps Process
	resp, err := s.client.Do(ctx, request, &ps)
	if err != nil {
		return nil, resp, err
	}

	return &ps, resp, err
}

// KillProcess kills a server process.
func (s *ServerAdminService) KillProcess(ctx context.Context, processID string) (*Response, error) {
	url := fmt.Sprintf("admin/processes/%s", processID)
	request, err := s.client.NewRequest(http.MethodDelete, url, nil, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, request, nil)
}

package stardog

import (
	"context"
	"fmt"
)

type ServerAdminService service

// Represents a Stardog server process' progress
type ProcessProgress struct {
	Max     int    `json:"max"`
	Current int    `json:"current"`
	Stage   string `json:"stage"`
}

// Represents a Stardog server process
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

// Determine whether the Stardog server is running
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Server-Admin/operation/aliveCheck
func (s *ServerAdminService) IsAlive(ctx context.Context) (*bool, *Response, error) {
	url := "admin/alive"
	request, err := s.client.NewRequest("GET", url, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.Do(ctx, request, nil)
	isAlive, err := parseBoolResponse(err)
	return &isAlive, resp, err
}

// Get all server processes
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Monitoring/operation/listProcesses
func (s *ServerAdminService) GetProcesses(ctx context.Context) (*[]Process, *Response, error) {
	url := "admin/processes"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", url, &headerOpts, nil)
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

// Get details for a given process
func (s *ServerAdminService) GetProcess(ctx context.Context, processID string) (*Process, *Response, error) {
	url := fmt.Sprintf("admin/processes/%s", processID)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", url, &headerOpts, nil)
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

// Kill a given server process
func (s *ServerAdminService) KillProcess(ctx context.Context, processID string) (*Response, error) {
	url := fmt.Sprintf("admin/processes/%s", processID)
	request, err := s.client.NewRequest("DELETE", url, nil, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, request, nil)
}

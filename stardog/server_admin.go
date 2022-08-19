package stardog

import (
	"context"
	"fmt"
	"net/http"
)

type ServerAdminService service

// Determine whether the Stardog server is running
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Server-Admin/operation/aliveCheck
func (s *ServerAdminService) Alive(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/admin/alive", s.client.BaseURL)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var b struct{}
	if err := s.client.sendRequest(request, &b); err != nil {
		return false, err
	}

	return true, nil
}

type Process struct {
	Type      string `json:"type"`
	KernelID  string `json:"kernelId"`
	ID        string `json:"id"`
	Db        string `json:"db"`
	User      string `json:"user"`
	StartTime int64  `json:"startTime"`
	Status    string `json:"status"`
	Progress  struct {
		Max     int    `json:"max"`
		Current int    `json:"current"`
		Stage   string `json:"stage"`
	} `json:"progress"`
}

type Processes []Process

// Get all server processes
func (s *ServerAdminService) GetProcesses(ctx context.Context) (Processes, error) {

	url := fmt.Sprintf("%s/admin/processes", s.client.BaseURL)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	var pList Processes
	if err := s.client.sendRequest(request, &pList); err != nil {
		return nil, err
	}
	return pList, nil
}

// Get details for a given process
func (s *ServerAdminService) GetProcess(ctx context.Context, processId string) (*Process, error) {
	url := fmt.Sprintf("%s/admin/processes/%s", s.client.BaseURL, processId)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	var ps Process
	if err := s.client.sendRequest(request, &ps); err != nil {
		return nil, err
	}
	return &ps, nil
}

// Kill a given server process
func (s *ServerAdminService) KillProcess(ctx context.Context, processId string) (bool, error) {
	url := fmt.Sprintf("%s/admin/processes/%s", s.client.BaseURL, processId)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

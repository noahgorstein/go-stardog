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

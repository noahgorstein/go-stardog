package stardog

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// TransactionService provides access to the transaction related functions in the Stardog API.
type TransactionService service

// Begin creates a transaction. The transaction ID returned can be passed into other methods/functions
// that accept a transaction ID.
//
// Stardog API docs: https://stardog-union.github.io/http-docs/#tag/Transactions/operation/beginTransaction
func (s *TransactionService) Begin(ctx context.Context, database string) (string, *Response, error) {
	u := fmt.Sprintf("%s/transaction/begin", database)
  headerOpts := requestHeaderOptions{
    Accept: mediaTypePlainText,
  }
	req, err := s.client.NewRequest(http.MethodPost, u, &headerOpts, nil)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	resp, err := s.client.Do(ctx, req, &buf)
	if err != nil {
		return "", resp, err
	}

	return buf.String(), resp, nil
}

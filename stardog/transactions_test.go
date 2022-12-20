package stardog

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Begin(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	transactionUUID := "43FD6C7B-EE53-4618-A90D-7E45ADD8B433"
	database := "myDatabase"

	mux.HandleFunc(fmt.Sprintf("/%s/transaction/begin", database), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(transactionUUID))
	})

	ctx := context.Background()
	got, _, err := client.Transaction.Begin(ctx, database)
	if err != nil {
		t.Errorf("Transaction.Begin returned error: %v", err)
	}
	if want := transactionUUID; !cmp.Equal(got, want) {
		t.Errorf("Transaction.Begin = %+v, want %+v", got, want)
	}

	const methodName = "Begin"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Transaction.Begin(nil, database)
		if got == "" && err == nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want \"\"", methodName, got)
		}
		return resp, err
	})
}

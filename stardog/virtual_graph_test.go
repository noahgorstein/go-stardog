package stardog

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"testing"
)

func TestVirtualGraphService_ListNames(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var vgNamesJSON = []byte(`{
    "virtual_graphs": ["virtual://graph1", "virtual://graph2"] 
  }`)
	var wantVgNames = []string{"virtual://graph1", "virtual://graph2"}

	mux.HandleFunc("/admin/virtual_graphs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(vgNamesJSON)
	})

	ctx := context.Background()
	got, _, err := client.Virtual.ListNames(ctx)
	if err != nil {
		t.Errorf("Virtual.ListNames returned error: %v", err)
	}
	if want := wantVgNames; !cmp.Equal(got, want) {
		t.Errorf("Virtual.ListNames = %+v, want %+v", got, want)
	}

	const methodName = "Virtual.ListNames"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Virtual.ListNames(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestVirtualGraphService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var vgsJSON = []byte(`
    {
      "virtual_graphs": [
        {
          "name": "databricks",
          "data_source": "data-source://bricks",
          "database": "*",
          "available": true
        },
        {
          "name": "oracle",
          "data_source": "data-source://oracle",
          "database": "*",
          "available": true
        }
      ]
    }
  `)
	var wantVgs = []VirtualGraph{
		{
			Name:       "databricks",
			DataSource: "data-source://bricks",
			Database:   "*",
			Available:  true,
		},
		{
			Name:       "oracle",
			DataSource: "data-source://oracle",
			Database:   "*",
			Available:  true,
		},
	}

	mux.HandleFunc("/admin/virtual_graphs/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(vgsJSON)
	})

	ctx := context.Background()
	got, _, err := client.Virtual.List(ctx)
	if err != nil {
		t.Errorf("Virtual.List returned error: %v", err)
	}
	if want := wantVgs; !cmp.Equal(got, want) {
		t.Errorf("Virtual.List = %+v, want %+v", got, want)
	}

	const methodName = "Virtual.List"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Virtual.List(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

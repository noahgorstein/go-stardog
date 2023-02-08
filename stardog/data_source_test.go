package stardog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDataSourceService_ListNames(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var dsNamesJSON = []byte(`{
    "data_sources": ["postgres", "mysql"] 
  }`)
	var wantDsNames = []string{"postgres", "mysql"}

	mux.HandleFunc("/admin/data_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(dsNamesJSON)
	})

	ctx := context.Background()
	got, _, err := client.DataSource.ListNames(ctx)
	if err != nil {
		t.Errorf("DataSource.ListNames returned error: %v", err)
	}
	if want := wantDsNames; !cmp.Equal(got, want) {
		t.Errorf("DataSource.ListNames = %+v, want %+v", got, want)
	}

	const methodName = "DataSource.ListNames"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DataSource.ListNames(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDataSourceService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var vgNamesJSON = []byte(`
    {
      "data_sources": [
        {
          "entityName": "data-source://postgres",
          "sharable": true,
          "available": true
        }
      ]
    }
    `)
	var wantVgNames = []DataSource{
		{
			Name:      "data-source://postgres",
			Shareable: true,
			Available: true,
		},
	}

	mux.HandleFunc("/admin/data_sources/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(vgNamesJSON)
	})

	ctx := context.Background()
	got, _, err := client.DataSource.List(ctx)
	if err != nil {
		t.Errorf("DataSource.List returned error: %v", err)
	}
	if want := wantVgNames; !cmp.Equal(got, want) {
		t.Errorf("DataSource.List = %+v, want %+v", got, want)
	}

	const methodName = "DataSource.List"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DataSource.List(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDataSourceService_Available(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	responseString := "true"
	want := newTrue()

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/available", dsName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseString))
	})

	ctx := context.Background()
	got, _, err := client.DataSource.IsAvailable(ctx, dsName)
	if err != nil {
		t.Errorf("DataSource.IsAvailable returned error: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("DataSource.IsAvailable = %+v, want %+v", got, want)
	}

	const methodName = "IsAvailable"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DataSource.IsAvailable(nil, dsName)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDataSourceService_IsAvailable_nonIntegerResponse(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"
	responseString := "not an boolean"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/available", dsName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseString))
	})

	ctx := context.Background()
	_, _, err := client.DataSource.IsAvailable(ctx, dsName)
	if err == nil {
		t.Fatalf("DataSource.IsAvailable should return an error if response cannot be converted to an integer")
	}
}

func TestDataSourceService_Options(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var optionsJSON = []byte(`
    {
      "options": {
        "query.translation": "DEFAULT",
        "jdbc.url": "jdbc:postgresql://localhost:5432/employees",
        "jdbc.driver": "org.postgresql.Driver"
      }
    }
    `)
	var optionsMap = map[string]interface{}{
		"query.translation": "DEFAULT",
		"jdbc.url":          "jdbc:postgresql://localhost:5432/employees",
		"jdbc.driver":       "org.postgresql.Driver",
	}
	ds := "postgres"
	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/options", ds), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(optionsJSON)
	})

	ctx := context.Background()
	got, _, err := client.DataSource.Options(ctx, ds)
	if err != nil {
		t.Errorf("DataSource.Options returned error: %v", err)
	}
	if want := optionsMap; !cmp.Equal(got, want) {
		t.Errorf("DataSource.Options = %+v, want %+v", got, want)
	}

	const methodName = "DataSource.Options"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DataSource.Options(nil, ds)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDataSourceService_Add(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"
	dsOpts := map[string]interface{}{
		"jdbc.url":    "jdbc:postgresql://localhost:5432/employees",
		"jdbc.driver": "org.postgresql.Driver",
	}

	mux.HandleFunc("/admin/data_sources", func(w http.ResponseWriter, r *http.Request) {
		v := new(addDataSourceRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)

		want := &addDataSourceRequest{Name: dsName, Options: dsOpts}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.DataSource.Add(ctx, dsName, dsOpts)
	if err != nil {
		t.Errorf("DataSource.Add returned error: %v", err)
	}

	const methodName = "Add"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.Add(nil, dsName, dsOpts)
	})
}

func TestDataSourceService_Update(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"
	dsOpts := map[string]interface{}{
		"jdbc.url":    "jdbc:postgresql://localhost:5432/employees",
		"jdbc.driver": "org.postgresql.Driver",
	}

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s", dsName), func(w http.ResponseWriter, r *http.Request) {
		v := new(updateDataSourceRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)

		want := &updateDataSourceRequest{Options: dsOpts}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.DataSource.Update(ctx, dsName, dsOpts)
	if err != nil {
		t.Errorf("DataSource.Update returned error: %v", err)
	}

	const methodName = "Update"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.Update(nil, dsName, dsOpts)
	})
}

func TestDataSourceService_RefreshMetadata(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/refresh_metadata", dsName), func(w http.ResponseWriter, r *http.Request) {
		v := new(RefreshDataSourceMetadataOptions)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)

		want := &RefreshDataSourceMetadataOptions{Table: "people"}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	opts := &RefreshDataSourceMetadataOptions{
		Table: "people",
	}
	ctx := context.Background()
	_, err := client.DataSource.RefreshMetadata(ctx, dsName, opts)
	if err != nil {
		t.Errorf("DataSource.RefreshMetadata returned error: %v", err)
	}

	const methodName = "RefreshMetadata"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.RefreshMetadata(nil, dsName, opts)
	})
}

func TestDataSourceService_RefreshMetadata_noOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/refresh_metadata", dsName), func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := io.ReadAll(r.Body)
		json := string(bytes)
		emptyObj := "{}"
		if strings.Compare(strings.TrimSpace(json), emptyObj) != 0 {
			t.Errorf("if no options, req body should be an empty JSON object: %s", json)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	_, err := client.DataSource.RefreshMetadata(ctx, dsName, nil)
	if err != nil {
		t.Errorf("DataSource.RefreshMetadata returned error: %v", err)
	}
}

func TestDataSourceService_RefreshCounts(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/refresh_counts", dsName), func(w http.ResponseWriter, r *http.Request) {
		v := new(RefreshDataSourceCountsOptions)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)

		want := &RefreshDataSourceCountsOptions{Table: "people"}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	opts := &RefreshDataSourceCountsOptions{
		Table: "people",
	}
	ctx := context.Background()
	_, err := client.DataSource.RefreshCounts(ctx, dsName, opts)
	if err != nil {
		t.Errorf("DataSource.RefreshCounts returned error: %v", err)
	}

	const methodName = "RefreshCounts"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.RefreshCounts(nil, dsName, opts)
	})
}

func TestDataSourceService_RefreshCounts_noOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/refresh_counts", dsName), func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := io.ReadAll(r.Body)
		json := string(bytes)
		emptyObj := "{}"
		if strings.Compare(strings.TrimSpace(json), emptyObj) != 0 {
			t.Errorf("if no options, req body should be an empty JSON object: %s", json)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	_, err := client.DataSource.RefreshCounts(ctx, dsName, nil)
	if err != nil {
		t.Errorf("DataSource.RefreshCounts returned error: %v", err)
	}
}

func TestDataSourceService_Share(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/share", dsName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	_, err := client.DataSource.Share(ctx, dsName)
	if err != nil {
		t.Errorf("DataSource.Share returned error: %v", err)
	}

	const methodName = "Share"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.Share(nil, dsName)
	})
}

func TestDataSourceService_Online(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s/online", dsName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	_, err := client.DataSource.Online(ctx, dsName)
	if err != nil {
		t.Errorf("DataSource.Online returned error: %v", err)
	}

	const methodName = "Online"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.Online(nil, dsName)
	})
}

func TestDataSourceService_Delete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dsName := "postgres"

	mux.HandleFunc(fmt.Sprintf("/admin/data_sources/%s", dsName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	opts := &DeleteDataSourceOptions{
		Force: true,
	}
	_, err := client.DataSource.Delete(ctx, dsName, opts)
	if err != nil {
		t.Errorf("DataSource.Delete returned error: %v", err)
	}

	const methodName = "Delete"
	testBadOptions(t, methodName, func() (err error) {
		_, err = client.DataSource.Delete(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DataSource.Delete(nil, dsName, opts)
	})
}

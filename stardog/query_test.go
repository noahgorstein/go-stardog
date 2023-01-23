package stardog

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"testing"
)

func TestQueryResultFormat_Valid(t *testing.T) {
	f := QueryResultFormat(100)
	if f.Valid() {
		t.Errorf("should be an invalid QueryResultFormat")
	}
	if f.String() != QueryResultFormatUnknown.String() {
		t.Errorf("QueryResultFormat string value should be unknown")
	}
}

func TestQueryPlanFormat_Valid(t *testing.T) {
	f := QueryPlanFormat(100)
	if f.Valid() {
		t.Errorf("should be an invalid QueryPlanFormat")
	}
	if f.String() != QueryPlanFormatUnknown.String() {
		t.Errorf("QueryPlanFormat string value should be unknown")
	}
}

func TestSparqlService_Select(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantQueryResults := `
  s,o
  http://stardog.com/tutorial/The_Beatles,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/The_Beatles,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/Metallica,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/Genesis_(band),http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/The_Rolling_Stones,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/The_Beach_Boys,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/Van_Halen,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/Alabama_(band),http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/U2,http://stardog.com/tutorial/Band
  http://stardog.com/tutorial/Foreigner_(band),http://stardog.com/tutorial/Band
  `

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/query", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeTextCSV)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantQueryResults))
	})

	ctx := context.Background()
	query := `
  SELECT * { ?s a ?o }
  `

	queryOpts := &SelectOptions{
		ResultFormat: QueryResultFormatCSV,
		Limit:        10,
	}

	got, _, err := client.Sparql.Select(ctx, db, query, queryOpts)
	if err != nil {
		t.Errorf("Sparql.Select returned error: %v", err)
	}
	if want := wantQueryResults; !cmp.Equal(got.String(), want) {
		t.Errorf("Sparql.Select = %+v, want %+v", got, want)
	}

	const methodName = "Select"
	testBadOptions(t, methodName, func() (err error) {
		_, _, err = client.Sparql.Select(ctx, "\n", "\n", queryOpts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Sparql.Select(nil, db, query, nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestSparqlService_Select_noReturnFormatSpecified(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantQueryResults := `
    {
      "head": {
        "vars": [
          "s",
          "o"
        ]
      },
      "results": {
        "bindings": [
          {
            "s": {
              "type": "uri",
              "value": "http://stardog.com/tutorial/The_Beatles"
            },
            "o": {
              "type": "uri",
              "value": "http://stardog.com/tutorial/Band"
            }
          }
        ]
      }
    }
  `

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/query", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationSparqlResultsJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantQueryResults))
	})

	ctx := context.Background()
	query := `
  SELECT * { ?s a ?o }
  `

	got, _, err := client.Sparql.Select(ctx, db, query, nil)
	if err != nil {
		t.Errorf("Sparql.Select returned error: %v", err)
	}
	if want := wantQueryResults; !cmp.Equal(got.String(), want) {
		t.Errorf("Sparql.Select = %+v, want %+v", got, want)
	}
}

func TestSparqlService_Explain(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantPlan := `
  {"prefixes":{},"dataset":{},"plan":{"children":[{"children":[],"label":"Scan[POSC](?s, rdf:type, ?o)","cardinality":1}],"label":"Projection(?s, ?o)","cardinality":1}}
  `
	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/explain", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantPlan))
	})

	ctx := context.Background()
	query := `
  SELECT * { ?s a ?o }
  `

	explainOpts := &ExplainOptions{
		QueryPlanFormat: QueryPlanFormatJSON,
	}

	got, _, err := client.Sparql.Explain(ctx, db, query, explainOpts)
	if err != nil {
		t.Errorf("Sparql.Explain returned error: %v", err)
	}
	if want := wantPlan; !cmp.Equal(got.String(), want) {
		t.Errorf("Sparql.Explain = %+v, want %+v", got, want)
	}

	const methodName = "Explain"
	testBadOptions(t, methodName, func() (err error) {
		_, _, err = client.Sparql.Explain(ctx, "\n", "\n", explainOpts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Sparql.Explain(nil, db, query, nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestSparqlService_Explain_noPlanFormatSpecified(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/explain", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	query := `
  SELECT * { ?s a ?o }
  `
	_, _, err := client.Sparql.Explain(ctx, db, query, nil)
	if err != nil {
		t.Errorf("Sparql.Explain returned error: %v", err)
	}
}

func TestSparqlService_Update(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"
	mux.HandleFunc(fmt.Sprintf("/%s/update", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	query := `
  INSERT DATA { GRAPH <urn:data:graph> { <foo:a> a <foo:b> } }
  `

	updateOpts := &UpdateOptions{
		DefaultGraphURI: "tag:stardog:api:context:default",
	}

	_, err := client.Sparql.Update(ctx, db, query, updateOpts)
	if err != nil {
		t.Errorf("Sparql.Update returned error: %v", err)
	}

	const methodName = "Update"
	testBadOptions(t, methodName, func() (err error) {
		_, err = client.Sparql.Update(ctx, "\n", "\n", updateOpts)
		return err
	})

	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Sparql.Update(nil, db, query, nil)
	})
}

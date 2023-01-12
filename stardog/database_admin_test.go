package stardog

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_CreateDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/admin/databases"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)

		if r.PostFormValue("root") == "" {
			t.Errorf("DatabaseAdmin.CreateDatabase should have a key with the name 'root' in the POST'd form")
		}
	})

	dbName := "db1"
	dbOpts := map[string]interface{}{
		"spatial.enabled": true,
	}
	datasetsWithRealFilePaths := []Dataset{
		{
			Path:    "./test-resources/beatles.ttl",
			Context: "http://beatles",
		},
		{
			Path:    "./test-resources/music_schema.ttl",
			Context: "http://schema",
		},
	}
	datasetsWithFakeFilePaths := []Dataset{
		{
			Path:    "./fake-directory/beatles.ttl",
			Context: "http://beatles",
		},
		{
			Path:    "./fake-directory/music_schema.ttl",
			Context: "http://schema",
		},
	}

	ctx := context.Background()
	_, err := client.DatabaseAdmin.CreateDatabase(ctx, dbName, datasetsWithRealFilePaths, dbOpts, true)
	if err != nil {
		t.Errorf("DatabaseAdmin.CreateDatabase returned error: %v", err)
	}

	_, err = client.DatabaseAdmin.CreateDatabase(ctx, dbName, datasetsWithFakeFilePaths, dbOpts, true)
	if err == nil {
		t.Error("DatabaseAdmin.CreateDatabase should return an error due to not being able to find the files.")
	}

	const methodName = "CreateDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.CreateDatabase(nil, dbName, datasetsWithRealFilePaths, dbOpts, true)
	})

}

func Test_RestoreDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	pathToBackup := "/path/to/backup"
	restoreDatabaseOptions := &RestoreDatabaseOptions{
		Force: true,
		Name:  "restoredDatabaseName",
	}

	mux.HandleFunc(fmt.Sprintf("/admin/restore"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.RestoreDatabase(ctx, pathToBackup, restoreDatabaseOptions)
	if err != nil {
		t.Errorf("DatabaseAdmin.RestoreDatabase returned error: %v", err)
	}

	const methodName = "RestoreDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.RestoreDatabase(nil, pathToBackup, restoreDatabaseOptions)
	})
}

func Test_RepairDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/repair", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.RepairDatabase(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.RepairDatabase returned error: %v", err)
	}

	const methodName = "RepairDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.RepairDatabase(nil, db)
	})
}

func Test_OptimizeDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/optimize", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.OptimizeDatabase(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.OptimizeDatabase returned error: %v", err)
	}

	const methodName = "OptimizeDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.OptimizeDatabase(nil, db)
	})
}

func Test_DropDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.DropDatabase(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.DropDatabase returned error: %v", err)
	}

	const methodName = "DropDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.DropDatabase(nil, db)
	})
}

func Test_GetAllDatabaseOptionsDetails(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var optionsJSON = []byte(`
    {
      "auto.schema.reasoning": {
        "name": "auto.schema.reasoning",
        "type": "Boolean",
        "server": false,
        "mutable": true,
        "mutableWhenOnline": true,
        "category": "Reasoning",
        "label": "Auto Schema Reasoning",
        "description": "Enables reasoning when automatically generating schemas from OWL. This setting will affect automatic schema generation for GraphQL (if graphql.auto.schema is enabled) and BI/SQL (if sql.schema.auto is enabled).",
        "defaultValue": true
      },
      "database.archetypes": {
        "name": "database.archetypes",
        "type": "String",
        "server": false,
        "mutable": true,
        "mutableWhenOnline": false,
        "category": "Database",
        "label": "Database Archetypes",
        "description": "The name of one or more database archetypes, used to associate ontologies and constraints with new databases. See the docs for instructions to create your own archetype.",
        "defaultValue": []
      }
    }
    `)
	var databaseOptions = map[string]DatabaseOptionDetails{
		"auto.schema.reasoning": {
			Name:              "auto.schema.reasoning",
			Type:              "Boolean",
			Server:            false,
			Mutable:           true,
			MutableWhenOnline: true,
			Category:          "Reasoning",
			Label:             "Auto Schema Reasoning",
			Description:       "Enables reasoning when automatically generating schemas from OWL. This setting will affect automatic schema generation for GraphQL (if graphql.auto.schema is enabled) and BI/SQL (if sql.schema.auto is enabled).",
			DefaultValue:      true,
		},
		"database.archetypes": {
			Name:              "database.archetypes",
			Type:              "String",
			Server:            false,
			Mutable:           true,
			MutableWhenOnline: false,
			Category:          "Database",
			Label:             "Database Archetypes",
			Description:       "The name of one or more database archetypes, used to associate ontologies and constraints with new databases. See the docs for instructions to create your own archetype.",
			DefaultValue:      []interface{}{},
		},
	}

	mux.HandleFunc("/admin/config_properties", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(optionsJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetAllDatabaseOptionDetails(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetAllDatabaseOptionDetails returned error: %v", err)
	}
	if want := databaseOptions; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.GetAllDatabaseOptionDetails = %+v, want %+v", got, want)
	}

	const methodName = "GetAllDatabaseOptionDetails"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetAllDatabaseOptionDetails(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetDatabases(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var databasesJSON = []byte(`{"databases": ["db1", "db2"]}`)
	var wantDatabases = []string{"db1", "db2"}

	mux.HandleFunc("/admin/databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(databasesJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetDatabases(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetDatabases returned error: %v", err)
	}
	if want := wantDatabases; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.GetDatabases = %+v, want %+v", got, want)
	}

	const methodName = "GetDatabases"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetDatabases(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetDatabaseWithOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var databasesWithOptionsJSON = []byte(`
  {
    "optimize.vacuum.data": true,
    "index.dictionary.compress.literal": 2048,
    "search.version": 3,
    "security.masking.function": "",
    "index.persist": true,
    "search.index.contexts.excluded": true,
    "index.lucene.mmap": true,
    "preserve.bnode.ids": true,
    "index.aggregate": "Off",
    "service.sparql.result.limit": 1000,
    "index.type": "Disk",
    "index.statistics.sketch.capacity": 100000000,
    "search.index.contexts.filter": [],
    "index.strategy": "NO_AGGREGATE_INDEXES",
    "spatial.index.dirty": true,
    "sql.schema.graph": "tag:stardog:api:sql:schema",
    "graph.aliases": false,
    "docs.path": "docs",
    "database.archetypes": [],
    "reasoning.schema.timeout": "1m",
    "transaction.logging.rotation.remove": false,
    "transaction.logging": false,
    "edge.properties": false,
    "reasoning.consistency.automatic": false,
    "reasoning.virtual.graph.enabled": true,
    "auto.schema.reasoning": true,
    "index.disk.page.count.used": 0,
    "progress.monitor.enabled": true,
    "metrics.native.reportingInterval": "10s",
    "index.literals.merge.limit": 100000,
    "security.properties.sensitive.groups": [],
    "reasoning.schemas": [],
    "reasoning.schemas.memory.count": 5,
    "index.last.tx": "0329f741-228f-460d-88b3-1e78065e3793"
  }`)
	var wantDatabases = map[string]interface{}{
		"optimize.vacuum.data":                 true,
		"index.dictionary.compress.literal":    2048,
		"search.version":                       3,
		"security.masking.function":            "",
		"index.persist":                        true,
		"search.index.contexts.excluded":       true,
		"index.lucene.mmap":                    true,
		"preserve.bnode.ids":                   true,
		"index.aggregate":                      "Off",
		"service.sparql.result.limit":          1000,
		"index.type":                           "Disk",
		"index.statistics.sketch.capacity":     1e+08,
		"search.index.contexts.filter":         []string{},
		"index.strategy":                       "NO_AGGREGATE_INDEXES",
		"spatial.index.dirty":                  true,
		"sql.schema.graph":                     "tag:stardog:api:sql:schema",
		"graph.aliases":                        false,
		"docs.path":                            "docs",
		"database.archetypes":                  []string{},
		"reasoning.schema.timeout":             "1m",
		"transaction.logging.rotation.remove":  false,
		"transaction.logging":                  false,
		"edge.properties":                      false,
		"reasoning.consistency.automatic":      false,
		"reasoning.virtual.graph.enabled":      true,
		"auto.schema.reasoning":                true,
		"index.disk.page.count.used":           0,
		"progress.monitor.enabled":             true,
		"metrics.native.reportingInterval":     "10s",
		"index.literals.merge.limit":           100000,
		"security.properties.sensitive.groups": []string{},
		"reasoning.schemas":                    []string{},
		"reasoning.schemas.memory.count":       5,
		"index.last.tx":                        "0329f741-228f-460d-88b3-1e78065e3793",
	}

	mux.HandleFunc("/admin/databases/db1/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(databasesWithOptionsJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetDatabaseWithOptions(ctx, "db1")
	if err != nil {
		t.Errorf("DatabaseAdmin.GetDatabaseWithOptions returned error: %v", err)
	}
	if want := wantDatabases; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.GetDatabaseWithOptions returned map with length %+v, want %+v", got, want)
	}

	const methodName = "GetDatabaseWithOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetDatabaseWithOptions(nil, "db1")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetDatabasesWithOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var databasesWithOptionsJSON = []byte(`
    {
    "databases": [
      {
        "optimize.vacuum.data": true,
        "index.dictionary.compress.literal": 2048,
        "search.version": 3,
        "security.masking.function": "",
        "index.persist": true,
        "search.index.contexts.excluded": true,
        "index.lucene.mmap": true,
        "preserve.bnode.ids": true,
        "index.aggregate": "Off",
        "service.sparql.result.limit": 1000,
        "index.type": "Disk",
        "index.statistics.sketch.capacity": 100000000,
        "search.index.contexts.filter": [],
        "index.strategy": "NO_AGGREGATE_INDEXES",
        "spatial.index.dirty": true,
        "sql.schema.graph": "tag:stardog:api:sql:schema",
        "graph.aliases": false,
        "docs.path": "docs",
        "database.archetypes": [],
        "reasoning.schema.timeout": "1m",
        "transaction.logging.rotation.remove": false,
        "transaction.logging": false,
        "edge.properties": false,
        "reasoning.consistency.automatic": false,
        "reasoning.virtual.graph.enabled": true,
        "auto.schema.reasoning": true,
        "index.disk.page.count.used": 0,
        "progress.monitor.enabled": true,
        "metrics.native.reportingInterval": "10s",
        "index.literals.merge.limit": 100000,
        "security.properties.sensitive.groups": [],
        "reasoning.schemas": [],
        "reasoning.schemas.memory.count": 5,
        "index.last.tx": "0329f741-228f-460d-88b3-1e78065e3793"
      }
    ]
    }
`)
	var wantDatabasesWithOptions = []map[string]interface{}{{
		"optimize.vacuum.data":                 true,
		"index.dictionary.compress.literal":    2048,
		"search.version":                       3,
		"security.masking.function":            "",
		"index.persist":                        true,
		"search.index.contexts.excluded":       true,
		"index.lucene.mmap":                    true,
		"preserve.bnode.ids":                   true,
		"index.aggregate":                      "Off",
		"service.sparql.result.limit":          1000,
		"index.type":                           "Disk",
		"index.statistics.sketch.capacity":     1e+08,
		"search.index.contexts.filter":         []string{},
		"index.strategy":                       "NO_AGGREGATE_INDEXES",
		"spatial.index.dirty":                  true,
		"sql.schema.graph":                     "tag:stardog:api:sql:schema",
		"graph.aliases":                        false,
		"docs.path":                            "docs",
		"database.archetypes":                  []string{},
		"reasoning.schema.timeout":             "1m",
		"transaction.logging.rotation.remove":  false,
		"transaction.logging":                  false,
		"edge.properties":                      false,
		"reasoning.consistency.automatic":      false,
		"reasoning.virtual.graph.enabled":      true,
		"auto.schema.reasoning":                true,
		"index.disk.page.count.used":           0,
		"progress.monitor.enabled":             true,
		"metrics.native.reportingInterval":     "10s",
		"index.literals.merge.limit":           100000,
		"security.properties.sensitive.groups": []string{},
		"reasoning.schemas":                    []string{},
		"reasoning.schemas.memory.count":       5,
		"index.last.tx":                        "0329f741-228f-460d-88b3-1e78065e3793",
	}}

	mux.HandleFunc("/admin/databases/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(databasesWithOptionsJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetDatabasesWithOptions(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetDatabasesWithOptions returned error: %v", err)
	}
	t.Log(len(got))
	if want := wantDatabasesWithOptions; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.GetDatabasesWithOptions returned slice has length %+v, want %+v", got, want)
	}

	const methodName = "GetDatabasesWithOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetDatabasesWithOptions(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestGetDatabaseSize(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	getDatabaseSizeOptions := &GetDatabaseSizeOptions{
		Exact: true,
	}
	dbName := "db1"

	responseString := "1000"
	want := newInt(1000)

	mux.HandleFunc(fmt.Sprintf("/%s/size", dbName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseString))
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetDatabaseSize(ctx, dbName, getDatabaseSizeOptions)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetDatabaseSize returned error: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.GetDatabaseSize = %+v, want %+v", got, want)
	}

	const methodName = "GetDatabaseSize"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetDatabaseSize(nil, dbName, getDatabaseSizeOptions)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

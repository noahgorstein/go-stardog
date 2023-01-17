package stardog

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ExportData_server_side(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()

	opts := &ExportDataOptions{
		NamedGraph:  []string{"tag:stardog:api:context:default"},
		Format:      Turtle,
		ServerSide:  true,
		Compression: BZ2,
	}

	_, _, err := client.DatabaseAdmin.ExportData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportData returned error: %v", err)
	}

	const methodName = "ExportData"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_ExportData_client_side(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"
	returnedRDF :=
		`
    PREFIX : <http://stardog.com/tutorial/>

    :The_Beatles      rdf:type  :Band .
    :The_Beatles      :name     "The Beatles" .
    :The_Beatles      :member   :John_Lennon .
    :The_Beatles      :member   :Paul_McCartney .
  `

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", string(Turtle))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnedRDF))
	})

	ctx := context.Background()

	opts := &ExportDataOptions{
		NamedGraph: []string{"tag:stardog:api:context:default"},
		Format:     Turtle,
	}

	got, _, err := client.DatabaseAdmin.ExportData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportData returned error: %v", err)
	}

	if want := returnedRDF; !cmp.Equal(got.String(), want) {
		t.Errorf("DatabaseAdmin.ExportData = %+v, want %+v", got, want)
	}

	const methodName = "ExportData"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_ExportObfuscatedData_client_side(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"
	returnedRDF :=
		`
    @prefix : <http://api.stardog.com/> .
    @prefix stardog: <tag:stardog:api:> .
    @prefix obf: <tag:stardog:api:obf:> .
    @prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
    @prefix owl: <http://www.w3.org/2002/07/owl#> .
    @prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
    @prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

    {
      obf:c61219651e7f0bf78ef1ab754768a6eb1bd9d53df39aa5ef153fcf55b4f12b1f "1971-10-11"^^xsd:date ;
      obf:9d57d43c21ec38ec19d5a782367aaa3f9ab92230068d71a68c73cc4b3c0670e9 obf:638a34f33c0ceb352ad944c901e924d64683bc99ea895ca0a9a8142bdecc72fe .
    }
  `

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", string(Trig))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnedRDF))
	})

	ctx := context.Background()

	opts := &ExportObfuscatedDataOptions{
		NamedGraph: []string{"tag:stardog:api:context:default"},
		Format:     Trig,
	}

	got, _, err := client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData returned error: %v", err)
	}

	if want := returnedRDF; !cmp.Equal(got.String(), want) {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData = %+v, want %+v", got, want)
	}

	const methodName = "ExportObfuscatedData"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportObfuscatedData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_ExportObfuscatedData_client_side_custom_obf_config(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"
	returnedRDF :=
		`
    @prefix : <http://api.stardog.com/> .
    @prefix stardog: <tag:stardog:api:> .
    @prefix obf: <tag:stardog:api:obf:> .
    @prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
    @prefix owl: <http://www.w3.org/2002/07/owl#> .
    @prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
    @prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

    {
      obf:c61219651e7f0bf78ef1ab754768a6eb1bd9d53df39aa5ef153fcf55b4f12b1f "1971-10-11"^^xsd:date ;
      obf:9d57d43c21ec38ec19d5a782367aaa3f9ab92230068d71a68c73cc4b3c0670e9 obf:638a34f33c0ceb352ad944c901e924d64683bc99ea895ca0a9a8142bdecc72fe .
    }
  `

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", string(Turtle))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnedRDF))
	})

	ctx := context.Background()

	config, err := os.Open("./test-resources/obfuscation-config.ttl")
	if err != nil {
		t.Errorf("error opening the obfuscation configuration file")
	}

	opts := &ExportObfuscatedDataOptions{
		NamedGraph:        []string{"tag:stardog:api:context:default"},
		Format:            Turtle,
		ObfuscationConfig: config,
	}

	got, _, err := client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData returned error: %v", err)
	}

	if want := returnedRDF; !cmp.Equal(got.String(), want) {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData = %+v, want %+v", got, want)
	}

	const methodName = "ExportObfuscatedData"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportObfuscatedData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_ExportObfuscatedData_server_side(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()

	opts := &ExportObfuscatedDataOptions{
		NamedGraph:  []string{"tag:stardog:api:context:default"},
		Format:      Turtle,
		ServerSide:  true,
		Compression: BZ2,
	}

	_, _, err := client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData returned error: %v", err)
	}

	const methodName = "ExportObfuscatedData"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportObfuscatedData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_OnlineDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/online", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.OnlineDatabase(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.OnlineDatabase returned error: %v", err)
	}

	const methodName = "OnlineDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.OnlineDatabase(nil, db)
	})
}

func Test_OfflineDatabase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/offline", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.OfflineDatabase(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.OfflineDatabase returned error: %v", err)
	}

	const methodName = "OfflineDatabase"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.OfflineDatabase(nil, db)
	})
}

func Test_GenerateDataModel(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"
	dataModel :=
		`
  PREFIX catalog: <tag:stardog:api:catalog:>
  PREFIX dcat: <http://www.w3.org/ns/dcat#>
  PREFIX r2rml: <http://www.w3.org/ns/r2rml#>

  # A database table column.
  Class catalog:Column extends dcat:Resource
          catalog:columnName xsd:string
          catalog:columnType xsd:string
  `

	mux.HandleFunc(fmt.Sprintf("/%s/model", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(dataModel))
	})

	ctx := context.Background()
	opts := &GenerateDataModelOptions{
		Reasoning: true,
		Output:    "text",
	}
	got, _, err := client.DatabaseAdmin.GenerateDataModel(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.GenerateDataModel returned error: %v", err)
	}
	if want := dataModel; !cmp.Equal(got.String(), want) {
		t.Errorf("DatabaseAdmin.GenerateDataModel = %+v, want %+v", got, want)
	}

	const methodName = "GenerateDataModel"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GenerateDataModel(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

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
			Path:       "./test-resources/beatles.ttl",
			NamedGraph: "http://beatles",
		},
		{
			Path:       "./test-resources/music_schema.ttl",
			NamedGraph: "http://schema",
		},
	}
	datasetsWithFakeFilePaths := []Dataset{
		{
			Path:       "./fake-directory/beatles.ttl",
			NamedGraph: "http://beatles",
		},
		{
			Path:       "./fake-directory/music_schema.ttl",
			NamedGraph: "http://schema",
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

	_, err = client.DatabaseAdmin.CreateDatabase(ctx, dbName, nil, nil, false)
	if err != nil {
		t.Errorf("DatabaseAdmin.CreateDatabase returned error: %v", err)
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

func Test_GetNamespaces(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var namespacesJSON = []byte(
		`
    { "namespaces": [
      {
        "prefix": "",
        "name": "http://stardog.com/tutorial/"
      },
      {
        "prefix": "schema",
        "name": "http://schema.org/"
      },
      {
        "prefix": "stardog",
        "name": "tag:stardog:api:"
      }
    ]}
    `)
	wantNamespaces := []Namespace{
		{
			Prefix: "",
			Name:   "http://stardog.com/tutorial/",
		},
		{
			Prefix: "schema",
			Name:   "http://schema.org/",
		},
		{
			Prefix: "stardog",
			Name:   "tag:stardog:api:",
		},
	}
	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/namespaces", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(namespacesJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetNamespaces(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetNamespaces returned error: %v", err)
	}
	if want := wantNamespaces; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.GetNamespaces = %+v, want %+v", got, want)
	}

	const methodName = "GetNamespaces"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetNamespaces(nil, db)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_ImportNamespaces(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	importNamespacesResponseJSON := `
  {
    "numImportedNamespaces": 1,
    "namespaces": [
      "\u003dhttp://stardog.com/tutorial/",
      "schema\u003dhttp://schema.org/",
      "stardog\u003dtag:stardog:api:"
    ]
  }
  `
	wantImportNamespacesResponse := &ImportNamespacesResponse{
		NumberImportedNamespaces: 1,
		UpdatedNamespaces: []string{
			"=http://stardog.com/tutorial/",
			"schema=http://schema.org/",
			"stardog=tag:stardog:api:",
		},
	}

	rdf, err := os.Open("./test-resources/music_schema.ttl")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux.HandleFunc(fmt.Sprintf("/%s/namespaces", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(importNamespacesResponseJSON))
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.ImportNamespaces(ctx, db, rdf)
	if err != nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces returned error: %v", err)
	}
	if want := wantImportNamespacesResponse; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.ImportNamespaces = %+v, want %+v", got, want)
	}

	const methodName = "ImportNamespaces"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ImportNamespaces(nil, db, rdf)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetDatabaseOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var databaseOptionsJSON = []byte(`{"search.enabled": true}`)
	var wantDatabasOptions = map[string]interface{}{"search.enabled": true}

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/options", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(databaseOptionsJSON)
	})

	ctx := context.Background()
	opts := []string{"search.enabled"}
	got, _, err := client.DatabaseAdmin.GetDatabaseOptions(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetDatabaseOptions returned error: %v", err)
	}
	if want := wantDatabasOptions; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.GetDatabaseOptions = %+v, want %+v", got, want)
	}

	const methodName = "GetDatabaseOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetDatabaseOptions(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_SetDatabaseOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/options", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)
		testBody(t, r, `{"search.enabled":true}`+"\n")
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	opts := map[string]interface{}{"search.enabled": true}
	_, err := client.DatabaseAdmin.SetDatabaseOptions(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.SetDatabaseOptions returned error: %v", err)
	}
	const methodName = "SetDatabaseOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.SetDatabaseOptions(nil, db, opts)
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

func Test_GetAllDatabaseOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var databaseOptionsJSON = []byte(`
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
	var wantDatabaseOptions = map[string]interface{}{
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
		w.Write(databaseOptionsJSON)
	})

	ctx := context.Background()
	got, _, err := client.DatabaseAdmin.GetAllDatabaseOptions(ctx, "db1")
	if err != nil {
		t.Errorf("DatabaseAdmin.GetAllDatabaseOptions returned error: %v", err)
	}
	if want := wantDatabaseOptions; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.GetAllDatabaseOptions returned map with length %+v, want %+v", got, want)
	}

	const methodName = "GetAllDatabaseOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetAllDatabaseOptions(nil, "db1")
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
	got, _, err := client.DatabaseAdmin.GetAllDatabasesWithOptions(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.GetAllDatabasesWithOptions returned error: %v", err)
	}
	t.Log(len(got))
	if want := wantDatabasesWithOptions; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.GetAllDatabasesWithOptions returned slice has length %+v, want %+v", got, want)
	}

	const methodName = "GetAllDatabasesWithOptions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.GetAllDatabasesWithOptions(nil)
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

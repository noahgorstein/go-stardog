package stardog

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDatabaseAdminService_DataModelFormat_Valid(t *testing.T) {
	f := DataModelFormat(100)
	if f.Valid() {
		t.Errorf("should be an invalid DataModelFormat")
	}
	if f.String() != DataModelFormatUnknown.String() {
		t.Errorf("DataModelFormat string value should be an empty string")
	}
}

func TestDatabaseAdminService_ExportData_serverSide(t *testing.T) {
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
		Format:      RDFFormatTurtle,
		ServerSide:  true,
		Compression: CompressionBZ2,
	}

	_, _, err := client.DatabaseAdmin.ExportData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportData returned error: %v", err)
	}

	const methodName = "ExportData"
	testBadOptions(t, methodName, func() (err error) {
		_, _, err = client.DatabaseAdmin.ExportData(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_ExportData_clientSide(t *testing.T) {
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
		testHeader(t, r, "Accept", RDFFormatTurtle.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnedRDF))
	})

	ctx := context.Background()

	opts := &ExportDataOptions{
		NamedGraph: []string{"tag:stardog:api:context:default"},
		Format:     RDFFormatTurtle,
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

func TestDatabaseAdminService_ExportObfuscatedData_client_side(t *testing.T) {
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
		testHeader(t, r, "Accept", RDFFormatTrig.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(returnedRDF))
	})

	ctx := context.Background()

	opts := &ExportObfuscatedDataOptions{
		NamedGraph: []string{"tag:stardog:api:context:default"},
		Format:     RDFFormatTrig,
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

func TestDatabaseAdminService_ExportObfuscatedData_clientSideCustomObfConfig(t *testing.T) {
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
		testHeader(t, r, "Accept", RDFFormatTurtle.String())
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
		Format:            RDFFormatTurtle,
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

	// close the obfuscation to force an error
	config.Close()
	_, _, err = client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err == nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData didn't return an error about passing a closed configuration file")
	}

	opts.ObfuscationConfig, err = os.Open("./test-resources")
	if err != nil {
		t.Errorf("unexpected error opening a directory")
	}
	_, _, err = client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err == nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData didn't return an error about passing a directory instead of a file")
	}
}

func TestDatabaseAdminService_ExportObfuscatedData_serverSide(t *testing.T) {
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
		Format:      RDFFormatTurtle,
		ServerSide:  true,
		Compression: CompressionBZ2,
	}

	_, _, err := client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData returned error: %v", err)
	}

	const methodName = "ExportObfuscatedData"
	testBadOptions(t, methodName, func() (err error) {
		_, _, err = client.DatabaseAdmin.ExportObfuscatedData(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ExportObfuscatedData(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_ExportObfuscatedData_serverSideConfig(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/%s/export", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypePlainText)

		if strings.Contains(r.RequestURI, "?obf=DEFAULT") {
			t.Errorf("request URI should not contain ?obf=DEFAULT if configuration file provided")
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()

	config, err := os.Open("./test-resources/obfuscation-config.ttl")
	if err != nil {
		t.Errorf("error opening the obfuscation configuration file")
	}
	defer config.Close()

	opts := &ExportObfuscatedDataOptions{
		Format:            RDFFormatTrig,
		ServerSide:        true,
		ObfuscationConfig: config,
	}
	_, _, err = client.DatabaseAdmin.ExportObfuscatedData(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.ExportObfuscatedData returned error: %v", err)
	}
}

func TestDatabaseAdminService_Online(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/online", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.Online(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Online returned error: %v", err)
	}

	const methodName = "Online"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Online(nil, db)
	})
}

func TestDatabaseAdminService_Offline(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/offline", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.Offline(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Offline returned error: %v", err)
	}

	const methodName = "Offline"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Offline(nil, db)
	})
}

func TestDatabaseAdminService_DataModel(t *testing.T) {
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
	opts := &DataModelOptions{
		Reasoning:    true,
		OutputFormat: DataModelFormatText,
	}
	got, _, err := client.DatabaseAdmin.DataModel(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.DataModel returned error: %v", err)
	}
	if want := dataModel; !cmp.Equal(got.String(), want) {
		t.Errorf("DatabaseAdmin.DataModel = %+v, want %+v", got, want)
	}

	const methodName = "DataModel"
	testBadOptions(t, methodName, func() (err error) {
		opts := &DataModelOptions{}
		_, _, err = client.DatabaseAdmin.DataModel(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.DataModel(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	respInfoJSON := `{"message":"Bulk loading data to new database movies.\nLoaded 41,099 triples to movies from 1 file(s) in 00:00:00.351 @ 117.1K triples/sec.\nSuccessfully created database 'movies'.\n"}`

	mux.HandleFunc(fmt.Sprintf("/admin/databases"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respInfoJSON))

		if r.PostFormValue("root") == "" {
			t.Errorf("DatabaseAdmin.Create should have a key with the name 'root' in the POST'd form")
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

	optsWithRealDatasets := &CreateDatabaseOptions{
		Datasets:        datasetsWithRealFilePaths,
		DatabaseOptions: dbOpts,
		CopyToServer:    true,
	}

	optsWithFakeDatasets := &CreateDatabaseOptions{
		Datasets:        datasetsWithFakeFilePaths,
		DatabaseOptions: dbOpts,
		CopyToServer:    true,
	}

	ctx := context.Background()
	info, _, err := client.DatabaseAdmin.Create(ctx, dbName, optsWithRealDatasets)
	if err != nil {
		t.Errorf("DatabaseAdmin.Create returned error: %v", err)
	}
	if info == nil {
		t.Errorf("DatabaseAdmin.Create should return information string for succesful db creation.")
	}

	info, _, err = client.DatabaseAdmin.Create(ctx, dbName, optsWithFakeDatasets)
	if err == nil {
		t.Error("DatabaseAdmin.Create should return an error due to not being able to find the files.")
	}
	if info != nil {
		t.Errorf("DatabaseAdmin.Create should not return information string for succesful db creation.")
	}

	info, _, err = client.DatabaseAdmin.Create(ctx, dbName, nil)
	if err != nil {
		t.Errorf("DatabaseAdmin.Create returned error: %v", err)
	}
	if info == nil {
		t.Errorf("DatabaseAdmin.Create should return information string for succesful db creation.")
	}

	const methodName = "Create"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.Create(nil, dbName, nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})

}

func TestDatabaseAdminService_Restore(t *testing.T) {
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
	_, err := client.DatabaseAdmin.Restore(ctx, pathToBackup, restoreDatabaseOptions)
	if err != nil {
		t.Errorf("DatabaseAdmin.Restore returned error: %v", err)
	}

	const methodName = "Restore"
	testBadOptions(t, methodName, func() (err error) {
		opts := &RestoreDatabaseOptions{
			Name: "restoredDb",
		}
		_, err = client.DatabaseAdmin.Restore(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Restore(nil, pathToBackup, restoreDatabaseOptions)
	})
}

func TestDatabaseAdminService_Repair(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/repair", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.Repair(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Repair returned error: %v", err)
	}

	const methodName = "Repair"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Repair(nil, db)
	})
}

func TestDatabaseAdminService_Optimize(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s/optimize", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.Optimize(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Optimize returned error: %v", err)
	}

	const methodName = "Optimize"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Optimize(nil, db)
	})
}

func TestDatabaseAdminService_Drop(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	db := "db1"

	mux.HandleFunc(fmt.Sprintf("/admin/databases/%s", db), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeApplicationJSON)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.DatabaseAdmin.Drop(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Drop returned error: %v", err)
	}

	const methodName = "Drop"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.Drop(nil, db)
	})
}

func TestDatabaseAdminService_MetadataDocumentation(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.MetadataDocumentation(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.MetadataDocumentation returned error: %v", err)
	}
	if want := databaseOptions; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.MetadataDocumentation = %+v, want %+v", got, want)
	}

	const methodName = "MetadataDocumentation"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.MetadataDocumentation(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_Namespaces(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.Namespaces(ctx, db)
	if err != nil {
		t.Errorf("DatabaseAdmin.Namespaces returned error: %v", err)
	}
	if want := wantNamespaces; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.Namespaces = %+v, want %+v", got, want)
	}

	const methodName = "Namespaces"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.Namespaces(nil, db)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_ImportNamespaces(t *testing.T) {
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
		t.Errorf("DatabaseAdmin.ImportNamespaces: unexpected error during test: %v", err)
	}
	defer rdf.Close()

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

	// pass a directory to force an error
	directory, err := os.Open("./test-resources/")
	if err != nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces: unexpected error during test: %v", err)
	}
	_, _, err = client.DatabaseAdmin.ImportNamespaces(ctx, db, directory)
	if err == nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces expected to return an error passing a directory instead of a file")
	}

	// pass a file with an non rdf file extension
	tempFile, err := os.CreateTemp(".", "import-namespaces-test")
	if err != nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces: unexpected error creating a temp file: %v", err)
	}
	_, _, err = client.DatabaseAdmin.ImportNamespaces(ctx, db, tempFile)
	if err == nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces expected to return an error passing a file without a non-RDF file extension")
	}
	err = os.Remove(tempFile.Name())
	if err != nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces: unexpected error deleting a temp file: %v", err)
	}

	const methodName = "ImportNamespaces"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ImportNamespaces(nil, db, rdf)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})

	// close the file to force an error
	err = rdf.Close()
	if err != nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces: unexpected error during test: %v", err)
	}
	got, _, err = client.DatabaseAdmin.ImportNamespaces(ctx, db, rdf)
	if err == nil {
		t.Errorf("DatabaseAdmin.ImportNamespaces expected to return an error passing a directory instead of a file")
	}
}

func TestDatabaseAdminService_Metadata(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.Metadata(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.Metadata returned error: %v", err)
	}
	if want := wantDatabasOptions; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.Metadata = %+v, want %+v", got, want)
	}

	const methodName = "Metadata"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.Metadata(nil, db, opts)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_SetMetadata(t *testing.T) {
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
	_, err := client.DatabaseAdmin.SetMetadata(ctx, db, opts)
	if err != nil {
		t.Errorf("DatabaseAdmin.SetMetadata returned error: %v", err)
	}
	const methodName = "SetMetadata"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.DatabaseAdmin.SetMetadata(nil, db, opts)
	})
}

func TestDatabaseAdminService_ListDatabases(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.ListDatabases(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.ListDatabases returned error: %v", err)
	}
	if want := wantDatabases; !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.ListDatabases = %+v, want %+v", got, want)
	}

	const methodName = "ListDatabases"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ListDatabases(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_AllMetadata(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.AllMetadata(ctx, "db1")
	if err != nil {
		t.Errorf("DatabaseAdmin.AllMetadata returned error: %v", err)
	}
	if want := wantDatabaseOptions; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.AllMetadata returned map with length %+v, want %+v", got, want)
	}

	const methodName = "AllMetadata"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.AllMetadata(nil, "db1")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_ListWithMetadata(t *testing.T) {
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
	got, _, err := client.DatabaseAdmin.ListWithMetadata(ctx)
	if err != nil {
		t.Errorf("DatabaseAdmin.ListWithMetadata returned error: %v", err)
	}
	if want := wantDatabasesWithOptions; !cmp.Equal(len(got), len(want)) {
		t.Errorf("DatabaseAdmin.ListWithMetadata returned slice has length %+v, want %+v", got, want)
	}

	const methodName = "ListWithMetadata"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.ListWithMetadata(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_Size(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	getDatabaseSizeOptions := &DatabaseSizeOptions{
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
	got, _, err := client.DatabaseAdmin.Size(ctx, dbName, getDatabaseSizeOptions)
	if err != nil {
		t.Errorf("DatabaseAdmin.Size returned error: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("DatabaseAdmin.Size = %+v, want %+v", got, want)
	}

	const methodName = "Size"
	testBadOptions(t, methodName, func() (err error) {
		opts := &DatabaseSizeOptions{
			Exact: true,
		}
		_, _, err = client.DatabaseAdmin.Size(ctx, "\n", opts)
		return err
	})
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.DatabaseAdmin.Size(nil, dbName, getDatabaseSizeOptions)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestDatabaseAdminService_Size_nonIntegerResponse(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	dbName := "db1"

	responseString := "not an integer"

	mux.HandleFunc(fmt.Sprintf("/%s/size", dbName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseString))
	})

	ctx := context.Background()
	_, _, err := client.DatabaseAdmin.Size(ctx, dbName, nil)
	if err == nil {
		t.Fatalf("DatabaseAdmin.Size should return an error if response cannot be converted to an integer")
	}
}

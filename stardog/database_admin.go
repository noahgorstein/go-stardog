package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// SecurityService handles communication with the security related methods of the Stardog API.
type DatabaseAdminService service

// GetDatabaseSizeOptions specifies the optional parameters to the DatabaseAdminService.GetDatabaseSize method.
type GetDatabaseSizeOptions struct {
	Exact bool `url:"exact"`
}

type getDatabasesWithOptionsResponse struct {
	Databases []map[string]interface{} `json:"databases"`
}

type getDatabasesResponse struct {
	Databases []string `json:"databases"`
}

type DatabaseOptionDetails struct {
	Name              string      `json:"name"`
	Type              string      `json:"type"`
	Server            bool        `json:"server"`
	Mutable           bool        `json:"mutable"`
	MutableWhenOnline bool        `json:"mutableWhenOnline"`
	Category          string      `json:"category"`
	Label             string      `json:"label"`
	Description       string      `json:"description"`
	DefaultValue      interface{} `json:"defaultValue"`
}

// GetDatabaseWithOptions returns all the database configuration options and their set values for a database. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getAllDatabaseOptions
func (s *DatabaseAdminService) GetDatabaseWithOptions(ctx context.Context, database string) (map[string]interface{}, *Response, error) {
	u := fmt.Sprintf("admin/databases/%s/options", database)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data map[string]interface{}
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, err
}

// GetDatabasesWithOptions returns all the database configuration options and their set values for all database. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabasesWithOptions
func (s *DatabaseAdminService) GetDatabasesWithOptions(ctx context.Context) ([]map[string]interface{}, *Response, error) {
	u := "admin/databases/options"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data getDatabasesWithOptionsResponse
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Databases, resp, err
}

// GetDatabases returns the names of all databases in the server.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabases
func (s *DatabaseAdminService) GetDatabases(ctx context.Context) ([]string, *Response, error) {
	u := "admin/databases"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data getDatabasesResponse
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Databases, resp, err
}

// GetDatabaseSize returns the sizze of the database. Size is approximate unless the GetDatabaseSizeOptions.Exact field is set to true.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabases
func (s *DatabaseAdminService) GetDatabaseSize(ctx context.Context, database string, opts *GetDatabaseSizeOptions) (*int, *Response, error) {
	u := fmt.Sprintf("%s/size", database)
	urlWithOptions, err := addOptions(u, opts)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypePlainText,
	}
	req, err := s.client.NewRequest(http.MethodGet, urlWithOptions, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	resp, err := s.client.Do(ctx, req, &buf)
	if err != nil {
		return nil, resp, err
	}
	resultAsInt, err := strconv.Atoi(buf.String())
	if err != nil {
		return nil, resp, err
	}
	return &resultAsInt, resp, err
}

// GetAllDatabaseOptionDetails returns information on all database configuration options, including description and example values.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getAllMetaProperties
func (s *DatabaseAdminService) GetAllDatabaseOptionDetails(ctx context.Context) (map[string]DatabaseOptionDetails, *Response, error) {
	u := "admin/config_properties"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data map[string]DatabaseOptionDetails
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, err
}


// Dataset contains a Path and optional Context (A.K.A named graph) to add the 
// the data contained in the Path to the file into.
type Dataset struct {
  // Path to the file to be uploaded to the server
	Path    string
  // The context (A.K.A named graph) for the data contained in the file to be added to.
	Context string
}

type createDatabaseRequest struct {
	Name         string                      `json:"dbname"`
	Options      map[string]interface{}      `json:"options"`
	Files        []createDatabaseRequestFile `json:"files"`
	CopyToServer bool                        `json:"copyToServer"`
}

type createDatabaseRequestFile struct {
	Filename string `json:"filename"`
	Context  string `json:"context,omitempty"`
}

// CreateDatabase creates a database, optionally with RDF and database options.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/createNewDatabase
func (s *DatabaseAdminService) CreateDatabase(ctx context.Context, name string, ds []Dataset, dbOpts map[string]interface{}, copyToServer bool) (*Response, error) {

	body, writer, err := newCreateDatabaseRequestBody(name, dbOpts, ds, copyToServer)
	if err != nil {
		return nil, err
	}
	headerOpts := &requestHeaderOptions{ContentType: writer.FormDataContentType()}
	req, err := s.client.NewMultipartFormDataRequest(
		http.MethodPost,
		"admin/databases",
		headerOpts,
		body)
	return s.client.Do(ctx, req, nil)
}

// newCreateDatabaseRequestBody creates the request body needed for DatabaseAdminService.CreateDatabase
func newCreateDatabaseRequestBody(name string, dbOpts map[string]interface{}, datasets []Dataset, copyToServer bool) (*bytes.Buffer, *multipart.Writer, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	var files = make([]createDatabaseRequestFile, len(datasets))
	for i, dataset := range datasets {
		files[i] = createDatabaseRequestFile{
			Filename: filepath.Base(dataset.Path),
			Context:  dataset.Context,
		}
	}

	req := createDatabaseRequest{
		Name:         name,
		Options:      dbOpts,
		Files:        files,
		CopyToServer: copyToServer,
	}
	json, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}
	writer.WriteField("root", string(json))

	// if files are to be sent to server, check that they exist on host
	if copyToServer {
		for _, dataset := range datasets {
			file, err := os.Open(dataset.Path)
			if err != nil {
				return nil, nil, err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(filepath.Base(dataset.Path), filepath.Base(dataset.Path))
			if err != nil {
				return nil, nil, err
			}
			_, err = io.Copy(part, file)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, nil, err
	}
	return body, writer, err
}

// DropDatabase deletes a database
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/dropDatabase
func (s *DatabaseAdminService) DropDatabase(ctx context.Context, name string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s", name)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodDelete, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// OptimizeDatabase optimizes a database
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/optimizeDatabase
func (s *DatabaseAdminService) OptimizeDatabase(ctx context.Context, name string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/optimize", name)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// RepairDatabase attempts to recover a corrupted database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/repairDatabase
func (s *DatabaseAdminService) RepairDatabase(ctx context.Context, name string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/repair", name)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPost, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

type RestoreDatabaseOptions struct {
	// Whether or not to overwrite an existing database with this backup
	Force bool `url:"force,omitempty"`

	// The name of the restored database, if different
	Name string `url:"name,omitempty"`
}

// RestoreDatabase restores a database backup located at the path on the server
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/restoreDatabase
func (s *DatabaseAdminService) RestoreDatabase(ctx context.Context, path string, opts *RestoreDatabaseOptions) (*Response, error) {
	u := fmt.Sprintf("admin/restore?from=%s", path)
	urlWithOptions, err := addOptions(u, opts)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, urlWithOptions, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

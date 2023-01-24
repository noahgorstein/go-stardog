package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DatabaseAdminService handles communication with the database admin related methods of the Stardog API.
type DatabaseAdminService service

// GetDatabaseSizeOptions specifies the optional parameters to the DatabaseAdminService.GetDatabaseSize method.
type GetDatabaseSizeOptions struct {
	Exact bool `url:"exact"`
}

// GenerateDataModelOutputFormat represents an output format for DatabaseAdminService.GenerateDataModel
type GenerateDataModelOutputFormat string

const (
	Text    GenerateDataModelOutputFormat = "text"
	OWL     GenerateDataModelOutputFormat = "owl"
	SHACL   GenerateDataModelOutputFormat = "shacl"
	SQL     GenerateDataModelOutputFormat = "sql"
	GraphQL GenerateDataModelOutputFormat = "graphql"
)

// ImportNamespacesResponse contains information returned
// after DatabaseAdminService.ImportNamespaces completed successfully.
type ImportNamespacesResponse struct {
	NumberImportedNamespaces int      `json:"numImportedNamespaces"`
	UpdatedNamespaces        []string `json:"namespaces"`
}

// GenerateDataModelOptions are options for the DatabaseAdminService.GenerateDataModel method
type GenerateDataModelOptions struct {
	// Enable reasoning
	Reasoning bool `url:"reasoning,omitempty"`

	// Desired output format
	Output GenerateDataModelOutputFormat `url:"output,omitempty"`
}

// RestoreDatabaseOptions are options for the DatabaseAdminService.RestoreDatabase method
type RestoreDatabaseOptions struct {
	// Whether or not to overwrite an existing database with this backup
	Force bool `url:"force,omitempty"`

	// The name of the restored database, if different
	Name string `url:"name,omitempty"`
}

// Namespace represents a namespace
type Namespace struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

// DatabaseOptionDetails represents a database configuration option's details.
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

type createDatabaseResponse struct {
	Message *string `json:"message"`
}

// CreateDatabaseOptions specifies the optional parameters to the DatabaseAdminService.CreateDatabase method.
type CreateDatabaseOptions struct {
	// The data to be bulk-loaded to the database at creation time
	Datasets []Dataset
	// Database configuration options
	DatabaseOptions map[string]interface{}
	// Whether to send the file contents to the server. Use if data exists client-side.
	CopyToServer bool
}

type Dataset struct {
	// Path to the file to be uploaded to the server
	Path string
	// The optional named-graph (A.K.A context) for the data contained in the file to be added to.
	NamedGraph string
}

// ExportDataOptions specifies the optional parameters to the DatabaseAdmin.ExportData method.
type ExportDataOptions struct {
	// The named graph(s) to export from the dataset
	NamedGraph []string `url:"named-graph-uri"`

	// The RDF format for the exported data
	Format RDFFormat `url:"-"`

	// Compression format for the exported data. **Only applicable if data is exported ServerSide**
	Compression Compression `url:"compression,omitempty"`

	// Export the data to the server
	ServerSide bool `url:"server-side,omitempty"`
}

// ExportObfuscatedDataOptions specifies the optional parameters to the DatabaseAdmin.ExportObfuscatedData method.
type ExportObfuscatedDataOptions struct {
	// The named graph(s) to export from the dataset
	NamedGraph []string `url:"named-graph-uri"`

	// The RDF format for the exported data
	Format RDFFormat `url:"-"`

	// Compression format for the exported data. **Only applicable if data is exported ServerSide**
	Compression Compression `url:"compression,omitempty"`

	// Export the data to Stardog's export dir ($STARDOG_HOME/.exports by default)
	ServerSide bool `url:"server-side,omitempty"`

	// Configuration file for obfuscation.
	// See https://github.com/stardog-union/stardog-examples/blob/master/config/obfuscation.ttl for an example configuration file.
	ObfuscationConfig *os.File
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

type getNamespaceResponse struct {
	Namespaces []Namespace `json:"namespaces"`
}

type getDatabasesWithOptionsResponse struct {
	Databases []map[string]interface{} `json:"databases"`
}

type getDatabasesResponse struct {
	Databases []string `json:"databases"`
}

// GetDatabaseOptions returns the value of specific metadata options for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getDatabaseOptions
func (s *DatabaseAdminService) GetDatabaseOptions(ctx context.Context, database string, opts []string) (map[string]interface{}, *Response, error) {
	u := fmt.Sprintf("admin/databases/%s/options", database)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
		Accept:      mediaTypeApplicationJSON,
	}

	optionMap := map[string]interface{}{}
	for _, opt := range opts {
		optionMap[opt] = ""
	}

	req, err := s.client.NewRequest(http.MethodPut, u, &headerOpts, optionMap)
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

// SetDatabaseOptions sets the value of specific configuration options for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/setDatabaseOption
func (s *DatabaseAdminService) SetDatabaseOptions(ctx context.Context, database string, opts map[string]interface{}) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/options", database)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
		Accept:      mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPost, u, &headerOpts, opts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, err
}

// GetAllDatabaseOptions returns all the database configuration options and their set values for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getAllDatabaseOptions
func (s *DatabaseAdminService) GetAllDatabaseOptions(ctx context.Context, database string) (map[string]interface{}, *Response, error) {
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

// GetAllDatabasesWithOptions returns all the database configuration options and their set values for all databases.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabasesWithOptions
func (s *DatabaseAdminService) GetAllDatabasesWithOptions(ctx context.Context) ([]map[string]interface{}, *Response, error) {
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

// GetNamespaces retrieves the namespaces stored in the database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getNamespaces
func (s *DatabaseAdminService) GetNamespaces(ctx context.Context, database string) ([]Namespace, *Response, error) {
	u := fmt.Sprintf("%s/namespaces", database)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data getNamespaceResponse
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Namespaces, resp, err
}

// ImportNamespaces adds namespaces to the database that are declared in the RDF file.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getNamespaces
func (s *DatabaseAdminService) ImportNamespaces(ctx context.Context, database string, file *os.File) (*ImportNamespacesResponse, *Response, error) {
	u := fmt.Sprintf("%s/namespaces", database)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	var requestBody *bytes.Buffer
	if file != nil {
		stat, err := file.Stat()
		if err != nil {
			return nil, nil, err
		}
		if stat.IsDir() {
			return nil, nil, errors.New("the file containing the namespaces can't be a directory")
		}

		requestBytes, err := io.ReadAll(file)
		if err != nil {
			return nil, nil, err
		}
		requestBody = bytes.NewBuffer(requestBytes)

		rdfFormat, err := GetRDFFormatFromExtension(file.Name())
		if err != nil {
			return nil, nil, err
		}

		headerOpts.ContentType = rdfFormat.String()
	}

	req, err := s.client.NewRequest(http.MethodPost, u, &headerOpts, requestBody)
	if err != nil {
		return nil, nil, err
	}

	var importNamespacesResponse ImportNamespacesResponse
	resp, err := s.client.Do(ctx, req, &importNamespacesResponse)
	if err != nil {
		return nil, resp, err
	}
	return &importNamespacesResponse, resp, err
}

// GetDatabaseSize returns the size of the database. Size is approximate unless the GetDatabaseSizeOptions.Exact field is set to true.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabases
func (s *DatabaseAdminService) GetDatabaseSize(ctx context.Context, database string, opts *GetDatabaseSizeOptions) (*int, *Response, error) {
	u := fmt.Sprintf("%s/size", database)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}
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

// GetAllDatabaseOptionDetails returns information about all available database configuration options, including description and example values.
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

// CreateDatabase creates a database, optionally with RDF and database options. CreateDatabase assumes that the
// Paths in the Dataset(s) provided for CreateDatabaseOptions.Datasets exist on the server. If they are client side,
// provide a value of true for CreateDatabaseOptions.CopyToServer
//
// If the database creation is successful a *string containing details about the database creation will be returned
// such as:
// Bulk loading data to new database db1.
// Loaded 41,099 triples to db1 from 1 file(s) in 00:00:00.487 @ 84.4K triples/sec.
// Successfully created database 'db1'.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/createNewDatabase
func (s *DatabaseAdminService) CreateDatabase(ctx context.Context, name string, opts *CreateDatabaseOptions) (*string, *Response, error) {
	body, writer, err := newCreateDatabaseRequestBody(name, opts)
	if err != nil {
		return nil, nil, err
	}
	headerOpts := &requestHeaderOptions{
		ContentType: writer.FormDataContentType(),
		Accept:      mediaTypeApplicationJSON,
	}
	req, err := s.client.NewMultipartFormDataRequest(
		http.MethodPost,
		"admin/databases",
		headerOpts,
		body)
	if err != nil {
		return nil, nil, err
	}

	var createDatabaseResponse createDatabaseResponse
	resp, err := s.client.Do(ctx, req, &createDatabaseResponse)
	if err != nil {
		return nil, resp, err
	}
	return createDatabaseResponse.Message, resp, nil
}

// newCreateDatabaseRequestBody creates the request body needed for DatabaseAdminService.CreateDatabase
func newCreateDatabaseRequestBody(name string, opts *CreateDatabaseOptions) (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	req := createDatabaseRequest{
		Name: name,
	}

	if opts != nil {
		if opts.Datasets != nil {
			var files = make([]createDatabaseRequestFile, len(opts.Datasets))
			for i, dataset := range opts.Datasets {
				files[i] = createDatabaseRequestFile{
					Filename: dataset.Path,
					Context:  dataset.NamedGraph,
				}
			}
			req.Files = files
		}
		if opts.DatabaseOptions == nil {
			req.Options = make(map[string]interface{})
		} else {
			req.Options = opts.DatabaseOptions
		}
		req.CopyToServer = opts.CopyToServer
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}
	err = writer.WriteField("root", string(jsonReq))
	if err != nil {
		return nil, nil, err
	}

	// if files are to be sent to server, check that they exist on host
	if opts != nil && opts.CopyToServer && opts.Datasets != nil {
		for _, dataset := range opts.Datasets {
			file, err := os.Open(dataset.Path)
			if err != nil {
				return nil, nil, err
			}

			part, err := writer.CreateFormFile(filepath.Base(dataset.Path), filepath.Base(dataset.Path))
			if err != nil {
				return nil, nil, err
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return nil, nil, err
			}

			err = file.Close()
			if err != nil {
				return nil, nil, err
			}
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

// RestoreDatabase restores a database backup located at the path on the server
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/restoreDatabase
func (s *DatabaseAdminService) RestoreDatabase(ctx context.Context, path string, opts *RestoreDatabaseOptions) (*Response, error) {
	u := fmt.Sprintf("admin/restore?from=%s", path)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}
	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, urlWithOptions, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// OnlineDatabase onlines a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/onlineDatabase
func (s *DatabaseAdminService) OnlineDatabase(ctx context.Context, name string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/online", name)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// OfflineDatabase onlines a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/offlineDatabase
func (s *DatabaseAdminService) OfflineDatabase(ctx context.Context, name string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/offline", name)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// GenerateDataModel generates the reasoning model used by this database in various formats
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/generateModel
func (s *DatabaseAdminService) GenerateDataModel(ctx context.Context, database string, opts *GenerateDataModelOptions) (*bytes.Buffer, *Response, error) {
	u := fmt.Sprintf("%s/model", database)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, urlWithOptions, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var writer bytes.Buffer
	resp, err := s.client.Do(ctx, req, &writer)
	if err != nil {
		return nil, resp, err
	}
	return &writer, resp, err
}

func getExportFormatFromRDFFormat(format RDFFormat) (string, error) {
	switch format {
	case RDFFormatTrig:
		return "trig", nil
	case RDFFormatTurtle:
		return "turtle", nil
	case RDFFormatJSONLD:
		return "jsonld", nil
	case RDFFormatNQuads:
		return "nquads", nil
	case RDFFormatNTriples:
		return "ntriples", nil
	case RDFFormatRDFXML:
		return "rdfxml", nil
	default:
		return "", errors.New("supported RDF formats for export are Trig, Turtle, JSONLD, NQUADS, NTRIPLES, and RDFXML")
	}
}

// ExportData exports RDF data from the database.
// If ExportDataOptions.ServerSide=true, the RDF using the specified format will be saved in the export directory
// for the server. The default server export directory is ‘.exports’ in the $STARDOG_HOME
// but can be changed via ‘export.dir’ in the stardog.properties file.
// In this case, some information will be returned about the export instead of the RDF such as:
// "Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms"
//
// Starodg API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/exportDatabase
func (s *DatabaseAdminService) ExportData(ctx context.Context, database string, opts *ExportDataOptions) (*bytes.Buffer, *Response, error) {
	u := fmt.Sprintf("%s/export", database)

	requestHeaderOptions := &requestHeaderOptions{}

	if opts != nil {
		if opts.Format.Valid() {
			if !opts.ServerSide {
				requestHeaderOptions.Accept = opts.Format.String()
			} else {
				// if server side export, Stardog will return some details about the successful import in plain text
				// i.e. Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms
				requestHeaderOptions.Accept = mediaTypePlainText
				format, err := getExportFormatFromRDFFormat(opts.Format)
				if err != nil {
					return nil, nil, err
				}
				u += fmt.Sprintf("?format=%s", format)
			}
		}
	}

	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, urlWithOptions, requestHeaderOptions, nil)
	if err != nil {
		return nil, nil, err
	}

	var writer bytes.Buffer
	resp, err := s.client.Do(ctx, req, &writer)
	if err != nil {
		return nil, resp, err
	}
	return &writer, resp, err
}

// ExportObfuscatedData exports obfuscated RDF data from the database.
//
// If nil is provided for ExportObfuscatedDataOptions.ObfuscationConfig, Stardog will use its default
// obfuscation configuration. All URIs, bnodes, and string literals in the database will be
// obfuscated using the SHA256 message digest algorithm. Non-string typed literals (numbers, dates, etc.)
// are left unchanged as well as URIs from built-in namespaces (e.g. RDF, RDFS, OWL, etc.)
//
// If ExportObfuscatedDataOptions.ServerSide=true, the RDF using the specified format will be saved in the export directory
// for the server. The default server export directory is ‘.exports’ in the $STARDOG_HOME
// but can be changed via ‘export.dir’ in the stardog.properties file.
// In this case, some information will be returned about the export instead of the RDF such as:
// "Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms"
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/exportDatabaseObfuscated
func (s *DatabaseAdminService) ExportObfuscatedData(ctx context.Context, database string, opts *ExportObfuscatedDataOptions) (*bytes.Buffer, *Response, error) {
	u := fmt.Sprintf("%s/export", database)

	requestHeaderOptions := &requestHeaderOptions{}

	// in order to use Stardog's default obfuscation configuration, it expects a GET request
	httpMethod := http.MethodGet

	var requestBody *bytes.Buffer
	if opts != nil && opts.ObfuscationConfig != nil {
		// if using custom obfuscation configuration, request should be a POST
		httpMethod = http.MethodPost

		stat, err := opts.ObfuscationConfig.Stat()
		if err != nil {
			return nil, nil, err
		}
		if stat.IsDir() {
			return nil, nil, errors.New("the obfuscation configuration file can't be a directory")
		}

		requestBytes, err := io.ReadAll(opts.ObfuscationConfig)
		if err != nil {
			return nil, nil, err
		}
		requestBody = bytes.NewBuffer(requestBytes)
		requestHeaderOptions.ContentType = RDFFormatTurtle.String()
	} else {
		// if no obfuscation configuration is provided use Stardog's default one
		u = u + "?obf=DEFAULT"
	}

	if opts != nil {
		if opts.Format.Valid() {
			if !opts.ServerSide {
				requestHeaderOptions.Accept = opts.Format.String()
			} else {
				requestHeaderOptions.Accept = mediaTypePlainText
				format, err := getExportFormatFromRDFFormat(opts.Format)
				if err != nil {
					return nil, nil, err
				}
				if strings.Contains(u, "?obf=DEFAULT") {
					u += "&"
				} else {
					u += "?"
				}
				u += fmt.Sprintf("format=%s", format)
			}
		}
	}

	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	var req *http.Request
	if requestBody != nil && len(requestBody.Bytes()) > 0 {
		req, err = s.client.NewRequest(httpMethod, urlWithOptions, requestHeaderOptions, requestBody)
		if err != nil {
			return nil, nil, err
		}
	} else {
		req, err = s.client.NewRequest(httpMethod, urlWithOptions, requestHeaderOptions, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	var writer bytes.Buffer
	resp, err := s.client.Do(ctx, req, &writer)
	if err != nil {
		return nil, resp, err
	}
	return &writer, resp, err
}

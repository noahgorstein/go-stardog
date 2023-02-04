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

// DatabaseSizeOptions specifies the optional parameters to the [DatabaseAdminService.Size] method.
type DatabaseSizeOptions struct {
	Exact bool `url:"exact"`
}

// DataModelFormat represents an output format for [DatabaseAdminService.DataModel].
// The zero value for a DataModelFormat is  [DataModelFormatUnknown]
type DataModelFormat int

// All available DataModelFormats
const (
	DataModelFormatUnknown DataModelFormat = iota
	DataModelFormatText
	DataModelFormatOWL
	DataModelFormatSHACL
	DataModelFormatSQL
	DataModelFormatGraphQL
)

// dataModelFormatValues maps each DataModelFormat to its
// string representation
var dataModelFormatValues = [6]string{
	DataModelFormatUnknown: "",
	DataModelFormatText:    "text",
	DataModelFormatOWL:     "owl",
	DataModelFormatSHACL:   "shacl",
	DataModelFormatSQL:     "sql",
	DataModelFormatGraphQL: "graphql",
}

// Valid returns if the DataModelFormat is known (valid) or not.
func (f DataModelFormat) Valid() bool {
	return !(f <= DataModelFormatUnknown || int(f) >= len(dataModelFormatValues))
}

// String will return the string representation of the DataModelFormat
func (f DataModelFormat) String() string {
	if !f.Valid() {
		return dataModelFormatValues[DataModelFormatUnknown]
	}
	return dataModelFormatValues[f]
}

// ImportNamespacesResponse contains information returned
// after [DatabaseAdminService.ImportNamespaces] completed successfully.
type ImportNamespacesResponse struct {
	NumberImportedNamespaces int      `json:"numImportedNamespaces"`
	UpdatedNamespaces        []string `json:"namespaces"`
}

// DataModelOptions are options for the [DatabaseAdminService.DataModel] method
type DataModelOptions struct {
	// Enable reasoning
	Reasoning bool `url:"reasoning,omitempty"`

	// Desired output format
	OutputFormat DataModelFormat `url:"output,omitempty"`
}

// RestoreDatabaseOptions are options for the [DatabaseAdminService.Restore] method
type RestoreDatabaseOptions struct {
	// Whether or not to overwrite an existing database with this backup
	Force bool `url:"force,omitempty"`

	// The name of the restored database, if different than the name of the backup being restored
	Name string `url:"name,omitempty"`
}

// Namespace represents a [Stardog database namespace].
//
// [Stardog database namespace]: https://docs.stardog.com/operating-stardog/database-administration/managing-databases#namespaces
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

// CreateDatabaseOptions specifies the optional parameters to the [DatabaseAdminService.CreateDatabase] method.
type CreateDatabaseOptions struct {
	// The data to be bulk-loaded to the database at creation time
	Datasets []Dataset
	// Database configuration options
	DatabaseOptions map[string]interface{}
	// Whether to send the file contents to the server. Use if data exists client-side.
	CopyToServer bool
}

// Dataset is used to specify a dataset (filepath and named graph to add data into) to be added to a Stardog database.
type Dataset struct {
	// Path to the file to be uploaded to the server
	Path string
	// The optional named-graph (A.K.A context) for the data contained in the file to be added to.
	NamedGraph string
}

// ExportDataOptions specifies the optional parameters to the [DatabaseAdminService.ExportData] method.
type ExportDataOptions struct {
	// The named graph(s) to export from the dataset
	NamedGraph []string `url:"named-graph-uri,omitempty"`

	// The RDF format for the exported data
	Format RDFFormat `url:"-"`

	// Compression format for the exported data. **Only applicable if data is exported ServerSide**
	Compression Compression `url:"compression,omitempty"`

	// Export the data to the server
	ServerSide bool `url:"server-side,omitempty"`
}

// ExportObfuscatedDataOptions specifies the optional parameters to
// the [DatabaseAdminService.ExportObfuscatedData] method.
type ExportObfuscatedDataOptions struct {
	// The named graph(s) to export from the dataset
	NamedGraph []string `url:"named-graph-uri,omitempty"`

	// The RDF format for the exported data
	Format RDFFormat `url:"-"`

	// Compression format for the exported data. **Only applicable if data is exported ServerSide**
	Compression Compression `url:"compression,omitempty"`

	// Export the data to Stardog's export dir ($STARDOG_HOME/.exports by default)
	ServerSide bool `url:"server-side,omitempty"`

	// Configuration file for obfuscation.
	// See https://github.com/stardog-union/stardog-examples/blob/master/config/obfuscation.ttl for an example configuration file.
	ObfuscationConfig *os.File `url:"-"`
}

// response for Namespaces
type databaseNamespacesResponse struct {
	Namespaces []Namespace `json:"namespaces"`
}

// response for ListWithMetadata
type listDatabasesWithMetadataResponse struct {
	Databases []map[string]interface{} `json:"databases"`
}

// response for List
type listDatabasesResponse struct {
	Databases []string `json:"databases"`
}

// createDatabaseRequest is the JSON Create needs to satisfy the request body
// Stardog requires for datbase creation
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

// response for Create
type createDatabaseResponse struct {
	Message *string `json:"message"`
}

// Metadata returns the value of specific metadata options for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getDatabaseOptions
func (s *DatabaseAdminService) Metadata(ctx context.Context, database string, opts []string) (map[string]interface{}, *Response, error) {
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

// SetMetadata sets the value of specific configuration options (a.k.a. metadata) for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/setDatabaseOption
func (s *DatabaseAdminService) SetMetadata(ctx context.Context, database string, opts map[string]interface{}) (*Response, error) {
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

// AllMetadata returns all the database configuration options (a.k.a. metadata)
// and their set values for a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getAllDatabaseOptions
func (s *DatabaseAdminService) AllMetadata(ctx context.Context, database string) (map[string]interface{}, *Response, error) {
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

// ListWithMetadata returns all databases with their database configuration options (a.k.a. metadata)
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabasesWithOptions
func (s *DatabaseAdminService) ListWithMetadata(ctx context.Context) ([]map[string]interface{}, *Response, error) {
	u := "admin/databases/options"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data listDatabasesWithMetadataResponse
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Databases, resp, err
}

// ListDatabases returns the names of all databases in the server.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabases
func (s *DatabaseAdminService) ListDatabases(ctx context.Context) ([]string, *Response, error) {
	u := "admin/databases"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data listDatabasesResponse
	resp, err := s.client.Do(ctx, req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Databases, resp, err
}

// Namespaces retrieves the namespaces stored in the database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getNamespaces
func (s *DatabaseAdminService) Namespaces(ctx context.Context, database string) ([]Namespace, *Response, error) {
	u := fmt.Sprintf("%s/namespaces", database)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var data databaseNamespacesResponse
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

	var requestBody bytes.Buffer
	if file != nil {
		stat, err := file.Stat()
		if err != nil {
			return nil, nil, err
		}
		if stat.IsDir() {
			return nil, nil, errors.New("the file containing the namespaces can't be a directory")
		}

		rdfFormat, err := GetRDFFormatFromExtension(file.Name())
		if err != nil {
			return nil, nil, err
		}
		headerOpts.ContentType = rdfFormat.String()

		_, err = io.Copy(&requestBody, file)
		if err != nil {
			return nil, nil, err
		}
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

// Size returns the size of the database. Size is approximate unless the GetDatabaseSizeOptions.Exact field is set to true.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/listDatabases
func (s *DatabaseAdminService) Size(ctx context.Context, database string, opts *DatabaseSizeOptions) (*int, *Response, error) {
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

// MetadataDocumentation returns information about all available database configuration options
// (a.k.a. metadata) including description and example values.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/getAllMetaProperties
func (s *DatabaseAdminService) MetadataDocumentation(ctx context.Context) (map[string]DatabaseOptionDetails, *Response, error) {
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

// Create creates a database, optionally with RDF and database options. Create assumes that the
// Paths in the Dataset(s) provided for CreateDatabaseOptions.Datasets exist on the server. If they are client side,
// provide a value of true for CreateDatabaseOptions.CopyToServer
//
// If the database creation is successful a *string containing details about the database creation will be returned
// such as:
//
//	Bulk loading data to new database db1.
//	Loaded 41,099 triples to db1 from 1 file(s) in 00:00:00.487 @ 84.4K triples/sec.
//	Successfully created database 'db1'.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/createNewDatabase
func (s *DatabaseAdminService) Create(ctx context.Context, name string, opts *CreateDatabaseOptions) (*string, *Response, error) {
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
		// initialize Files and Options to make sure [], {} respectively instead of null
		// is in the JSON request sent to Stardog if no Files or Options
		Files:   make([]createDatabaseRequestFile, 0),
		Options: make(map[string]interface{}),
	}

	if opts != nil {
		if opts.Datasets != nil {
			req.Files = make([]createDatabaseRequestFile, len(opts.Datasets))
			for i, dataset := range opts.Datasets {
				req.Files[i] = createDatabaseRequestFile{
					Filename: dataset.Path,
					Context:  dataset.NamedGraph,
				}
			}
		}
		if opts.DatabaseOptions != nil {
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

// Drop deletes a database
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/dropDatabase
func (s *DatabaseAdminService) Drop(ctx context.Context, database string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s", database)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodDelete, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Optimize optimizes a database
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/optimizeDatabase
func (s *DatabaseAdminService) Optimize(ctx context.Context, database string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/optimize", database)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Repair attempts to recover a corrupted database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/repairDatabase
func (s *DatabaseAdminService) Repair(ctx context.Context, database string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/repair", database)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPost, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Restore restores a database backup located at the path on the server
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/restoreDatabase
func (s *DatabaseAdminService) Restore(ctx context.Context, path string, opts *RestoreDatabaseOptions) (*Response, error) {
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

// Online onlines a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/onlineDatabase
func (s *DatabaseAdminService) Online(ctx context.Context, database string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/online", database)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Offline onlines a database.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/offlineDatabase
func (s *DatabaseAdminService) Offline(ctx context.Context, database string) (*Response, error) {
	u := fmt.Sprintf("admin/databases/%s/offline", database)

	reqHeaderOpts := &requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	req, err := s.client.NewRequest(http.MethodPut, u, reqHeaderOpts, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// DataModel generates the reasoning model used by this database in various formats
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/generateModel
func (s *DatabaseAdminService) DataModel(ctx context.Context, database string, opts *DataModelOptions) (*bytes.Buffer, *Response, error) {
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

// ExportData exports RDF data from the database.
// If ExportDataOptions.ServerSide=true, the RDF using the specified format will be saved in the export directory
// for the server. The default server export directory is ‘.exports’ in the $STARDOG_HOME
// but can be changed via ‘export.dir’ in the stardog.properties file.
// In this case, some information will be returned about the export instead of the RDF such as:
//
//	Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms
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
				format, err := opts.Format.toExportFormat()
				// this is very unlikely to happen because a check to see if format is valid is done earlier
				if err != nil {
					return nil, nil, err
				}
				u += fmt.Sprintf("?format=%s", format)

				// if server side export, Stardog will return some details about the successful import in plain text
				// i.e. Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms
				requestHeaderOptions.Accept = mediaTypePlainText
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

// ExportObfuscatedData exports [obfuscated RDF data] from the database.
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
//
//	Exported 28 statements from db1 to /stardog-home/.exports/db1-2023-01-15.trig in 2.551 ms
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/DB-Admin/operation/exportDatabaseObfuscated
//
// [obfuscated RDF data]: https://docs.stardog.com/query-stardog/obfuscating-data
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
				format, err := opts.Format.toExportFormat()
				// this is unlikely to occur, since we check if RDFFormat is Valid
				if err != nil {
					return nil, nil, err
				}
				// if obfuscation configuration was NOT provided
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

package stardog

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SPARQLService handles communication with the Sparql methods of the Stardog API.
type SPARQLService service

// SelectOptions specifies the optional parameters to the SPARQLService.Select method
type SelectOptions struct {
	// Enable reasoning
	Reasoning bool `url:"reasoning,omitempty"`
	// The name of the schema
	Schema string `url:"schema,omitempty"`
	// The transaction ID
	TxID string `url:"txid,omitempty"`
	// Base URI against which to resolve relative URIs
	BaseURI string `url:"baseURI,omitempty"`
	// The number of milliseconds after which the query should timeout
	Timeout int `url:"timeout,omitempty"`
	// The maximum number of results to return
	Limit int `url:"limit,omitempty"`
	// How far into the result set to offset
	Offset int `url:"offset,omitempty"`
	// Request query results with namespace substitution/prefix lines
	UseNamespaces bool `url:"useNamespaces,omitempty"`
	// URI(s) to be used as the default graph (equivalent to FROM)
	DefaultGraphURI string `url:"default-graph-uri,omitempty"`
	// URI(s) to be used as named graphs (equivalent to FROM NAMED)
	NamedGraphURI string `url:"named-graph-uri,omitempty"`

	// Result format of the query results
	ResultFormat QueryResultFormat `url:"-"`
}

// UpdateOptions specifies the optional parameters to the SPARQLService.Update method
type UpdateOptions struct {
	// Enable reasoning
	Reasoning bool `url:"reasoning,omitempty"`
	// The name of the schema
	Schema string `url:"schema,omitempty"`
	// The transaction ID
	TxID string `url:"txid,omitempty"`
	// Base URI against which to resolve relative URIs
	BaseURI string `url:"baseURI,omitempty"`
	// The number of milliseconds after which the query should timeout
	Timeout int `url:"timeout,omitempty"`
	// The maximum number of results to return
	Limit int `url:"limit,omitempty"`
	// How far into the result set to offset
	Offset int `url:"offset,omitempty"`
	// Request query results with namespace substitution/prefix lines
	UseNamespaces bool `url:"useNamespaces,omitempty"`
	// URI(s) to be used as the default graph (equivalent to FROM)
	DefaultGraphURI string `url:"default-graph-uri,omitempty"`
	// URI(s) to be used as named graphs (equivalent to FROM NAMED)
	NamedGraphURI string `url:"named-graph-uri,omitempty"`
	// URI(s) to be used as default graph (equivalent to USING)
	UsingGraphURI string `url:"using-graph-uri,omitempty"`
	// URI(s) to be used as named graphs (equivalent to USING NAMED)
	UsingNamedGraphURI string `url:"using-named-graph-uri,omitempty"`
	// URI of the graph to be inserted into
	InsertGraphURI string `url:"insert-graph-uri,omitempty"`
	// URI of the graph to be removed from
	RemoveGraphURI string `url:"remove-graph-uri,omitempty"`
}

type QueryResultFormat int

const (
	QueryResultFormatUnknown QueryResultFormat = iota
	QueryResultFormatTrig
	QueryResultFormatTurtle
	QueryResultFormatRDFXML
	QueryResultFormatNTriples
	QueryResultFormatNQuads
	QueryResultFormatJSONLD
	QueryResultFormatSparqlResultsJSON
	QueryResultFormatSparqlResultsXML
	QueryResultFormatCSV
	QueryResultFormatTSV
)

func (q QueryResultFormat) Valid() bool {
	return !(q <= QueryResultFormatUnknown || int(q) >= len(queryResultFormatValues()))
}

//revive:disable:add-constant
func queryResultFormatValues() [11]string {
	return [11]string{
		QueryResultFormatUnknown:           "UNKNOWN",
		QueryResultFormatTrig:              mediaTypeApplicationTrig,
		QueryResultFormatTurtle:            mediaTypeTextTurtle,
		QueryResultFormatRDFXML:            mediaTypeApplicationRDFXML,
		QueryResultFormatNTriples:          mediaTypeApplicationNTriples,
		QueryResultFormatNQuads:            mediaTypeApplicationNQuads,
		QueryResultFormatJSONLD:            mediaTypeApplicationJSONLD,
		QueryResultFormatSparqlResultsJSON: mediaTypeApplicationSparqlResultsJSON,
		QueryResultFormatSparqlResultsXML:  mediaTypeApplicationSparqlResultsXML,
		QueryResultFormatCSV:               mediaTypeTextCSV,
		QueryResultFormatTSV:               mediaTypeTextTSV,
	}
}

//revive:enable:add-constant

func (q QueryResultFormat) String() string {
	if !q.Valid() {
		return queryResultFormatValues()[QueryPlanFormatUnknown]
	}
	return queryResultFormatValues()[q]
}

type QueryPlanFormat int

const (
	QueryPlanFormatUnknown QueryPlanFormat = iota
	QueryPlanFormatText
	QueryPlanFormatJSON
)

func (q QueryPlanFormat) Valid() bool {
	return !(q <= QueryPlanFormatUnknown || int(q) >= len(queryPlanFormatValues()))
}

//revive:disable:add-constant
func queryPlanFormatValues() [3]string {
	return [3]string{
		QueryPlanFormatUnknown: "UNKNOWN",
		QueryPlanFormatText:    mediaTypePlainText,
		QueryPlanFormatJSON:    mediaTypeApplicationJSON,
	}
}

//revive:enable:add-constant

func (q QueryPlanFormat) String() string {
	if !q.Valid() {
		return queryPlanFormatValues()[QueryPlanFormatUnknown]
	}
	return queryPlanFormatValues()[q]
}

// ExplainOptions specifies the optional parameters to the SPARQLService.Explain method
type ExplainOptions struct {
	// Enable reasoning
	Reasoning bool `url:"reasoning,omitempty"`
	// Run the query profiler
	Profile bool `url:"profile,omitempty"`

	// Format to return query plan in (QueryPlanFormatText is the default)
	QueryPlanFormat QueryPlanFormat `url:"-"`
}

// Select performs a SPARQL select query
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/SPARQL/operation/getSparqlQuery
func (s *SPARQLService) Select(ctx context.Context, database string, query string, opts *SelectOptions) (*bytes.Buffer, *Response, error) {
	encodedQuery := url.QueryEscape(query)
	u := fmt.Sprintf("%s/query?query=%s", database, encodedQuery)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}
	headerOpts := requestHeaderOptions{}

	if opts == nil || (opts != nil && !opts.ResultFormat.Valid()) {
		headerOpts.Accept = QueryResultFormatSparqlResultsJSON.String()
	} else {
		headerOpts.Accept = opts.ResultFormat.String()
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
	return &buf, resp, err
}

// Update performs a SPARQL update query
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/SPARQL/operation/updateGet
func (s *SPARQLService) Update(ctx context.Context, database string, query string, opts *UpdateOptions) (*Response, error) {
	encodedQuery := url.QueryEscape(query)
	u := fmt.Sprintf("%s/update?query=%s", database, encodedQuery)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, urlWithOptions, nil, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Retrieves a query plan
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/SPARQL/operation/explainQueryGet
func (s *SPARQLService) Explain(ctx context.Context, database string, query string, opts *ExplainOptions) (*bytes.Buffer, *Response, error) {
	encodedQuery := url.QueryEscape(query)
	u := fmt.Sprintf("%s/explain?query=%s", database, encodedQuery)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}
	headerOpts := requestHeaderOptions{}

	if opts == nil || (opts != nil && !opts.QueryPlanFormat.Valid()) {
		headerOpts.Accept = QueryPlanFormatText.String()
	} else {
		headerOpts.Accept = opts.QueryPlanFormat.String()
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
	return &buf, resp, err
}

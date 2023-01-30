package stardog

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// RDFFormat represents an [RDF Serialization Format].
// The zero value for an RDFFormat is [RDFFormatUnknown]
//
// [RDF Serialization Format]: https://www.w3.org/wiki/RdfSyntax
type RDFFormat int

// All available RDF Formats in Stardog.
const (
	RDFFormatUnknown RDFFormat = iota
	RDFFormatTrig
	RDFFormatTurtle
	RDFFormatRDFXML
	RDFFormatNTriples
	RDFFormatNQuads
	RDFFormatJSONLD
)

// Valid returns if a given RDFFormat is known (valid) or not.
func (r RDFFormat) Valid() bool {
	return !(r <= RDFFormatUnknown || int(r) >= len(rdfFormatValues()))
}

//revive:disable:add-constant
func rdfFormatValues() [7]string {
	return [7]string{
		RDFFormatUnknown:  "UNKNOWN",
		RDFFormatTrig:     mediaTypeApplicationTrig,
		RDFFormatTurtle:   mediaTypeTextTurtle,
		RDFFormatRDFXML:   mediaTypeApplicationRDFXML,
		RDFFormatNTriples: mediaTypeApplicationNTriples,
		RDFFormatNQuads:   mediaTypeApplicationNQuads,
		RDFFormatJSONLD:   mediaTypeApplicationJSONLD,
	}
}

//revive:enable:add-constant

// String will return the string representation of the RDFFormat, which is the [MIME-type]
//
// [MIME-type]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
func (r RDFFormat) String() string {
	if !r.Valid() {
		return rdfFormatValues()[RDFFormatUnknown]
	}
	return rdfFormatValues()[r]
}

// helper function to get a string representation of the RDFFormat that [DatabaseAdminService.ExportData]
// and [DatabaseAdminService.ExportObfuscatedData] need to satisfy the Stardog API.
func (r RDFFormat) toExportFormat() (string, error) {
	switch r {
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

// GetRDFFormatFromExtension attempts to determine the RDFFormat from a given filepath.
func GetRDFFormatFromExtension(path string) (RDFFormat, error) {
	extension := strings.TrimPrefix(filepath.Ext(path), ".")
	switch extension {
	case "ttl":
		return RDFFormatTurtle, nil
	case "rdf", "rdfs", "xml", "owl":
		return RDFFormatRDFXML, nil
	case "trig":
		return RDFFormatTrig, nil
	case "jsonld", "json":
		return RDFFormatJSONLD, nil
	case "nq", "nquads":
		return RDFFormatNQuads, nil
	case "nt":
		return RDFFormatNTriples, nil
	default:
		return RDFFormatUnknown, fmt.Errorf("unable to determine the RDF Format from file: %s", path)
	}
}

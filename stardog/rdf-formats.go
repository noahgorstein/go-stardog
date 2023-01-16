package stardog

import (
	"fmt"
	"path/filepath"
	"strings"
)

type RDFFormat string

const (
	Trig     RDFFormat = "application/trig"
	Turtle   RDFFormat = "text/turtle"
	RDFXML   RDFFormat = "application/rdf+xml"
	NTriples RDFFormat = "application/n-triples"
	NQuads   RDFFormat = "application/n-quads"
	JSONLD   RDFFormat = "application/ld+json"
)

// GetRDFFormatFromExtension attempts to determine the RDFFormat from a given filepath.
func GetRDFFormatFromExtension(path string) (RDFFormat, error) {
	extension := strings.TrimPrefix(filepath.Ext(path), ".")
	switch extension {
	case "ttl":
		return Turtle, nil
	case "rdf", "xml":
		return RDFXML, nil
	case "trig":
		return Trig, nil
	case "jsonld":
		return JSONLD, nil
	case "nq":
		return NQuads, nil
	case "nt":
		return NTriples, nil
	default:
		return "", fmt.Errorf("unable to determine the RDF Format from file: %s", path)
	}
}

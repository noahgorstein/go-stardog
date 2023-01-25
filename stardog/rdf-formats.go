package stardog

import (
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

// String will return the string representation of the RDFFormat, which is the MIME-type
func (r RDFFormat) String() string {
	if !r.Valid() {
		return rdfFormatValues()[RDFFormatUnknown]
	}
	return rdfFormatValues()[r]
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

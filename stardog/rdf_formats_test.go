package stardog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRDFFormat_Valid(t *testing.T) {
	f := RDFFormat(100)
	if f.Valid() {
		t.Errorf("should be an invalid RDFFormat")
	}
	if f.String() != RDFFormatUnknown.String() {
		t.Errorf("RDFFormat string value should be unknown")
	}
}

func TestRDFFormat_GetRDFFormatFromExtension(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RDFFormat
	}{
		{name: "turtle", input: "file.ttl", want: RDFFormatTurtle},
		{name: "trig", input: "file.trig", want: RDFFormatTrig},
		{name: "rdfxml", input: "file.rdf", want: RDFFormatRDFXML},
		{name: "ntriples", input: "file.nt", want: RDFFormatNTriples},
		{name: "nquads", input: "file.nq", want: RDFFormatNQuads},
		{name: "jsonld", input: "file.jsonld", want: RDFFormatJSONLD},
	}

	for _, tc := range tests {
		got, err := GetRDFFormatFromExtension(tc.input)
		if err != nil {
			t.Errorf("RDFFormat.GetRDFFormatFromExtension unexpected failure: %v: ", err)
		}
		if !cmp.Equal(got, tc.want) {
			t.Errorf("RDFFormat.GetRDFFormatFromExtension failure: %s: expected: %v, got: %v", tc.name, tc.want, got)
		}
	}

	fileWithoutRDFExtension := "file.pdf"
	_, err := GetRDFFormatFromExtension(fileWithoutRDFExtension)
	if err == nil {
		t.Errorf("RDFFormat.GetRDFFormatFromExtension failure: %s should not have an extension that matches an RDF Format.", fileWithoutRDFExtension)
	}
}

func TestRDFFormat_toExportFormat(t *testing.T) {
	tests := []struct {
		name  string
		input RDFFormat
		want  string
	}{
		{name: "turtle", input: RDFFormatTurtle, want: "turtle"},
		{name: "trig", input: RDFFormatTrig, want: "trig"},
		{name: "rdfxml", input: RDFFormatRDFXML, want: "rdfxml"},
		{name: "ntriples", input: RDFFormatNTriples, want: "ntriples"},
		{name: "nquads", input: RDFFormatNQuads, want: "nquads"},
		{name: "jsonld", input: RDFFormatJSONLD, want: "jsonld"},
	}

	for _, tc := range tests {
		got, err := tc.input.toExportFormat()
		if err != nil {
			t.Errorf("RDFFormat.toExportFormat unexpected failure: %v: ", err)
		}
		if !cmp.Equal(got, tc.want) {
			t.Errorf("RDFFormat.toExportFormat failure: %s: expected: %v, got: %v", tc.name, tc.want, got)
		}
	}

	unknownRDFFormat := RDFFormatUnknown
	_, err := unknownRDFFormat.toExportFormat()
	if err == nil {
		t.Errorf("RDFFormat.toExportFormat failure: %s should have failed because this is not a known format", unknownRDFFormat)
	}
}

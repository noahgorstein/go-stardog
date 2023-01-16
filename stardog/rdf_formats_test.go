package stardog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_GetRDFFormatFromExtension(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RDFFormat
	}{
		{name: "turtle", input: "file.ttl", want: Turtle},
		{name: "trig", input: "file.trig", want: Trig},
		{name: "rdfxml", input: "file.rdf", want: RDFXML},
		{name: "ntriples", input: "file.nt", want: NTriples},
		{name: "nquads", input: "file.nq", want: NQuads},
		{name: "jsonld", input: "file.jsonld", want: JSONLD},
	}

	for _, tc := range tests {
		got, err := GetRDFFormatFromExtension(tc.input)
		if err != nil {
			t.Errorf("GetRDFFormatFromExtension unexpected failure: %v: ", err)
		}
		if !cmp.Equal(got, tc.want) {
			t.Errorf("GetRDFFormatFromExtension failure: %s: expected: %v, got: %v", tc.name, tc.want, got)
		}
	}

	fileWithoutRDFExtension := "file.pdf"
	_, err := GetRDFFormatFromExtension(fileWithoutRDFExtension)
	if err == nil {
		t.Errorf("GetRDFFormatFromExtension failure: %s should not have an extension that matches an RDF Format.", fileWithoutRDFExtension)
	}
}

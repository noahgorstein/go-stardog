package stardog

import "testing"

func Test_Compression_Valid(t *testing.T) {
	c := Compression(100)
	if c.Valid() {
		t.Errorf("should be an invalid Compression")
	}
	if c.String() != CompressionUnknown.String() {
		t.Errorf("Compression string value should be empty string")
	}
}

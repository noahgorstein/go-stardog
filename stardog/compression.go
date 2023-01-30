package stardog

// Data compression formats available in Stardog.
// The zero-value for Compression is CompressionUnknown
type Compression int

// All available compression formats in Stardog.
const (
	CompressionUnknown Compression = iota
	CompressionBZ2
	CompressionZIP
	CompressionGZIP
)

// compressionValues returns an array mapping each
// compressionValues to its string value
//
//revive:disable:add-constant
func compressionValues() [4]string {
	return [4]string{
		CompressionUnknown: "",
		CompressionBZ2:     "BZ2",
		CompressionZIP:     "ZIP",
		CompressionGZIP:    "GZIP",
	}
}

//revive:enable:add-constant

// Valid returns if a Compression is known (valid) or not.
func (c Compression) Valid() bool {
	return !(c <= CompressionUnknown || int(c) >= len(compressionValues()))
}

func (c Compression) String() string {
	if !c.Valid() {
		return compressionValues()[CompressionUnknown]
	}
	return compressionValues()[c]
}

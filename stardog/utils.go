package stardog

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

func newString(str string) *string {
  return &str
}

package stardog

// indexOf returns the index of the first occurrence of the target in the slice.
// If target is not found in the slice, -1 will be returned
func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

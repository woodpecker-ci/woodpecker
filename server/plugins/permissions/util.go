package permissions

// sliceToMap is a helper function to convert a string slice to a map.
func sliceToMap(s []string) map[string]bool {
	v := map[string]bool{}
	for _, ss := range s {
		if ss == "" {
			continue
		}
		v[ss] = true
	}
	return v
}

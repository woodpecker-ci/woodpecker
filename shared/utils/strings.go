package utils

// DedupStrings deduplicate string list, empty items are dropped
func DedupStrings(list []string) []string {
	m := make(map[string]struct{}, len(list))

	for i := range list {
		if s := list[i]; len(s) > 0 {
			m[list[i]] = struct{}{}
		}
	}

	newList := make([]string, 0, len(m))
	for k := range m {
		newList = append(newList, k)
	}
	return newList
}

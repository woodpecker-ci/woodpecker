package model

type ListOptions struct {
	All     bool
	Page    int
	PerPage int
}

func ApplyPagination[T any](d *ListOptions, slice []T) []T {
	if d.All {
		return slice
	}
	if d.PerPage*(d.Page-1) > len(slice) || d.PerPage*(d.Page) > len(slice) {
		return []T{}
	}
	return slice[d.PerPage*(d.Page-1) : d.PerPage*(d.Page)]
}

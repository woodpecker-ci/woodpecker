package woodpecker

import (
	"fmt"
	"net/url"
)

// ListOptions represents the options for the Woodpecker API pagination.
type ListOptions struct {
	Page    int
	PerPage int
}

// getURLQuery returns the query string for the ListOptions.
func (o ListOptions) getURLQuery() url.Values {
	query := make(url.Values)
	if o.Page > 0 {
		query.Add("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PerPage > 0 {
		query.Add("perPage", fmt.Sprintf("%d", o.PerPage))
	}

	return query
}

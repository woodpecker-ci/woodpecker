package woodpecker

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListOptions_getURLQuery(t *testing.T) {
	tests := []struct {
		name     string
		opts     ListOptions
		expected url.Values
	}{
		{
			name:     "no options",
			opts:     ListOptions{},
			expected: url.Values{},
		},
		{
			name:     "with page",
			opts:     ListOptions{Page: 2},
			expected: url.Values{"page": {"2"}},
		},
		{
			name:     "with per page",
			opts:     ListOptions{PerPage: 10},
			expected: url.Values{"perPage": {"10"}},
		},
		{
			name:     "with page and per page",
			opts:     ListOptions{Page: 3, PerPage: 20},
			expected: url.Values{"page": {"3"}, "perPage": {"20"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.opts.getURLQuery()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

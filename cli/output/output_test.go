package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputOptions(t *testing.T) {
		t.Parallel()

	testCases := []struct {
		in string
		out string
		opts []string
	}{
	{
		in: "output",
		out: "output",
	},
	{
		in: "output=a",
		out: "output",
		opts: []string{"a"},
	},
	{
		in: "output=",
		out: "output",
	},
	{
		in: "output=a,b",
		out: "output",
		opts: []string{"a", "b"},
	},
	}

	for _, tc := range testCases {
		out, opts := ParseOutputOptions(tc.in)
		assert.Equal(t, tc.out, out)
		assert.Equal(t, tc.opts, opts)
	}
}

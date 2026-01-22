package badges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Keep newlines, but clean up whitespace.
func TestStripXmlWhitespace(t *testing.T) {
	const mock = `<xml>
	<prop></prop>
<prop></prop><prop>
aaa
</prop>
	<prop>aaaa
	</prop>
<xml>  `
	const expected = `<xml><prop></prop><prop></prop><prop>aaa
</prop><prop>aaaa
	</prop><xml>`

	assert.Equal(t, stripXMLWhitespace(mock), expected)
}

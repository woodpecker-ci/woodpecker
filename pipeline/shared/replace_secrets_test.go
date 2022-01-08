package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretsReplacer(t *testing.T) {
	tc := []struct {
		log     string
		secrets []string
		expect  string
	}{{
		log:     "start log\ndone",
		secrets: []string{""},
		expect:  "start log\ndone",
	}, {
		log:     `this IS secret: password`,
		secrets: []string{"password", " IS "},
		expect:  `this ******** secret: ********`,
	}, {
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)",
		secrets: []string{"an\nmulti line secret!!"},
		expect:  "start log\ndone\nnow\n******** ;)",
	}}

	for _, c := range tc {
		rep := NewSecretsReplacer(c.secrets)
		result := rep.Replace(c.log)
		assert.EqualValues(t, c.expect, result)
	}
}

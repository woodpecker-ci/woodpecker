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
		expect:  "start log\ndone\nnow\n********\n******** ;)",
	}, {
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)",
		secrets: []string{"Test\nwith\n\ntwo new lines"},
		expect:  "start log\ndone\nnow\nan\nmulti line secret!! ;)",
	}, {
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)\nwith\ntwo\n\nnewlines",
		secrets: []string{"an\nmulti line secret!!", "two\n\nnewlines"},
		expect:  "start log\ndone\nnow\n********\n******** ;)\nwith\n********\n\n********",
	}}

	for _, c := range tc {
		rep := NewSecretsReplacer(c.secrets)
		result := rep.Replace(c.log)
		assert.EqualValues(t, c.expect, result)
	}
}

package yml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJSON(t *testing.T) {
	tests := []struct {
		yaml string
		json string
	}{{
		yaml: `- name: Jack
- name: Jill
`,
		json: `[{"name":"Jack"},{"name":"Jill"}]`,
	}, {
		yaml: `name: Jack`,
		json: `{"name":"Jack"}`,
	}, {
		yaml: `name: Jack
job: Butcher
`,
		json: `{"job":"Butcher","name":"Jack"}`,
	}, {
		yaml: `- name: Jack
  job: Butcher
- name: Jill
  job: Cook
  obj:
    empty: false
    data: |
      some data 123
      with new line
`,
		json: `[{"job":"Butcher","name":"Jack"},{"job":"Cook","name":"Jill","obj":{"data":"some data 123\nwith new line\n","empty":false}}]`,
	}, {
		yaml: `vars:
  - &node_image 'node:16-alpine'
  - &when_path
    # web source code
    - "web/**"
    - some

pipeline:
  deps:
    image: *node_image
    commands:
    - "cd web/"
    - yarn install
    when:
      path: *when_path
`,
		json: `{"pipeline":{"deps":{"commands":["cd web/","yarn install"],"image":"node:16-alpine","when":{"path":["web/**","some"]}}},"vars":["node:16-alpine",["web/**","some"]]}`,
	}}

	for _, tc := range tests {
		result, err := ToJSON([]byte(tc.yaml))
		assert.NoError(t, err)
		assert.EqualValues(t, tc.json, string(result))
	}
}

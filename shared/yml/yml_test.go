package yml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJSON(t *testing.T) {
	result, err := ToJSON([]byte(`- name: Jack
- name: Jill
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `[{"name":"Jack"},{"name":"Jill"}]`, string(result))

	result, err = ToJSON([]byte(`name: Jack`))
	assert.NoError(t, err)
	assert.EqualValues(t, `{"name":"Jack"}`, string(result))

	result, err = ToJSON([]byte(`name: Jack
job: Butcher
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `{"job":"Butcher","name":"Jack"}`, string(result))

	result, err = ToJSON([]byte(`- name: Jack
  job: Butcher
- name: Jill
  job: Cook
  obj:
    empty: false
    data: |
      some data 123
      with new line
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `[{"job":"Butcher","name":"Jack"},{"job":"Cook","name":"Jill","obj":{"data":"some data 123\nwith new line\n","empty":false}}]`, string(result))
}

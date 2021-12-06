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

	result, err = ToJSON([]byte(`name: Jack
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `{"name":"Jack"}`, string(result))

	result, err = ToJSON([]byte(`name: Jack
job: Butcher
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `{"name":"Jack","job":"Butcher"}`, string(result))

	result, err = ToJSON([]byte(`- name: Jack
  job: Butcher
- name: Jill
`))
	assert.NoError(t, err)
	assert.EqualValues(t, `[{"name":"Jack","job":"Butcher"},{"name":"Jill"}]`, string(result))
}

package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructCommand struct {
	Entrypoint Command `yaml:"entrypoint,flow,omitempty"`
	Command    Command `yaml:"command,flow,omitempty"`
}

func TestUnmarshalCommand(t *testing.T) {
	s := &StructCommand{}
	err := yaml.Unmarshal([]byte(`command: bash`), s)

	assert.Nil(t, err)
	assert.Equal(t, Command{"bash"}, s.Command)
	assert.Nil(t, s.Entrypoint)
	bytes, err := yaml.Marshal(s)
	assert.Nil(t, err)

	s2 := &StructCommand{}
	err = yaml.Unmarshal(bytes, s2)

	assert.Nil(t, err)
	assert.Equal(t, Command{"bash"}, s2.Command)
	assert.Nil(t, s2.Entrypoint)

	s3 := &StructCommand{}
	err = yaml.Unmarshal([]byte(`command:
    - echo AAA; echo "wow"
    - sleep 3s`), s3)
	assert.Nil(t, err)
	assert.Equal(t, Command{`echo AAA; echo "wow"`, `sleep 3s`}, s3.Command)

	s4 := &StructCommand{}
	err = yaml.Unmarshal([]byte(`command: echo AAA; echo "wow"`), s4)
	assert.Nil(t, err)
	assert.Equal(t, Command{`echo AAA; echo "wow"`}, s4.Command)
}

var sampleEmptyCommand = `{}`

func TestUnmarshalEmptyCommand(t *testing.T) {
	s := &StructCommand{}
	err := yaml.Unmarshal([]byte(sampleEmptyCommand), s)

	assert.Nil(t, err)
	assert.Nil(t, s.Command)

	bytes, err := yaml.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, "{}", strings.TrimSpace(string(bytes)))

	s2 := &StructCommand{}
	err = yaml.Unmarshal(bytes, s2)

	assert.Nil(t, err)
	assert.Nil(t, s2.Command)
}

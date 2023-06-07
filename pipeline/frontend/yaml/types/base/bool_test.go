package base

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBoolTrue(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Yaml bool type", func() {
		g.Describe("given a yaml file", func() {
			g.It("should unmarshal true", func() {
				in := []byte("true")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(true)
			})

			g.It("should unmarshal false", func() {
				in := []byte("false")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(false)
			})

			g.It("should unmarshal true when empty", func() {
				in := []byte("")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(true)
			})

			g.It("should throw error when invalid", func() {
				in := []byte("abc") // string value should fail parse
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				g.Assert(err).IsNotNil("expects error")
			})
		})
	})

	g.Describe("marshal", func() {
		g.It("marshal empty", func() {
			in := &BoolTrue{}
			out, err := yaml.Marshal(&in)
			g.Assert(err).IsNil("expect no error")
			assert.EqualValues(t, "true\n", string(out))
		})

		g.It("marshal true", func() {
			in := BoolTrue{value: true}
			out, err := yaml.Marshal(&in)
			g.Assert(err).IsNil("expect no error")
			assert.EqualValues(t, "true\n", string(out))
		})

		g.It("marshal false", func() {
			in := BoolTrue{value: false}
			out, err := yaml.Marshal(&in)
			g.Assert(err).IsNil("expect no error")
			assert.EqualValues(t, "false\n", string(out))
		})
	})
}

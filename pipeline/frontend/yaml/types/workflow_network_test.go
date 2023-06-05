package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalNetwork(t *testing.T) {
	testdata := []struct {
		from string
		want Network
	}{
		{
			from: "{ name: foo, driver: bar }",
			want: Network{
				Name:   "foo",
				Driver: "bar",
			},
		},
		{
			from: "{ name: foo, driver: bar, driver_opts: { baz: qux } }",
			want: Network{
				Name:   "foo",
				Driver: "bar",
				DriverOpts: map[string]string{
					"baz": "qux",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := Network{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got, "problem parsing network %q", test.from)
	}
}

func TestUnmarshalNetworks(t *testing.T) {
	testdata := []struct {
		from string
		want []*Network
	}{
		{
			from: "foo: { driver: bar }",
			want: []*Network{
				{
					Name:   "foo",
					Driver: "bar",
				},
			},
		},
		{
			from: "foo: { name: baz }",
			want: []*Network{
				{
					Name:   "baz",
					Driver: "bridge",
				},
			},
		},
		{
			from: "foo: { name: baz, driver: bar }",
			want: []*Network{
				{
					Name:   "baz",
					Driver: "bar",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := WorkflowNetworks{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.Networks, "problem parsing network %q", test.from)
	}
}

func TestUnmarshalNetworkErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}

	for _, test := range testdata {
		in := []byte(test)
		err := yaml.Unmarshal(in, new(WorkflowNetworks))
		assert.Error(t, err, "wanted error for networks %q", test)
	}
}

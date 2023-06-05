package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalNetwork(t *testing.T) {
	testdata := []struct {
		from string
		want WorkflowNetwork
	}{
		{
			from: "{ name: foo, driver: bar }",
			want: WorkflowNetwork{
				Name:   "foo",
				Driver: "bar",
			},
		},
		{
			from: "{ name: foo, driver: bar, driver_opts: { baz: qux } }",
			want: WorkflowNetwork{
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
		got := WorkflowNetwork{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got, "problem parsing network %q", test.from)
	}
}

func TestUnmarshalWorkflowNetworks(t *testing.T) {
	testdata := []struct {
		from string
		want []*WorkflowNetwork
	}{
		{
			from: "foo: { driver: bar }",
			want: []*WorkflowNetwork{
				{
					Name:   "foo",
					Driver: "bar",
				},
			},
		},
		{
			from: "foo: { name: baz }",
			want: []*WorkflowNetwork{
				{
					Name:   "baz",
					Driver: "bridge",
				},
			},
		},
		{
			from: "foo: { name: baz, driver: bar }",
			want: []*WorkflowNetwork{
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
		assert.EqualValues(t, test.want, got.WorkflowNetworks, "problem parsing network %q", test.from)
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

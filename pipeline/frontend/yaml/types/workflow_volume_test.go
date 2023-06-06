package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalVolume(t *testing.T) {
	testdata := []struct {
		from string
		want WorkflowVolume
	}{
		{
			from: "{ name: foo, driver: bar }",
			want: WorkflowVolume{
				Name:   "foo",
				Driver: "bar",
			},
		},
		{
			from: "{ name: foo, driver: bar, driver_opts: { baz: qux } }",
			want: WorkflowVolume{
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
		got := WorkflowVolume{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got, "problem parsing volume %q", test.from)
	}
}

func TestUnmarshalWorkflowVolumes(t *testing.T) {
	testdata := []struct {
		from string
		want []*WorkflowVolume
	}{
		{
			from: "foo: { driver: bar }",
			want: []*WorkflowVolume{
				{
					Name:   "foo",
					Driver: "bar",
				},
			},
		},
		{
			from: "foo: { name: baz }",
			want: []*WorkflowVolume{
				{
					Name:   "baz",
					Driver: "local",
				},
			},
		},
		{
			from: "foo: { name: baz, driver: bar }",
			want: []*WorkflowVolume{
				{
					Name:   "baz",
					Driver: "bar",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := WorkflowVolumes{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.WorkflowVolumes, "problem parsing volumes %q", test.from)
	}
}

func TestUnmarshalVolumesErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}

	for _, test := range testdata {
		in := []byte(test)
		err := yaml.Unmarshal(in, new(WorkflowVolumes))
		assert.Error(t, err, "wanted error for volumes %q", test)
	}
}

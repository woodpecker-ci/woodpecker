package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalVolume(t *testing.T) {
	testdata := []struct {
		from string
		want Volume
	}{
		{
			from: "{ name: foo, driver: bar }",
			want: Volume{
				Name:   "foo",
				Driver: "bar",
			},
		},
		{
			from: "{ name: foo, driver: bar, driver_opts: { baz: qux } }",
			want: Volume{
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
		got := Volume{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got, "problem parsing volume %q", test.from)
	}
}

func TestUnmarshalVolumes(t *testing.T) {
	testdata := []struct {
		from string
		want []*Volume
	}{
		{
			from: "foo: { driver: bar }",
			want: []*Volume{
				{
					Name:   "foo",
					Driver: "bar",
				},
			},
		},
		{
			from: "foo: { name: baz }",
			want: []*Volume{
				{
					Name:   "baz",
					Driver: "local",
				},
			},
		},
		{
			from: "foo: { name: baz, driver: bar }",
			want: []*Volume{
				{
					Name:   "baz",
					Driver: "bar",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := Volumes{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.Volumes, "problem parsing volumes %q", test.from)
	}
}

func TestUnmarshalVolumesErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}

	for _, test := range testdata {
		in := []byte(test)
		err := yaml.Unmarshal(in, new(Volumes))
		assert.Error(t, err, "wanted error for volumes %q", test)
	}
}

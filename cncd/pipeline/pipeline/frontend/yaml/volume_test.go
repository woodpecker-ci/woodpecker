package yaml

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
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
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(test.want, got) {
			t.Errorf("problem parsing volume %q", test.from)
			pretty.Ldiff(t, test.want, got)
		}
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
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(test.want, got.Volumes) {
			t.Errorf("problem parsing volumes %q", test.from)
			pretty.Ldiff(t, test.want, got.Volumes)
		}
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
		if err == nil {
			t.Errorf("wanted error for volumes %q", test)
		}
	}
}

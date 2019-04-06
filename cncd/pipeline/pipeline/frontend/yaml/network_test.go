package yaml

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"gopkg.in/yaml.v2"
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
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(test.want, got) {
			t.Errorf("problem parsing network %q", test.from)
			pretty.Ldiff(t, test.want, got)
		}
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
		got := Networks{}
		err := yaml.Unmarshal(in, &got)
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(test.want, got.Networks) {
			t.Errorf("problem parsing network %q", test.from)
			pretty.Ldiff(t, test.want, got.Networks)
		}
	}
}

func TestUnmarshalNetworkErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}

	for _, test := range testdata {
		in := []byte(test)
		err := yaml.Unmarshal(in, new(Networks))
		if err == nil {
			t.Errorf("wanted error for networks %q", test)
		}
	}
}

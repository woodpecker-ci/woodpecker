package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalSecrets(t *testing.T) {
	testdata := []struct {
		from string
		want []*Secret
	}{
		{
			from: "[ mysql_username, mysql_password]",
			want: []*Secret{
				{
					Source: "mysql_username",
					Target: "mysql_username",
				},
				{
					Source: "mysql_password",
					Target: "mysql_password",
				},
			},
		},
		{
			from: "[ { source: mysql_prod_username, target: mysql_username } ]",
			want: []*Secret{
				{
					Source: "mysql_prod_username",
					Target: "mysql_username",
				},
			},
		},
		{
			from: "[ { source: mysql_prod_username, target: mysql_username }, { source: redis_username, target: redis_username } ]",
			want: []*Secret{
				{
					Source: "mysql_prod_username",
					Target: "mysql_username",
				},
				{
					Source: "redis_username",
					Target: "redis_username",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := Secrets{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.Secrets, "problem parsing secrets %q", test.from)
	}
}

package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestOrgs(t *testing.T) {
	o := NewOrgs([]string{"woodpecker-ci"})
	assert.True(t, o.IsConfigured)
	assert.True(t, o.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.False(t, o.IsMember([]*model.Team{{Login: "not-woodpecker-ci"}}))
	empty := NewOrgs([]string{})
	assert.False(t, empty.IsConfigured)
	assert.False(t, empty.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.False(t, empty.IsMember([]*model.Team{{Login: "not-woodpecker-ci"}}))
}

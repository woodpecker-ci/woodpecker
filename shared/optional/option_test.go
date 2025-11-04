// Copyright 2025 Woodpecker Authors.
// Copyright 2024 The Gitea Authors.
//
// Licensed under the MIT License.

package optional_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

func TestOption(t *testing.T) {
	var uninitialized optional.Option[int]
	assert.False(t, uninitialized.Has())
	assert.Equal(t, int(0), uninitialized.Value())
	assert.Equal(t, int(1), uninitialized.ValueOrDefault(1))

	none := optional.None[int]()
	assert.False(t, none.Has())
	assert.Equal(t, int(0), none.Value())
	assert.Equal(t, int(1), none.ValueOrDefault(1))

	some := optional.Some[int](1)
	assert.True(t, some.Has())
	assert.Equal(t, int(1), some.Value())
	assert.Equal(t, int(1), some.ValueOrDefault(2))

	var ptr *int
	assert.False(t, optional.FromPtr(ptr).Has())

	var boolPtr *bool
	assert.Equal(t, boolPtr, optional.None[bool]().ToPtr())

	boolPtr = optional.Some[bool](false).ToPtr()
	assert.Equal(t, toPtr(false), boolPtr)

	opt1 := optional.FromPtr(toPtr(1))
	assert.True(t, opt1.Has())
	assert.Equal(t, int(1), opt1.Value())

	assert.False(t, optional.FromNonDefault("").Has())

	opt2 := optional.FromNonDefault("test")
	assert.True(t, opt2.Has())
	assert.Equal(t, "test", opt2.Value())

	assert.False(t, optional.FromNonDefault(0).Has())

	opt3 := optional.FromNonDefault(1)
	assert.True(t, opt3.Has())
	assert.Equal(t, int(1), opt3.Value())
}

func TestExtractValue(t *testing.T) {
	val, ok := optional.ExtractValue("aaaa")
	assert.False(t, ok)
	assert.Nil(t, val)

	val, ok = optional.ExtractValue(optional.Some("aaaa"))
	assert.True(t, ok)
	if assert.NotNil(t, val) {
		val, ok := val.(string)
		assert.True(t, ok)
		assert.EqualValues(t, "aaaa", val)
	}

	val, ok = optional.ExtractValue(optional.None[float64]())
	assert.True(t, ok)
	assert.Nil(t, val)

	val, ok = optional.ExtractValue(&fakeHas{})
	assert.False(t, ok)
	assert.Nil(t, val)

	wrongType := make(fakeHas2, 0, 1)
	val, ok = optional.ExtractValue(wrongType)
	assert.False(t, ok)
	assert.Nil(t, val)
}

func toPtr[T any](val T) *T {
	return &val
}

type fakeHas struct{}

func (fakeHas) Has() bool {
	return true
}

type fakeHas2 []string

func (fakeHas2) Has() bool {
	return true
}

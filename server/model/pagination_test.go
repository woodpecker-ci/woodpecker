package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyPagination(t *testing.T) {
	example := []int{
		0, 1, 2,
	}

	assert.Equal(t, ApplyPagination(&ListOptions{All: true}, example), example)
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 1, PerPage: 1}, example), []int{0})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 2, PerPage: 2}, example), []int{2})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 3, PerPage: 1}, example), []int{2})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 4, PerPage: 1}, example), []int{})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 5, PerPage: 1}, example), []int{})
}

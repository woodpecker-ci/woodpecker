package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginate(t *testing.T) {
	apiExec := 0
	apiMock := func(page int) []int {
		apiExec++
		switch page {
		case 0, 1:
			return []int{11, 12, 13}
		case 2:
			return []int{21, 22, 23}
		case 3:
			return []int{31, 32}
		default:
			return []int{}
		}
	}

	result, _ := Paginate(func(page int) ([]int, error) {
		return apiMock(page), nil
	})

	assert.EqualValues(t, 3, apiExec)
	if assert.Len(t, result, 8) {
		assert.EqualValues(t, []int{11, 12, 13, 21, 22, 23, 31, 32}, result)
	}
}

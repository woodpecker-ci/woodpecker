// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryOK(t *testing.T) {
	expected := 42
	result, retries, reasons, err := Retry(func() (int, bool, string, error) {
		return expected, false, "", nil
	}, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, retries)
	assert.Equal(t, expected, result)
	assert.Contains(t, reasons, "Retry reasons")

	i := 0
	result, retries, reasons, err = Retry(func() (int, bool, string, error) {
		i++
		if i < 3 {
			return 0, true, "fail", nil
		}
		return expected, false, "", nil
	}, 5)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.EqualValues(t, 2, retries)
	assert.Equal(t, "Retry reasons [fail fail]", reasons)
}

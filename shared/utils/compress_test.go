package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

func TestZStd(t *testing.T) {
	bytes := []byte(`WARNING: buildx: git was not found in the system. Current commit information was not captured by the build`)
	cBytes := utils.ZStdCompress(bytes)
	assert.Len(t, cBytes, 119)
	newBytes, err := utils.ZStdDecompress(cBytes)
	assert.NoError(t, err)
	assert.Len(t, newBytes, 106)
	assert.EqualValues(t, bytes, newBytes)
}

package woodpeckergo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	woodpeckergo "go.woodpecker-ci.org/woodpecker/v2/woodpecker-go"
)

func TestClient(t *testing.T) {
	client, err := woodpeckergo.New("http://localhost:8080")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

package migration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

func TestCopy(t *testing.T) {
	oldConfig := createSQLiteDB(t)
	defer func() {
		_ = os.Remove(oldConfig)
	}()

	srcEngine, err := xorm.NewEngine("sqlite3", oldConfig)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer srcEngine.Close()
	destEngine, _ := xorm.NewEngine("sqlite3", ":memory:")
	defer destEngine.Close()

	err = Copy(context.Background(), srcEngine, destEngine)
	assert.NoError(t, err)
}

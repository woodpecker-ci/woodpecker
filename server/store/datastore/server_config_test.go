package datastore

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestServerConfigGetSet(t *testing.T) {
	store, closer := newTestStore(t, new(model.ServerConfig))
	defer closer()

	serverConfig := &model.ServerConfig{
		Key:   "test",
		Value: "wonderland",
	}
	if err := store.ServerConfigSet(serverConfig.Key, serverConfig.Value); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	value, err := store.ServerConfigGet(serverConfig.Key)
	if err != nil {
		t.Errorf("Unexpected error: delete secret: %s", err)
		return
	}

	if value != serverConfig.Value {
		t.Errorf("Want server-config value %s, got %s", serverConfig.Value, value)
		return
	}
}

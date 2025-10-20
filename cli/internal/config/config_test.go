package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigMerge(t *testing.T) {
	config := &Config{
		ServerURL: "http://localhost:8080",
		Token:     "1234567890",
		LogLevel:  "debug",
	}

	configFromFile := &Config{
		ServerURL: "https://ci.woodpecker-ci.org",
		Token:     "",
		LogLevel:  "info",
	}

	config.MergeIfNotSet(configFromFile)

	assert.Equal(t, config.ServerURL, "http://localhost:8080")
	assert.Equal(t, config.Token, "1234567890")
	assert.Equal(t, config.LogLevel, "debug")
}

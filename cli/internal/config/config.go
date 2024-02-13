package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
	LogLevel  string `json:"log_level"`
}

func Load(configPath string) (*Config, error) {
	config, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		return nil, nil
	}

	var c *Config
	err = json.Unmarshal(config, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Save(configPath string, c *Config) error {
	config, err := json.Marshal(c)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0x755)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, config, 0x644)
	if err != nil {
		return err
	}

	return nil
}

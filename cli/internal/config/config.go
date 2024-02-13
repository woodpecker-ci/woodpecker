package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
	LogLevel  string `json:"log_level"`
}

func Load(c *cli.Context) error {
	// If the command is setup, we don't need to load the config
	if firstArg := c.Args().First(); firstArg == "setup" {
		return nil
	}

	// If the server URL and token are set, we don't need to load the config
	if c.IsSet("server-url") && c.IsSet("token") {
		return nil
	}

	config, err := Get(c.String("config"))
	if err != nil {
		return err
	}

	if config == nil {
		log.Info().Msg("The woodpecker-cli is not setup yet. Please run `woodpecker-cli setup`")
		return errors.New("woodpecker-cli is not setup")
	}

	if !c.IsSet("server-url") {
		err = c.Set("server-url", config.ServerURL)
		if err != nil {
			return err
		}
	}

	if !c.IsSet("token") {
		err = c.Set("token", config.Token)
		if err != nil {
			return err
		}
	}

	if !c.IsSet("log-level") {
		err = c.Set("log-level", config.LogLevel)
		if err != nil {
			return err
		}
	}

	return nil
}

func Get(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		return nil, nil
	}

	var c *Config
	err = json.Unmarshal(content, c)
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

package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/adrg/xdg"
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

	if !c.IsSet("server") {
		err = c.Set("server", config.ServerURL)
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

func getConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	configPath, err := xdg.ConfigFile("woodpecker/config.json")
	if err != nil {
		return "", err
	}

	return configPath, nil
}

func Get(_configPath string) (*Config, error) {
	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Debug().Err(err).Msg("Failed to read the config file")
		return nil, err
	} else if err != nil && os.IsNotExist(err) {
		log.Debug().Msg("The config file does not exist")
		return nil, nil
	}

	c := &Config{}
	err = json.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Save(_configPath string, c *Config) error {
	config, err := json.Marshal(c)
	if err != nil {
		return err
	}

	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, config, 0o600)
	if err != nil {
		return err
	}

	return nil
}

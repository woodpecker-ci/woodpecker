package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"-"`
	LogLevel  string `json:"log_level"`
}

var ErrNotSetup = errors.New("woodpecker-cli is not setup")

func Load(c *cli.Context) error {
	// If the command is setup, we don't need to load the config
	if firstArg := c.Args().First(); firstArg == "setup" {
		return nil
	}

	config, err := Get(c, c.String("config"))
	if err != nil {
		return err
	}

	if config == nil && !c.IsSet("server-url") && !c.IsSet("token") {
		log.Info().Msg("The woodpecker-cli is not yet set up. Please run `woodpecker-cli setup`")
		return ErrNotSetup
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

func Get(ctx *cli.Context, _configPath string) (*Config, error) {
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

	// load token from keyring
	service := ctx.App.Name
	secret, err := keyring.Get(service, c.ServerURL)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		c.Token = secret
	}

	return c, nil
}

func Save(ctx *cli.Context, _configPath string, c *Config) error {
	config, err := json.Marshal(c)
	if err != nil {
		return err
	}

	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return err
	}

	// save token to keyring
	service := ctx.App.Name
	err = keyring.Set(service, c.ServerURL, c.Token)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, config, 0o600)
	if err != nil {
		return err
	}

	return nil
}

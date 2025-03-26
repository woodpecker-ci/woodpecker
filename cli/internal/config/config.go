package config

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"github.com/zalando/go-keyring"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"-"`
	LogLevel  string `json:"log_level"`
}

func (c *Config) MergeIfNotSet(c2 *Config) {
	if c.ServerURL == "" {
		c.ServerURL = c2.ServerURL
	}
	if c.Token == "" {
		c.Token = c2.Token
	}
	if c.LogLevel == "" {
		c.LogLevel = c2.LogLevel
	}
}

var skipSetupForCommands = []string{"setup", "help", "h", "version", "update", "lint", "exec", ""}

func Load(ctx context.Context, c *cli.Command) error {
	if firstArg := c.Args().First(); slices.Contains(skipSetupForCommands, firstArg) {
		return nil
	}

	config, err := Get(ctx, c, c.String("config"))
	if err != nil {
		return err
	}

	if config.ServerURL == "" || config.Token == "" {
		log.Info().Msg("woodpecker-cli is not set up, run `woodpecker-cli setup` or provide required environment variables/flags")
		return errors.New("woodpecker-cli is not configured")
	}

	err = c.Set("server", config.ServerURL)
	if err != nil {
		return err
	}

	err = c.Set("token", config.Token)
	if err != nil {
		return err
	}

	err = c.Set("log-level", config.LogLevel)
	if err != nil {
		return err
	}

	log.Debug().Any("config", config).Msg("loaded config")

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

func Get(_ context.Context, c *cli.Command, _configPath string) (*Config, error) {
	conf := &Config{
		LogLevel:  c.String("log-level"),
		Token:     c.String("token"),
		ServerURL: c.String("server"),
	}

	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("configPath", configPath).Msg("checking for config file")

	content, err := os.ReadFile(configPath)
	switch {
	case err != nil && !os.IsNotExist(err):
		log.Debug().Err(err).Msg("failed to read the config file")
		return nil, err

	case err != nil && os.IsNotExist(err):
		log.Debug().Msg("config file does not exist")

	default:
		configFromFile := &Config{}
		err = json.Unmarshal(content, configFromFile)
		if err != nil {
			return nil, err
		}
		conf.MergeIfNotSet(configFromFile)
		log.Debug().Msg("loaded config from file")
	}

	// if server or token are explicitly set, use them
	if c.IsSet("server") || c.IsSet("token") {
		return conf, nil
	}

	// load token from keyring
	service := c.Root().Name
	secret, err := keyring.Get(service, conf.ServerURL)
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		log.Warn().Msg("keyring is not supported on this platform")
		return conf, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		log.Warn().Msg("token not found in keyring")
		return conf, nil
	}
	conf.Token = secret

	return conf, nil
}

func Save(_ context.Context, c *cli.Command, _configPath string, conf *Config) error {
	config, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return err
	}

	// save token to keyring
	service := c.Root().Name
	err = keyring.Set(service, conf.ServerURL, conf.Token)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, config, 0o600)
	if err != nil {
		return err
	}

	return nil
}

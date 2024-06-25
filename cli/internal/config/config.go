package config

import (
	"encoding/json"
	"errors"
	"os"
	"slices"

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

func Load(c *cli.Context) error {
	if firstArg := c.Args().First(); slices.Contains(skipSetupForCommands, firstArg) {
		return nil
	}

	config, err := Get(c, c.String("config"))
	if err != nil {
		return err
	}

	if config.ServerURL == "" || config.Token == "" {
		log.Info().Msg("The woodpecker-cli is not yet set up. Please run `woodpecker-cli setup` or provide the required environment variables / flags.")
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

	log.Debug().Any("config", config).Msg("Loaded config")

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
	c := &Config{
		LogLevel:  ctx.String("log-level"),
		Token:     ctx.String("token"),
		ServerURL: ctx.String("server"),
	}

	configPath, err := getConfigPath(_configPath)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("configPath", configPath).Msg("Checking for config file")

	content, err := os.ReadFile(configPath)
	switch {
	case err != nil && !os.IsNotExist(err):
		log.Debug().Err(err).Msg("Failed to read the config file")
		return nil, err

	case err != nil && os.IsNotExist(err):
		log.Debug().Msg("The config file does not exist")

	default:
		configFromFile := &Config{}
		err = json.Unmarshal(content, configFromFile)
		if err != nil {
			return nil, err
		}
		c.MergeIfNotSet(configFromFile)
		log.Debug().Msg("Loaded config from file")
	}

	// if server or token are explicitly set, use them
	if ctx.IsSet("server") || ctx.IsSet("token") {
		return c, nil
	}

	// load token from keyring
	service := ctx.App.Name
	secret, err := keyring.Get(service, c.ServerURL)
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		log.Warn().Msg("Keyring is not supported on this platform")
		return c, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		log.Warn().Msg("Token not found in keyring")
		return c, nil
	}
	c.Token = secret

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

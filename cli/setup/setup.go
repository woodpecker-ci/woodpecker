package setup

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal/config"
)

// Command exports the setup command.
var Command = &cli.Command{
	Name:  "setup",
	Usage: "setup the woodpecker-cli for the first time",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "server-url",
			Usage: "The URL of the woodpecker server",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "The token to authenticate with the woodpecker server",
		},
	},
	Action: setup,
}

func setup(c *cli.Context) error {
	_config, err := config.Get(c.String("config"))
	if err != nil {
		return err
	} else if _config != nil {
		log.Warn().Msg("The woodpecker-cli is already setup")
		return nil
	}

	// TODO: prompt for server URL
	serverURL := c.String("server-url")

	if serverURL == "" {
		log.Info().Msg("Please enter the URL of the woodpecker server like https://ci.woodpecker-ci.org")
		return errors.New("server URL is required")
	}

	if !strings.Contains(serverURL, "://") {
		serverURL = "https://" + serverURL
	}

	token := c.String("token")
	if token == "" {
		// TODO: wait for enter before opening the browser

		token, err = receiveTokenFromUI(c.Context, serverURL)
		if err != nil {
			return err
		}

		if token == "" {
			return errors.New("no token received from the UI")
		}
	}

	err = config.Save(c.String("config"), &config.Config{
		ServerURL: serverURL,
		Token:     token,
		LogLevel:  "info",
	})
	if err != nil {
		return err
	}

	log.Info().Msg("The woodpecker-cli has been successfully setup")

	return nil
}

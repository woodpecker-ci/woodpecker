package setup

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal/config"
	"go.woodpecker-ci.org/woodpecker/v2/cli/setup/ui"
)

// Command exports the setup command.
var Command = &cli.Command{
	Name:      "setup",
	Usage:     "setup the woodpecker-cli for the first time",
	ArgsUsage: "[server]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "server",
			Usage: "The URL of the woodpecker server",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "The token to authenticate with the woodpecker server",
		},
	},
	Action: setup,
}

func setup(ctx context.Context, c *cli.Command) error {
	_config, err := config.Get(ctx, c, c.String("config"))
	if err != nil {
		return err
	} else if _config != nil {
		setupAgain, err := ui.Confirm("The woodpecker-cli was already configured. Do you want to configure it again?")
		if err != nil {
			return err
		}

		if !setupAgain {
			log.Info().Msg("Configuration skipped")
			return nil
		}
	}

	serverURL := c.String("server")
	if serverURL == "" {
		serverURL = c.Args().First()
	}

	if serverURL == "" {
		serverURL, err = ui.Ask("Enter the URL of the woodpecker server", "https://ci.woodpecker-ci.org", true)
		if err != nil {
			return err
		}

		if serverURL == "" {
			return errors.New("server URL cannot be empty")
		}
	}

	if !strings.Contains(serverURL, "://") {
		serverURL = "https://" + serverURL
	}

	token := c.String("token")
	if token == "" {
		token, err = receiveTokenFromUI(ctx, serverURL)
		if err != nil {
			return err
		}

		if token == "" {
			return errors.New("no token received from the UI")
		}
	}

	err = config.Save(ctx, c, c.String("config"), &config.Config{
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

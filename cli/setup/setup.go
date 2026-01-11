package setup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal/config"
	"go.woodpecker-ci.org/woodpecker/v3/cli/setup/ui"
)

// Command exports the setup command.
var Command = &cli.Command{
	Name:      "setup",
	Usage:     "setup the woodpecker-cli for the first time",
	ArgsUsage: "[server]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "server",
			Usage: "URL of the woodpecker server",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "token to authenticate with the woodpecker server",
		},
		&cli.StringFlag{
			Name:    "context",
			Aliases: []string{"ctx"},
			Usage:   "name for the context (defaults to 'default')",
		},
	},
	Action: setup,
}

func setup(ctx context.Context, c *cli.Command) error {
	contextName := c.String("context")
	if contextName == "" {
		contextName = "default"
	}

	// Check if context already exists
	contexts, err := config.LoadContexts()
	if err != nil {
		return err
	}

	if existingCtx, exists := contexts.Contexts[contextName]; exists {
		setupAgain, err := ui.Confirm(fmt.Sprintf("Context '%s' already exists (server: %s). Do you want to reconfigure it?", contextName, existingCtx.ServerURL))
		if err != nil {
			return err
		}

		if !setupAgain {
			log.Info().Msg("configuration skipped")
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

	// Save as context
	err = config.AddOrUpdateContext(c, contextName, serverURL, token, "info", true)
	if err != nil {
		return err
	}

	log.Info().Msgf("Context '%s' has been successfully created and set as current", contextName)

	return nil
}

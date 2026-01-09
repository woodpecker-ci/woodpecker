// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal/config"
	"go.woodpecker-ci.org/woodpecker/v3/cli/output"
)

// Command exports the context command set.
var Command = &cli.Command{
	Name:    "context",
	Aliases: []string{"ctx"},
	Usage:   "manage contexts",
	Action:  listContexts,
	Commands: []*cli.Command{
		listCommand,
		useCommand,
		deleteCommand,
		renameCommand,
	},
}

var listCommand = &cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "list all contexts",
	Action:  listContexts,
}

var useCommand = &cli.Command{
	Name:      "use",
	Usage:     "set the current context",
	ArgsUsage: "<context-name>",
	Action:    useContext,
}

var deleteCommand = &cli.Command{
	Name:      "delete",
	Aliases:   []string{"rm"},
	Usage:     "delete a context",
	ArgsUsage: "<context-name>",
	Action:    deleteContext,
}

var renameCommand = &cli.Command{
	Name:      "rename",
	Usage:     "rename a context",
	ArgsUsage: "<old-name> <new-name>",
	Action:    renameContext,
}

func listContexts(_ context.Context, c *cli.Command) error {
	contexts, err := config.LoadContexts()
	if err != nil {
		return err
	}

	if len(contexts.Contexts) == 0 {
		fmt.Println("No contexts found. Run 'woodpecker-cli setup' to create one.")
		return nil
	}

	_, outOpt := output.ParseOutputOptions(c.String("output"))
	out := os.Stdout
	noHeader := c.Bool("output-no-headers")
	table := output.NewTable(out)

	// Add custom field mapping
	table.AddFieldFn("Name", func(obj any) string {
		c, ok := obj.(config.Context)
		if !ok {
			return "???"
		}

		if contexts.CurrentContext == c.Name {
			return c.Name + " *"
		}

		return c.Name
	})
	table.AddFieldAlias("ServerURL", "Server URL")
	table.AddFieldAlias("LogLevel", "Log Level")
	table.AddFieldAlias("Name", "Name (selected)")

	cols := []string{"Name (selected)", "Server URL"}

	if len(outOpt) > 0 {
		cols = outOpt
	}
	if !noHeader {
		table.WriteHeader(cols)
	}
	for _, c := range contexts.Contexts {
		if err := table.Write(cols, c); err != nil {
			return err
		}
	}

	return table.Flush()
}

func useContext(_ context.Context, c *cli.Command) error {
	contextName := c.Args().First()
	if contextName == "" {
		return fmt.Errorf("context name is required")
	}

	err := config.SetCurrentContext(contextName)
	if err != nil {
		return err
	}

	log.Info().Msgf("Switched to context '%s'", contextName)
	return nil
}

func deleteContext(_ context.Context, c *cli.Command) error {
	contextName := c.Args().First()
	if contextName == "" {
		return fmt.Errorf("context name is required")
	}

	err := config.DeleteContext(c, contextName)
	if err != nil {
		return err
	}

	log.Info().Msgf("Context '%s' deleted", contextName)
	return nil
}

func renameContext(_ context.Context, c *cli.Command) error {
	if c.Args().Len() < 2 {
		return fmt.Errorf("both old name and new name are required")
	}

	oldName := c.Args().Get(0)
	newName := c.Args().Get(1)

	err := config.RenameContext(oldName, newName)
	if err != nil {
		return err
	}

	log.Info().Msgf("Context renamed from '%s' to '%s'", oldName, newName)
	return nil
}

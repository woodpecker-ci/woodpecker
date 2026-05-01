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

package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"github.com/zalando/go-keyring"
)

// Context represents a single CLI context with its connection details.
type Context struct {
	Name      string `json:"name"`
	ServerURL string `json:"server_url"`
	LogLevel  string `json:"log_level,omitempty"`
}

// Contexts holds all contexts and tracks the current active one.
type Contexts struct {
	CurrentContext string             `json:"current_context"`
	Contexts       map[string]Context `json:"contexts"`
}

func getContextsPath() (string, error) {
	configPath, err := xdg.ConfigFile("woodpecker/contexts.json")
	if err != nil {
		return "", err
	}
	return configPath, nil
}

// LoadContexts loads all contexts from the contexts file.
func LoadContexts() (*Contexts, error) {
	contextsPath, err := getContextsPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(contextsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Contexts{
				Contexts: make(map[string]Context),
			}, nil
		}
		return nil, err
	}

	var contexts Contexts
	err = json.Unmarshal(content, &contexts)
	if err != nil {
		return nil, err
	}

	if contexts.Contexts == nil {
		contexts.Contexts = make(map[string]Context)
	}

	return &contexts, nil
}

// SaveContexts saves all contexts to the contexts file.
func SaveContexts(contexts *Contexts) error {
	data, err := json.MarshalIndent(contexts, "", "  ")
	if err != nil {
		return err
	}

	contextsPath, err := getContextsPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(contextsPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(contextsPath, data, 0o600)
}

// GetCurrentContext returns the current active context.
func GetCurrentContext(ctx context.Context, c *cli.Command) (*Config, error) {
	contexts, err := LoadContexts()
	if err != nil {
		return nil, err
	}

	if contexts.CurrentContext == "" {
		return nil, errors.New("no context is currently set")
	}

	context, exists := contexts.Contexts[contexts.CurrentContext]
	if !exists {
		return nil, fmt.Errorf("current context '%s' not found", contexts.CurrentContext)
	}

	return GetContextConfig(c, &context)
}

// GetContextConfig loads the config for a specific context including the token from keyring.
func GetContextConfig(c *cli.Command, ctx *Context) (*Config, error) {
	conf := &Config{
		ServerURL: ctx.ServerURL,
		LogLevel:  ctx.LogLevel,
	}

	// Load token from keyring
	service := c.Root().Name
	secret, err := keyring.Get(service, ctx.ServerURL)
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		log.Warn().Msg("keyring is not supported on this platform")
		return conf, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("token not found in keyring for context '%s'", ctx.Name)
	}
	if err != nil {
		return nil, err
	}

	conf.Token = secret
	return conf, nil
}

// AddOrUpdateContext adds or updates a context and optionally sets it as current.
func AddOrUpdateContext(c *cli.Command, name, serverURL, token, logLevel string, setCurrent bool) error {
	contexts, err := LoadContexts()
	if err != nil {
		return err
	}

	contexts.Contexts[name] = Context{
		Name:      name,
		ServerURL: serverURL,
		LogLevel:  logLevel,
	}

	if setCurrent || contexts.CurrentContext == "" {
		contexts.CurrentContext = name
	}

	// Save token to keyring
	service := c.Root().Name
	err = keyring.Set(service, serverURL, token)
	if err != nil {
		return err
	}

	return SaveContexts(contexts)
}

// DeleteContext removes a context.
func DeleteContext(c *cli.Command, name string) error {
	contexts, err := LoadContexts()
	if err != nil {
		return err
	}

	context, exists := contexts.Contexts[name]
	if !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	// Try to delete token from keyring
	service := c.Root().Name
	err = keyring.Delete(service, context.ServerURL)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		log.Warn().Err(err).Msg("failed to delete token from keyring")
	}

	delete(contexts.Contexts, name)

	// If we deleted the current context, unset it
	if contexts.CurrentContext == name {
		contexts.CurrentContext = ""
	}

	return SaveContexts(contexts)
}

// SetCurrentContext sets the current active context.
func SetCurrentContext(name string) error {
	contexts, err := LoadContexts()
	if err != nil {
		return err
	}

	if _, exists := contexts.Contexts[name]; !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	contexts.CurrentContext = name
	return SaveContexts(contexts)
}

// RenameContext renames an existing context.
func RenameContext(oldName, newName string) error {
	contexts, err := LoadContexts()
	if err != nil {
		return err
	}

	context, exists := contexts.Contexts[oldName]
	if !exists {
		return fmt.Errorf("context '%s' not found", oldName)
	}

	if _, exists := contexts.Contexts[newName]; exists {
		return fmt.Errorf("context '%s' already exists", newName)
	}

	// Update the name in the context
	context.Name = newName
	contexts.Contexts[newName] = context
	delete(contexts.Contexts, oldName)

	// Update current context if necessary
	if contexts.CurrentContext == oldName {
		contexts.CurrentContext = newName
	}

	return SaveContexts(contexts)
}

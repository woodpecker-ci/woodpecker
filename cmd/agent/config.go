// Copyright 2022 Woodpecker Authors
// Copyright 2019 Laszlo Fogas
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

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type AgentConfig struct {
	AgentID int64 `json:"agent_id"`
}

const defaultAgentIDValue = int64(-1)

func readAgentConfig(agentConfigPath string) AgentConfig {
	conf := AgentConfig{
		AgentID: defaultAgentIDValue,
	}

	rawAgentConf, err := os.ReadFile(agentConfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Info().Msgf("no agent config found at '%s', start with defaults", agentConfigPath)
		} else {
			log.Error().Err(err).Msgf("could not open agent config at '%s'", agentConfigPath)
		}
		return conf
	}
	if strings.TrimSpace(string(rawAgentConf)) == "" {
		return conf
	}

	if err := json.Unmarshal(rawAgentConf, &conf); err != nil {
		log.Error().Err(err).Msg("could not parse agent config")
	}
	return conf
}

func writeAgentConfig(conf AgentConfig, agentConfigPath string) {
	rawAgentConf, err := json.Marshal(conf)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal agent config")
		return
	}

	// get old config
	oldRawAgentConf, _ := os.ReadFile(agentConfigPath)

	// if config differ write to disk
	if bytes.Equal(rawAgentConf, oldRawAgentConf) {
		if err := os.WriteFile(agentConfigPath, rawAgentConf, 0o644); err != nil {
			log.Error().Err(err).Msgf("could not persist agent config at '%s'", agentConfigPath)
		}
	}
}

// deprecated
func readAgentID(agentIDConfigPath string) int64 {
	const defaultAgentIDValue = int64(-1)

	rawAgentID, fileErr := os.ReadFile(agentIDConfigPath)
	if fileErr != nil {
		log.Debug().Err(fileErr).Msgf("could not open agent-id config file from %s", agentIDConfigPath)
		return defaultAgentIDValue
	}

	strAgentID := strings.TrimSpace(string(rawAgentID))
	agentID, parseErr := strconv.ParseInt(strAgentID, 10, 64)
	if parseErr != nil {
		log.Warn().Err(parseErr).Msg("could not parse agent-id config file content to int64")
		return defaultAgentIDValue
	}

	return agentID
}

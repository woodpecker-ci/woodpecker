// Copyright 2018 Drone.IO Inc.
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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/agent"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
	"github.com/woodpecker-ci/woodpecker/version"
)

// the file implements some basic healthcheck logic based on the
// following specification:
//   https://github.com/mozilla-services/Dockerflow

func initHealth() {
	http.HandleFunc("/varz", handleStats)
	http.HandleFunc("/healthz", handleHeartbeat)
	http.HandleFunc("/version", handleVersion)
}

func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if counter.Healthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/json")
	if err := json.NewEncoder(w).Encode(versionResp{
		Source:  constant.WoodpeckerSourceCode,
		Version: version.String(),
	}); err != nil {
		log.Err(err).Msg("handleVersion could not encode")
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if counter.Healthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "text/json")
	if _, err := counter.WriteTo(w); err != nil {
		log.Error().Err(err).Msg("handleStats")
	}
}

type versionResp struct {
	Version string `json:"version"`
	Source  string `json:"source"`
}

// default statistics counter
var counter = &agent.State{
	Metadata: map[string]agent.Info{},
}

// handles pinging the endpoint and returns an error if the
// agent is in an unhealthy state.
func pinger(c *cli.Context) error {
	resp, err := http.Get("http://localhost:3000/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent returned non-200 status code")
	}
	return nil
}

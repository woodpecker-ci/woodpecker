// Copyright 2023 Woodpecker Authors
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
	"fmt"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
)

// handles pinging the endpoint and returns an error if the
// server is in an unhealthy state.
func pinger(c *cli.Context) error {
	scheme := "http"
	serverAddr := c.String("server-addr")
	if strings.HasPrefix(serverAddr, ":") {
		// this seems sufficient according to https://pkg.go.dev/net#Dial
		serverAddr = "localhost" + serverAddr
	}

	// if woodpecker do ssl on it's own
	if c.String("server-cert") != "" || c.Bool("lets-encrypt") {
		scheme = "https"
	}

	// create the health url
	healthURL := fmt.Sprintf("%s://%s/healthz", scheme, serverAddr)

	// ask server if all is healthy
	resp, err := http.Get(healthURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code")
	}
	return nil
}

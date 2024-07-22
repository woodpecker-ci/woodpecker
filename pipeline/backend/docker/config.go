// Copyright 2024 Woodpecker Authors
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

package docker

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

type config struct {
	enableIPv6    bool
	network       string
	volumes       []string
	resourceLimit resourceLimit
}

type resourceLimit struct {
	MemSwapLimit int64
	MemLimit     int64
	ShmSize      int64
	CPUQuota     int64
	CPUShares    int64
	CPUSet       string
}

func configFromCli(c *cli.Command) (config, error) {
	conf := config{
		enableIPv6: c.Bool("backend-docker-ipv6"),
		network:    c.String("backend-docker-network"),
		resourceLimit: resourceLimit{
			MemSwapLimit: c.Int("backend-docker-limit-mem-swap"),
			MemLimit:     c.Int("backend-docker-limit-mem"),
			ShmSize:      c.Int("backend-docker-limit-shm-size"),
			CPUQuota:     c.Int("backend-docker-limit-cpu-quota"),
			CPUShares:    c.Int("backend-docker-limit-cpu-shares"),
			CPUSet:       c.String("backend-docker-limit-cpu-set"),
		},
	}

	volumes := strings.Split(c.String("backend-docker-volumes"), ",")
	conf.volumes = make([]string, 0, len(volumes))
	// Validate provided volume definitions
	for _, v := range volumes {
		if v == "" {
			continue
		}
		parts, err := splitVolumeParts(v)
		if err != nil {
			log.Error().Err(err).Msgf("can not parse volume config")
			return conf, fmt.Errorf("invalid volume '%s' provided in WOODPECKER_BACKEND_DOCKER_VOLUMES: %w", v, err)
		}
		conf.volumes = append(conf.volumes, strings.Join(parts, ":"))
	}

	return conf, nil
}

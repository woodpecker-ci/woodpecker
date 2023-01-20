// Copyright 2018 Drone.IO Inc.
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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
//
// This file has been modified by Informatyka Boguslawski sp. z o.o. sp.k.

package server

import (
	"crypto"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/cache"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/logging"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/config"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/queue"
)

var Config = struct {
	Services struct {
		Pubsub              pubsub.Publisher
		Queue               queue.Queue
		Logs                logging.Log
		Secrets             model.SecretService
		Registries          model.RegistryService
		Environ             model.EnvironService
		Forge               forge.Forge
		Timeout             time.Duration
		Membership          cache.MembershipService
		ConfigService       config.Extension
		SignaturePrivateKey crypto.PrivateKey
		SignaturePublicKey  crypto.PublicKey
	}
	Storage struct {
		// Users  model.UserStore
		// Repos  model.RepoStore
		// Builds model.BuildStore
		// Logs   model.LogStore
		Files model.FileStore
		Steps model.StepStore
		// Registries model.RegistryStore
		// Secrets model.SecretStore
	}
	Server struct {
		Key                 string
		Cert                string
		OAuthHost           string
		Host                string
		Port                string
		Pass                string
		Docs                string
		StatusContext       string
		StatusContextFormat string
		SessionExpires      time.Duration
		// Open bool
		// Orgs map[string]struct{}
		// Admins map[string]struct{}
	}
	Prometheus struct {
		AuthToken string
	}
	Pipeline struct {
		AuthenticatePublicRepos             bool
		DefaultCancelPreviousPipelineEvents []model.WebhookEvent
		DefaultCloneImage                   string
		Limits                              model.ResourceLimit
		Volumes                             []string
		Networks                            []string
		Privileged                          []string
	}
	FlatPermissions bool // TODO(485) temporary workaround to not hit api rate limits
}{}

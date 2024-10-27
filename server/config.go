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

package server

import (
	"time"

	"go.woodpecker-ci.org/woodpecker/v2/server/cache"
	"go.woodpecker-ci.org/woodpecker/v2/server/logging"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/services"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/permissions"
)

var Config = struct {
	Services struct {
		Pubsub     *pubsub.Publisher
		Queue      queue.Queue
		Logs       logging.Log
		Membership cache.MembershipService
		Manager    services.Manager
		LogStore   log.Service
	}
	Server struct {
		JWTSecret           string
		Key                 string
		Cert                string
		OAuthHost           string
		Host                string
		WebhookHost         string
		Port                string
		PortTLS             string
		AgentToken          string
		StatusContext       string
		StatusContextFormat string
		SessionExpires      time.Duration
		RootPath            string
		CustomCSSFile       string
		CustomJsFile        string
	}
	WebUI struct {
		EnableSwagger    bool
		SkipVersionCheck bool
	}
	Prometheus struct {
		AuthToken string
	}
	Pipeline struct {
		AuthenticatePublicRepos             bool
		DefaultCancelPreviousPipelineEvents []model.WebhookEvent
		DefaultClonePlugin                  string
		TrustedClonePlugins                 []string
		Volumes                             []string
		Networks                            []string
		PrivilegedPlugins                   []string
		DefaultTimeout                      int64
		MaxTimeout                          int64
		Proxy                               struct {
			No    string
			HTTP  string
			HTTPS string
		}
	}
	Permissions struct {
		Open            bool
		Admins          *permissions.Admins
		Orgs            *permissions.Orgs
		OwnersAllowlist *permissions.OwnersAllowlist
	}
}{}

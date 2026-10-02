// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package setup

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/addon"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucketdatacenter"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/forgejo"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitea"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitlab"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Forge(forge *model.Forge) (forge.Forge, error) {
	switch forge.Type {
	case model.ForgeTypeAddon:
		return setupAddon(forge)
	case model.ForgeTypeGithub:
		return setupGitHub(forge)
	case model.ForgeTypeGitlab:
		return setupGitLab(forge)
	case model.ForgeTypeBitbucket:
		return setupBitbucket(forge)
	case model.ForgeTypeGitea:
		return setupGitea(forge)
	case model.ForgeTypeForgejo:
		return setupForgejo(forge)
	case model.ForgeTypeBitbucketDatacenter:
		return setupBitbucketDatacenter(forge)
	default:
		return nil, fmt.Errorf("forge not configured")
	}
}

func setupBitbucket(forge *model.Forge) (forge.Forge, error) {
	opts := &bitbucket.Opts{
		OAuthClientID:     forge.OAuthClientID,
		OAuthClientSecret: forge.OAuthClientSecret,
	}

	log.Debug().
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return bitbucket.New(forge.ID, opts)
}

func setupGitea(forge *model.Forge) (forge.Forge, error) {
	serverURL, err := url.Parse(forge.URL)
	if err != nil {
		return nil, err
	}

	opts := gitea.Opts{
		URL:               strings.TrimRight(serverURL.String(), "/"),
		OAuthClientID:     forge.OAuthClientID,
		OAuthClientSecret: forge.OAuthClientSecret,
		SkipVerify:        forge.SkipVerify,
		OAuthHost:         forge.OAuthHost,
	}
	if len(opts.URL) == 0 {
		return nil, fmt.Errorf("WOODPECKER_GITEA_URL must be set")
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth-host", opts.OAuthHost).
		Bool("skip-verify", opts.SkipVerify).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-secret-id-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return gitea.New(forge.ID, opts)
}

func setupForgejo(forge *model.Forge) (forge.Forge, error) {
	server, err := url.Parse(forge.URL)
	if err != nil {
		return nil, err
	}

	opts := forgejo.Opts{
		URL:               strings.TrimRight(server.String(), "/"),
		OAuthClientID:     forge.OAuthClientID,
		OAuthClientSecret: forge.OAuthClientSecret,
		SkipVerify:        forge.SkipVerify,
		OAuth2URL:         forge.OAuthHost,
	}
	if len(opts.URL) == 0 {
		return nil, fmt.Errorf("WOODPECKER_FORGEJO_URL must be set")
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth2-url", opts.OAuth2URL).
		Bool("skip-verify", opts.SkipVerify).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return forgejo.New(forge.ID, opts)
}

func setupGitLab(forge *model.Forge) (forge.Forge, error) {
	opts := gitlab.Opts{
		URL:               forge.URL,
		OAuthClientID:     forge.OAuthClientID,
		OAuthClientSecret: forge.OAuthClientSecret,
		SkipVerify:        forge.SkipVerify,
		OAuthHost:         forge.OAuthHost,
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth-host", opts.OAuthHost).
		Bool("skip-verify", opts.SkipVerify).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return gitlab.New(forge.ID, opts)
}

func setupGitHub(forge *model.Forge) (forge.Forge, error) {
	// get additional config and be false by default
	mergeRef, _ := forge.AdditionalOptions[model.ForgeGithubOptionMergeRef].(bool)
	publicOnly, _ := forge.AdditionalOptions[model.ForgeGithubOptionPublicOnly].(bool)
	// GitHub App credentials are optional, without them user OAuth tokens are used for all API calls
	appID, _ := forge.AdditionalOptions[model.ForgeGithubOptionAppID].(string)
	appPrivateKey, _ := forge.AdditionalOptions[model.ForgeGithubOptionAppPrivateKey].(string)
	appCloneTokenScope, _ := forge.AdditionalOptions[model.ForgeGithubOptionAppCloneTokenScope].(string)

	opts := github.Opts{
		URL:                forge.URL,
		OAuthClientID:      forge.OAuthClientID,
		OAuthClientSecret:  forge.OAuthClientSecret,
		SkipVerify:         forge.SkipVerify,
		MergeRef:           mergeRef,
		OnlyPublic:         publicOnly,
		OAuthHost:          forge.OAuthHost,
		AppID:              appID,
		AppPrivateKey:      appPrivateKey,
		AppCloneTokenScope: appCloneTokenScope,
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth-host", opts.OAuthHost).
		Bool("merge-ref", opts.MergeRef).
		Bool("only-public", opts.OnlyPublic).
		Bool("skip-verify", opts.SkipVerify).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Bool("app-id-set", opts.AppID != "").
		Bool("app-private-key-set", opts.AppPrivateKey != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return github.New(forge.ID, opts)
}

func setupBitbucketDatacenter(forge *model.Forge) (forge.Forge, error) {
	gitUsername, ok := forge.AdditionalOptions[model.ForgeBitbucketDCOptionGitUsername].(string)
	if !ok {
		return nil, fmt.Errorf("missing git-username")
	}
	gitPassword, ok := forge.AdditionalOptions[model.ForgeBitbucketDCOptionGitPassword].(string)
	if !ok {
		return nil, fmt.Errorf("missing git-password")
	}

	// this option is not exposed by the admin UI, so it may be missing when
	// the forge was configured through the UI/API - default to false like the
	// other optional booleans
	enableProjectAdminScope, _ := forge.AdditionalOptions[model.ForgeBitbucketDCOptionAdminScope].(bool)

	opts := bitbucketdatacenter.Opts{
		URL:                          forge.URL,
		OAuthClientID:                forge.OAuthClientID,
		OAuthClientSecret:            forge.OAuthClientSecret,
		Username:                     gitUsername,
		Password:                     gitPassword,
		OAuthHost:                    forge.OAuthHost,
		OAuthEnableProjectAdminScope: enableProjectAdminScope,
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth-host", opts.OAuthHost).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Bool("oauth-enable-project-admin-scope", opts.OAuthEnableProjectAdminScope).
		Msg("setting up forge")
	return bitbucketdatacenter.New(forge.ID, opts)
}

func setupAddon(forge *model.Forge) (forge.Forge, error) {
	executable, ok := forge.AdditionalOptions[model.ForgeAddonOptionExecutable].(string)
	if !ok {
		return nil, fmt.Errorf("missing addon executable")
	}

	log.Debug().Str("executable", executable).Msg("setting up forge")
	return addon.Load(executable)
}

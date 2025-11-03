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
	return bitbucket.New(opts)
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
	return gitea.New(opts)
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
	return forgejo.New(opts)
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
	return gitlab.New(opts)
}

func setupGitHub(forge *model.Forge) (forge.Forge, error) {
	// get additional config and be false by default
	mergeRef, _ := forge.AdditionalOptions["merge-ref"].(bool)
	publicOnly, _ := forge.AdditionalOptions["public-only"].(bool)

	opts := github.Opts{
		URL:               forge.URL,
		OAuthClientID:     forge.OAuthClientID,
		OAuthClientSecret: forge.OAuthClientSecret,
		SkipVerify:        forge.SkipVerify,
		MergeRef:          mergeRef,
		OnlyPublic:        publicOnly,
		OAuthHost:         forge.OAuthHost,
	}
	log.Debug().
		Str("url", opts.URL).
		Str("oauth-host", opts.OAuthHost).
		Bool("merge-ref", opts.MergeRef).
		Bool("only-public", opts.OnlyPublic).
		Bool("skip-verify", opts.SkipVerify).
		Bool("oauth-client-id-set", opts.OAuthClientID != "").
		Bool("oauth-client-secret-set", opts.OAuthClientSecret != "").
		Str("type", string(forge.Type)).
		Msg("setting up forge")
	return github.New(opts)
}

func setupBitbucketDatacenter(forge *model.Forge) (forge.Forge, error) {
	gitUsername, ok := forge.AdditionalOptions["git-username"].(string)
	if !ok {
		return nil, fmt.Errorf("missing git-username")
	}
	gitPassword, ok := forge.AdditionalOptions["git-password"].(string)
	if !ok {
		return nil, fmt.Errorf("missing git-password")
	}

	enableProjectAdminScope, ok := forge.AdditionalOptions["oauth-enable-project-admin-scope"].(bool)
	if !ok {
		return nil, fmt.Errorf("incorrect type for oauth-enable-project-admin-scope value")
	}

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
	return bitbucketdatacenter.New(opts)
}

func setupAddon(forge *model.Forge) (forge.Forge, error) {
	executable, ok := forge.AdditionalOptions["executable"].(string)
	if !ok {
		return nil, fmt.Errorf("missing addon executable")
	}

	log.Debug().Str("executable", executable).Msg("setting up forge")
	return addon.Load(executable)
}

package setup

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/addon"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucket"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucketdatacenter"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/forgejo"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitea"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/github"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitlab"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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
		Client: forge.Client,
		Secret: forge.ClientSecret,
	}
	log.Trace().Msgf("Forge (bitbucket) opts: %#v", opts)
	return bitbucket.New(opts)
}

func setupGitea(forge *model.Forge) (forge.Forge, error) {
	serverURL, err := url.Parse(forge.URL)
	if err != nil {
		return nil, err
	}

	opts := gitea.Opts{
		URL:        strings.TrimRight(serverURL.String(), "/"),
		Client:     forge.Client,
		Secret:     forge.ClientSecret,
		SkipVerify: forge.SkipVerify,
		OAuthHost:  forge.OAuthHost,
	}
	if len(opts.URL) == 0 {
		return nil, fmt.Errorf("WOODPECKER_GITEA_URL must be set")
	}
	log.Trace().Msgf("Forge (gitea) opts: %#v", opts)
	return gitea.New(opts)
}

func setupForgejo(forge *model.Forge) (forge.Forge, error) {
	server, err := url.Parse(forge.URL)
	if err != nil {
		return nil, err
	}

	opts := forgejo.Opts{
		URL:        strings.TrimRight(server.String(), "/"),
		Client:     forge.Client,
		Secret:     forge.ClientSecret,
		SkipVerify: forge.SkipVerify,
		OAuth2URL:  forge.OAuthHost,
	}
	if len(opts.URL) == 0 {
		return nil, fmt.Errorf("WOODPECKER_FORGEJO_URL must be set")
	}
	log.Trace().Msgf("Forge (forgejo) opts: %#v", opts)
	return forgejo.New(opts)
}

func setupGitLab(forge *model.Forge) (forge.Forge, error) {
	return gitlab.New(gitlab.Opts{
		URL:          forge.URL,
		ClientID:     forge.Client,
		ClientSecret: forge.ClientSecret,
		SkipVerify:   forge.SkipVerify,
		OAuthHost:    forge.OAuthHost,
	})
}

func setupGitHub(forge *model.Forge) (forge.Forge, error) {
	mergeRef, ok := forge.AdditionalOptions["merge-ref"].(bool)
	if !ok {
		return nil, fmt.Errorf("missing merge-ref")
	}

	publicOnly, ok := forge.AdditionalOptions["public-only"].(bool)
	if !ok {
		return nil, fmt.Errorf("missing public-only")
	}

	opts := github.Opts{
		URL:        forge.URL,
		Client:     forge.Client,
		Secret:     forge.ClientSecret,
		SkipVerify: forge.SkipVerify,
		MergeRef:   mergeRef,
		OnlyPublic: publicOnly,
		OAuthHost:  forge.OAuthHost,
	}
	log.Trace().Msgf("Forge (github) opts: %#v", opts)
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

	opts := bitbucketdatacenter.Opts{
		URL:          forge.URL,
		ClientID:     forge.Client,
		ClientSecret: forge.ClientSecret,
		Username:     gitUsername,
		Password:     gitPassword,
		OAuthHost:    forge.OAuthHost,
	}
	log.Trace().Msgf("Forge (bitbucketdatacenter) opts: %#v", opts)
	return bitbucketdatacenter.New(opts)
}

func setupAddon(forge *model.Forge) (forge.Forge, error) {
	executable, ok := forge.AdditionalOptions["executable"].(string)
	if !ok {
		return nil, fmt.Errorf("missing git-username")
	}

	log.Trace().Msgf("Forge (addon) executable: %#v", executable)
	return addon.Load(executable)
}

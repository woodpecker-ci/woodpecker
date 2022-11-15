package loader

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucket"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucketserver"
	"github.com/woodpecker-ci/woodpecker/server/forge/coding"
	"github.com/woodpecker-ci/woodpecker/server/forge/gitea"
	"github.com/woodpecker-ci/woodpecker/server/forge/github"
	"github.com/woodpecker-ci/woodpecker/server/forge/gitlab"
	"github.com/woodpecker-ci/woodpecker/server/forge/gogs"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// setupForge helper function to setup the forge from the CLI arguments.
func GetForge(store store.Store, repo *model.Repo) (forge.Forge, error) {
	forge, err := store.ForgeFind(repo)
	if err != nil {
		return nil, err
	}

	switch forge.Type {
	case "github":
		return setupGitHub(forge)
	case "gitlab":
		return setupGitLab(forge)
	case "bitbucket":
		return setupBitbucket(forge)
	case "stash":
		return setupStash(forge)
	case "gogs":
		return setupGogs(forge)
	case "gitea":
		return setupGitea(forge)
	case "coding":
		return setupCoding(forge)
	default:
		return nil, fmt.Errorf("version control system not configured")
	}

}

// helper function to setup the Bitbucket forge from the CLI arguments.
func setupBitbucket(forge *model.Forge) (forge.Forge, error) {
	opts := &bitbucket.Opts{
		Client: forge.Client,
		Secret: forge.ClientSecret,
	}
	log.Trace().Msgf("Forge (bitbucket) opts: %#v", opts)
	return bitbucket.New(opts)
}

// helper function to setup the Gogs forge from the CLI arguments.
func setupGogs(forge *model.Forge) (forge.Forge, error) {
	opts := gogs.Opts{
		URL: forge.URL,
		// Username:    c.String("gogs-git-username"),
		// Password:    c.String("gogs-git-password"),
		// PrivateMode: c.Bool("gogs-private-mode"),
		SkipVerify: forge.SkipVerify,
	}
	log.Trace().Msgf("Forge (gogs) opts: %#v", opts)
	return gogs.New(opts)
}

// helper function to setup the Gitea forge from the CLI arguments.
func setupGitea(forge *model.Forge) (forge.Forge, error) {
	server, err := url.Parse(forge.URL)
	if err != nil {
		return nil, err
	}
	opts := gitea.Opts{
		URL:        strings.TrimRight(server.String(), "/"),
		Client:     forge.Client,
		Secret:     forge.ClientSecret,
		SkipVerify: forge.SkipVerify,
	}
	if len(opts.URL) == 0 {
		log.Fatal().Msg("WOODPECKER_GITEA_URL must be set")
	}
	log.Trace().Msgf("Forge (gitea) opts: %#v", opts)
	return gitea.New(opts)
}

// helper function to setup the Stash forge from the CLI arguments.
func setupStash(forge *model.Forge) (forge.Forge, error) {
	opts := bitbucketserver.Opts{
		URL: forge.URL,
		// Username:          c.String("stash-git-username"),
		// Password:          c.String("stash-git-password"),
		// ConsumerKey:       c.String("stash-consumer-key"),
		// ConsumerRSA:       c.String("stash-consumer-rsa"),
		// ConsumerRSAString: c.String("stash-consumer-rsa-string"),
		SkipVerify: forge.SkipVerify,
	}
	log.Trace().Msgf("Forge (bitbucketserver) opts: %#v", opts)
	return bitbucketserver.New(opts)
}

// helper function to setup the GitLab forge from the CLI arguments.
func setupGitLab(forge *model.Forge) (forge.Forge, error) {
	return gitlab.New(gitlab.Opts{
		URL:          forge.URL,
		ClientID:     forge.Client,
		ClientSecret: forge.ClientSecret,
		SkipVerify:   forge.SkipVerify,
	})
}

// helper function to setup the GitHub forge from the CLI arguments.
func setupGitHub(forge *model.Forge) (forge.Forge, error) {
	opts := github.Opts{
		URL:        forge.URL,
		Client:     forge.Client,
		Secret:     forge.ClientSecret,
		SkipVerify: forge.SkipVerify,
		// MergeRef:   c.Bool("github-merge-ref"),
	}
	log.Trace().Msgf("Forge (github) opts: %#v", opts)
	return github.New(opts)
}

// helper function to setup the Coding forge from the CLI arguments.
func setupCoding(forge *model.Forge) (forge.Forge, error) {
	opts := coding.Opts{
		URL:    forge.URL,
		Client: forge.Client,
		Secret: forge.ClientSecret,
		// Scopes:     c.StringSlice("coding-scope"),
		// Username:   c.String("coding-git-username"),
		// Password:   c.String("coding-git-password"),
		SkipVerify: forge.SkipVerify,
	}
	log.Trace().Msgf("Forge (coding) opts: %#v", opts)
	return coding.New(opts)
}

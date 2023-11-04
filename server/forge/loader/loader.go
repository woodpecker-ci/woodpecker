package loader

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucket"
	"github.com/woodpecker-ci/woodpecker/server/forge/gitea"
	"github.com/woodpecker-ci/woodpecker/server/forge/github"
	"github.com/woodpecker-ci/woodpecker/server/forge/gitlab"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type forgeLoader struct {
	cache *ttlcache.Cache[int64, forge.Forge]
	store store.Store
	ttl   time.Duration
}

func NewForgeService(_store store.Store) forge.ForgeService {
	return &forgeLoader{
		ttl:   10 * time.Minute,
		store: _store,
		cache: ttlcache.New(ttlcache.WithDisableTouchOnHit[int64, forge.Forge]()),
	}
}

func (f *forgeLoader) getForgeByID(id int64) (forge.Forge, error) {
	item := f.cache.Get(id)
	if item != nil && !item.IsExpired() {
		return item.Value(), nil
	}

	forgeModel, err := f.store.ForgeGet(id)
	if err != nil {
		return nil, err
	}

	forge, err := setupForge(forgeModel)
	if err != nil {
		return nil, err
	}

	f.cache.Set(id, forge, f.ttl)

	return forge, nil
}

func (f *forgeLoader) FromRepo(repo *model.Repo) (forge.Forge, error) {
	return f.getForgeByID(repo.ForgeID)
}

func (f *forgeLoader) FromUser(user *model.User) (forge.Forge, error) {
	return f.getForgeByID(user.ForgeID)
}

func setupForge(forge *model.Forge) (forge.Forge, error) {
	switch forge.Type {
	case "github":
		return setupGitHub(forge)
	case "gitlab":
		return setupGitLab(forge)
	case "bitbucket":
		return setupBitbucket(forge)
	case "gitea":
		return setupGitea(forge)
	default:
		return nil, fmt.Errorf("forge not configured")
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
		return nil, fmt.Errorf("WOODPECKER_GITEA_URL must be set")
	}
	log.Trace().Msgf("Forge (gitea) opts: %#v", opts)
	return gitea.New(opts)
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

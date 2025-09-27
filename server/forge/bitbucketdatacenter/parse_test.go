package bitbucketdatacenter

import (
	"bytes"
	"net/http"
	"testing"

	bb "github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucketdatacenter/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_parseHook(t *testing.T) {
	t.Run("pull-request opened", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPull)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "pr:opened")

		result, err := parseHook(req, "https://bitbucket.example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, &bb.PullRequestEvent{}, result.Event)
		assert.NotNil(t, result.Repo)
		assert.NotNil(t, result.Pipeline)
		assert.NotNil(t, result.Payload)
		assert.Equal(t, "DEV/network-monitor", result.Repo.FullName)
		assert.Equal(t, "1c7589876bc8b5e83122b1656925d679915193d4", result.Pipeline.Commit)
		assert.Equal(t, model.EventPull, result.Pipeline.Event)
	})

	t.Run("pull-request opened from fork", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullFork)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "pr:opened")

		result, err := parseHook(req, "https://bitbucket.example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, &bb.PullRequestEvent{}, result.Event)
		assert.NotNil(t, result.Repo)
		assert.NotNil(t, result.Pipeline)
		assert.NotNil(t, result.Payload)
		assert.Equal(t, "DEV/deployment-automation", result.Repo.FullName)
		assert.Equal(t, "716e510cecbe203618609cf103c54e040b949739", result.Pipeline.Commit)
		assert.Equal(t, model.EventPull, result.Pipeline.Event)
	})

	t.Run("push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "repo:refs_changed")

		result, err := parseHook(req, "https://bitbucket.example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, &bb.RepositoryPushEvent{}, result.Event)
		assert.NotNil(t, result.Repo)
		assert.NotNil(t, result.Pipeline)
		assert.NotNil(t, result.Payload)
		assert.Equal(t, "DEV/deployment-automation", result.Repo.FullName)
		assert.Equal(t, "76797d54bca87db6d1e3e82ee40622c7908aa514", result.Pipeline.Commit)
		assert.Equal(t, model.EventPush, result.Pipeline.Event)
	})

	t.Run("pull-request merged", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullMerged)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "pr:merged")

		result, err := parseHook(req, "https://bitbucket.example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, &bb.PullRequestEvent{}, result.Event)
		assert.NotNil(t, result.Repo)
		assert.NotNil(t, result.Pipeline)
		assert.NotNil(t, result.Payload)
		assert.Equal(t, "DEV/deployment-automation", result.Repo.FullName)
		assert.Equal(t, "993203acecdb65ffe947424d0917768b0e5c3903", result.Pipeline.Commit)
		assert.Equal(t, model.EventPullClosed, result.Pipeline.Event)
	})
}

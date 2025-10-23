package bitbucketdatacenter

import (
	"fmt"
	"net/http"

	bb "github.com/neticdk/go-bitbucket/bitbucket"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type HookResult struct {
	Repo     *model.Repo
	Pipeline *model.Pipeline
	Event    any
	Payload  []byte
}

func parseHook(r *http.Request, baseURL string) (*HookResult, error) {
	ev, payload, err := bb.ParsePayloadWithoutSignature(r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse payload from webhook invocation: %w", err)
	}

	result := &HookResult{
		Event:   ev,
		Payload: payload,
	}

	switch e := ev.(type) {
	case *bb.RepositoryPushEvent:
		result.Repo = convertRepo(&e.Repository, nil, "")
		result.Pipeline = convertRepositoryPushEvent(e, baseURL)
	case *bb.PullRequestEvent:
		result.Repo = convertRepo(&e.PullRequest.Target.Repository, nil, "")
		result.Pipeline = convertPullRequestEvent(e, baseURL)
	default:
		return nil, &types.ErrIgnoreEvent{Event: fmt.Sprintf("%T", e), Reason: "unsupported webhook event type"}
	}

	return result, nil
}

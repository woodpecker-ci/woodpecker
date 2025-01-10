package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc/mocks"
)

type peery struct {
	*mocks.Peer
}

func (p *peery) Done(ctx context.Context, id string, state rpc.WorkflowState) error {
	return nil
}

func TestRunnerCanceledState(t *testing.T) {
	backend := dummy.New()
	_peer := mocks.NewPeer(t)

	peer := &peery{_peer}

	hostname := "dummy"
	filter := rpc.Filter{
		Labels: map[string]string{
			"hostname": hostname,
			"platform": "test",
			"backend":  backend.Name(),
			"repo":     "*", // allow all repos by default
		},
	}
	state := &State{
		Metadata: map[string]Info{},
		Polling:  1, // max workflows to poll
		Running:  0,
	}
	r := NewRunner(peer, filter, hostname, state, &backend)
	ctx, cancel := context.WithCancel(context.Background())

	workflow := &rpc.Workflow{
		ID: "1",
		Config: &types.Config{
			Stages: []*types.Stage{
				{
					Steps: []*types.Step{
						{

							Name: "test",
							Environment: map[string]string{
								"SLEEP": "10s",
							},
							Commands: []string{
								"echo 'hello world'",
							},
							OnSuccess: true,
						},
					},
				},
			},
		},
		Timeout: 1, // 1 minute
	}

	peer.On("Next", mock.Anything, filter).Return(workflow, nil).Once()
	peer.On("Init", mock.Anything, "1", mock.MatchedBy(func(state rpc.WorkflowState) bool {
		return state.Started != 0 && state.Finished == 0 && state.Error == ""
	})).Return(nil)
	peer.On("Done", mock.Anything, "1", mock.MatchedBy(func(state rpc.WorkflowState) bool {
		return state.Started != 0 && state.Finished != 0 && state.Error == ""
	})).Return(nil)
	peer.On("Log", mock.Anything, mock.Anything).Return(nil)
	peer.On("Wait", mock.Anything, "1").Return(nil)
	peer.On("Update", mock.Anything, "1", mock.Anything).Return(nil)
	peer.On("Extend", mock.Anything, "1").Return(nil).Maybe()

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("canceling ...")
		cancel()
	}()

	err := r.Run(ctx, ctx)
	assert.NoError(t, err)
}

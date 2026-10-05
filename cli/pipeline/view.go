// Copyright 2026 Woodpecker Authors
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

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal/tui"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var pipelineViewCmd = &cli.Command{
	Name:      "view",
	Usage:     "show steps and logs of a pipeline in an interactive view, live while it runs",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline]",
	Action:    pipelineView,
}

// viewRefreshInterval is how often the status of an unfinished pipeline is fetched.
var viewRefreshInterval = 2 * time.Second

func pipelineView(ctx context.Context, c *cli.Command) error {
	if !logger.IsInteractiveTerminal() {
		return errors.New("the pipeline view needs a terminal, use 'pipeline ps' and 'pipeline log show' without one")
	}

	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return fmt.Errorf("invalid repo '%s': %w", repoIDOrFullName, err)
	}

	var number int64
	if pipelineArg := c.Args().Get(1); pipelineArg == "last" || len(pipelineArg) == 0 {
		pipeline, err := client.PipelineLast(repoID, woodpecker.PipelineLastOptions{})
		if err != nil {
			return err
		}
		number = pipeline.Number
	} else {
		number, err = strconv.ParseInt(pipelineArg, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid pipeline '%s': %w", pipelineArg, err)
		}
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	model := tui.New()
	program := tea.NewProgram(model)
	viewer := &pipelineViewer{
		client: client,
		repoID: repoID,
		number: number,
		send:   program.Send,
		loaded: make(map[int64]bool),
	}
	model.OnSelect = func(step *woodpecker.Step) { viewer.loadLog(ctx, step) }
	go viewer.watch(ctx)

	_, err = program.Run()
	fmt.Print(model.Summary())
	return err
}

// pipelineViewer feeds the pipeline view with a pipeline of the server.
type pipelineViewer struct {
	client woodpecker.Client
	repoID int64
	number int64
	// send passes a message to the view
	send func(tea.Msg)

	mu sync.Mutex
	// steps the log is loaded or being loaded of, by step id
	loaded map[int64]bool
}

// watch sends the workflows of the pipeline to the view, again and again
// until the pipeline is finished.
func (v *pipelineViewer) watch(ctx context.Context) {
	errorsShown := false
	for {
		pipeline, err := v.client.Pipeline(v.repoID, v.number)
		if err != nil {
			v.send(tui.MessageMsg(fmt.Sprintf("could not load pipeline: %s", err)))
		} else {
			if !errorsShown {
				errorsShown = true
				for _, pipelineErr := range pipeline.Errors {
					v.send(tui.MessageMsg(fmt.Sprintf("%s: %s", pipelineErr.Type, pipelineErr.Message)))
				}
			}
			v.send(tui.StatusMsg(pipeline.Status))
			v.send(tui.WorkflowsMsg(pipeline.Workflows))

			switch status.Value(pipeline.Status) {
			case status.Created, status.Blocked, status.Pending, status.Running:
			default:
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(viewRefreshInterval):
		}
	}
}

// loadLog sends the log of a step to the view, once.
// The log of a step that is not finished yet is followed until its end.
func (v *pipelineViewer) loadLog(ctx context.Context, step *woodpecker.Step) {
	v.mu.Lock()
	loaded := v.loaded[step.ID]
	v.loaded[step.ID] = true
	v.mu.Unlock()
	if loaded {
		return
	}

	if err := v.sendLog(ctx, step); err != nil && ctx.Err() == nil {
		v.send(tui.MessageMsg(fmt.Sprintf("could not load log of step %s: %s", step.Name, err)))

		// try again the next time the step gets selected
		v.mu.Lock()
		delete(v.loaded, step.ID)
		v.mu.Unlock()
	}
}

func (v *pipelineViewer) sendLog(ctx context.Context, step *woodpecker.Step) error {
	switch state := status.Value(step.State); {
	case state == status.Skipped || state == status.Canceled:
		// never ran, so there is no log
		return nil
	case state.IsActive():
		err := v.client.StepLogStream(ctx, v.repoID, v.number, step.ID, func(entry *woodpecker.LogEntry) {
			v.send(tui.LogMsg{entry})
		})
		if err == nil || ctx.Err() != nil {
			return nil
		}
		// The stream fails if the step just finished. Either way the stored
		// log has all the lines, as far as there are any.
	}

	entries, err := v.client.StepLogEntries(v.repoID, v.number, step.ID)
	v.send(tui.LogMsg(entries))
	return err
}

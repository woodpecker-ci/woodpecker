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

package exec

import (
	"context"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal/tui"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// tuiOutput shows what a pipeline run reports in the pipeline view.
type tuiOutput struct {
	program *tea.Program
}

func (o tuiOutput) Workflow(workflow *woodpecker.Workflow) {
	o.program.Send(tui.WorkflowsMsg{workflow})
}

func (o tuiOutput) Log(entry *woodpecker.LogEntry) {
	o.program.Send(tui.LogMsg{entry})
}

func (o tuiOutput) Message(text string) {
	o.program.Send(tui.MessageMsg(text))
}

// Write shows text that belongs to no step.
func (o tuiOutput) Write(p []byte) (int, error) {
	o.Message(string(p))
	return len(p), nil
}

// executeWithTUI runs the pipeline while showing it in the pipeline view,
// until the user closes it. The warnings are shown as first messages.
func executeWithTUI(ctx context.Context, pipelineRun *pipelineRun, warnings string) (status.Value, error) {
	model := tui.New()
	model.OnCancel = pipelineRun.cancel
	program := tea.NewProgram(model)
	out := tuiOutput{program}
	pipelineRun.out = out

	// While the view owns the terminal, everything else that would write to
	// it has to go to the messages of the view instead, line by line.
	messagesReader, messages := io.Pipe()
	defer messages.Close()
	go func() {
		_ = pipeline_utils.CopyLineByLine(out, messagesReader, pipeline.MaxLogLineLength)
	}()

	logger := log.Logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: messages, NoColor: true})
	ctx = context.WithValue(ctx, backend_types.ImagePullOutput, messages)

	result := make(chan status.Value, 1)
	go func() {
		if warnings != "" {
			program.Send(tui.MessageMsg(warnings))
		}
		result <- pipelineRun.execute(ctx)
	}()

	_, err := program.Run()

	// the view is gone, so leave its last state in the terminal
	log.Logger = logger
	fmt.Print(model.Summary())

	// stop what still runs, in case the view got closed early
	pipelineRun.cancel()
	return <-result, err
}

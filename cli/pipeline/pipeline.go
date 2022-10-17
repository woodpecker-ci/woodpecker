// Copyright 2022 Woodpecker Authors
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
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the pipeline command set.
var Command = &cli.Command{
	Name:    "pipeline",
	Aliases: []string{"build"},
	Usage:   "manage pipelines",
	Flags:   common.GlobalFlags,
	Subcommands: []*cli.Command{
		pipelineListCmd,
		pipelineLastCmd,
		pipelineLogsCmd,
		pipelineInfoCmd,
		pipelineStopCmd,
		pipelineStartCmd,
		pipelineApproveCmd,
		pipelineDeclineCmd,
		pipelineQueueCmd,
		pipelineKillCmd,
		pipelinePsCmd,
		pipelineCreateCmd,
	},
}

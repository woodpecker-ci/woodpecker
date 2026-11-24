// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

var moveManualPipelineMessage = xormigrate.Migration{
	ID: "move-manual-pipeline-message",
	MigrateSession: func(sess *xorm.Session) error {
		type commitAuthor struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type commit struct {
			SHA       string       `json:"sha"`
			Message   string       `json:"message"`
			ForgeURL  string       `json:"forge_url"`
			Author    commitAuthor `json:"author"`
			Timestamp int64        `json:"timestamp"`
		}

		type pipelines struct {
			ID            int64   `xorm:"pk autoincr 'id'"`
			Event         string  `xorm:"event"`
			Branch        string  `xorm:"branch"`
			Commit        *commit `xorm:"json 'commit'"`
			ManualMessage string  `xorm:"TEXT 'manual_message'"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		// manual pipelines are rare compared to the others, so they are loaded
		// and updated one by one
		var manualPipelines []*pipelines
		if err := sess.Where(builder.Eq{"event": "manual"}).Find(&manualPipelines); err != nil {
			return err
		}

		for _, p := range manualPipelines {
			if p.Commit == nil {
				continue
			}

			// the message was generated as "MANUAL PIPELINE @ <branch>" or as
			// "MANUAL: <custom message> @ <branch>", anything else is kept
			suffix := " @ " + p.Branch
			if !strings.HasSuffix(p.Commit.Message, suffix) {
				continue
			}
			message := strings.TrimSuffix(p.Commit.Message, suffix)
			switch {
			case message == "MANUAL PIPELINE":
			case strings.HasPrefix(message, "MANUAL: "):
				p.ManualMessage = strings.TrimPrefix(message, "MANUAL: ")
			default:
				continue
			}
			p.Commit.Message = ""

			if _, err := sess.ID(p.ID).Cols("commit", "manual_message").Update(p); err != nil {
				return err
			}
		}

		return nil
	},
}

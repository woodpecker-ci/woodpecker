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

package badges

import (
	"fmt"

	"github.com/narqo/go-badge"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// cspell:words Verdana

var (
	// Status background colors
	badgeColorSuccess = "#44cc11"
	badgeColorFailure = "#e05d44"
	badgeColorStarted = "#dfb317"
	badgeColorError   = "#9f9f9f"
	badgeColorNone    = "#9f9f9f"

	// Status labels
	badgeStatusSuccess = "success"
	badgeStatusFailure = "failure"
	badgeStatusStarted = "started"
	badgeStatusError   = "error"
	badgeStatusNone    = "none"
)

func getBadgeStatus(status *model.StatusValue) (string, badge.Color) {
	if status == nil {
		return badgeStatusNone, badge.Color(badgeColorNone)
	}

	switch *status {
	case model.StatusSuccess:
		return badgeStatusSuccess, badge.Color(badgeColorSuccess)
	case model.StatusFailure:
		return badgeStatusFailure, badge.Color(badgeColorFailure)
	case model.StatusPending, model.StatusRunning:
		return badgeStatusStarted, badge.Color(badgeColorStarted)
	case model.StatusError, model.StatusKilled:
		return badgeStatusError, badge.Color(badgeColorError)
	default:
		return badgeStatusNone, badge.Color(badgeColorNone)
	}
}

// Generate an SVG badge based on a pipeline.
func Generate(name string, status *model.StatusValue) (string, error) {
	label, color := getBadgeStatus(status)
	bytes, err := badge.RenderBytes(name, label, color)
	if err != nil {
		log.Warn().Err(err).Msg("could not render badge")
		return "", err
	}
	return fmt.Sprintf("%s", bytes), nil
}

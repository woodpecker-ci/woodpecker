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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var (
	badgeNone    = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="98" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="98" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="57" height="20" fill="#555"/><rect x="57" width="41" height="20" fill="#9f9f9f"/><rect width="98" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="29.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="29.5" y="14">pipeline</text><text x="76.5" y="15" fill="#010101" fill-opacity=".3">none</text><text x="76.5" y="14">none</text></g></svg>`
	badgeSuccess = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="117" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="117" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="57" height="20" fill="#555"/><rect x="57" width="60" height="20" fill="#44cc11"/><rect width="117" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="29.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="29.5" y="14">pipeline</text><text x="86" y="15" fill="#010101" fill-opacity=".3">success</text><text x="86" y="14">success</text></g></svg>`
	badgeFailure = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="106" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="106" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="57" height="20" fill="#555"/><rect x="57" width="49" height="20" fill="#e05d44"/><rect width="106" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="29.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="29.5" y="14">pipeline</text><text x="80.5" y="15" fill="#010101" fill-opacity=".3">failure</text><text x="80.5" y="14">failure</text></g></svg>`
	badgeError   = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="99" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="99" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="57" height="20" fill="#555"/><rect x="57" width="42" height="20" fill="#9f9f9f"/><rect width="99" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="29.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="29.5" y="14">pipeline</text><text x="77" y="15" fill="#010101" fill-opacity=".3">error</text><text x="77" y="14">error</text></g></svg>`
	badgeStarted = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="111" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="111" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="57" height="20" fill="#555"/><rect x="57" width="54" height="20" fill="#dfb317"/><rect width="111" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="29.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="29.5" y="14">pipeline</text><text x="83" y="15" fill="#010101" fill-opacity=".3">started</text><text x="83" y="14">started</text></g></svg>`
)

// Generate an SVG badge based on a pipeline.
func TestGenerate(t *testing.T) {
	status := model.StatusDeclined
	badge, err := Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeNone, badge)
	status = model.StatusSuccess
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeSuccess, badge)
	status = model.StatusFailure
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeFailure, badge)
	status = model.StatusError
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeError, badge)
	status = model.StatusKilled
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeError, badge)
	status = model.StatusPending
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeStarted, badge)
	status = model.StatusRunning
	badge, err = Generate("pipeline", &status)
	assert.Nil(t, err)
	assert.Equal(t, badgeStarted, badge)
}

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
	badgeNone    = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="102" height="20" role="img" aria-label="pipeline: none"><title>pipeline: none</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="102" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="59" height="20" fill="#555"/><rect x="59" width="59" height="20" fill="#9f9f9f"/><rect width="102" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="431">pipeline</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="431">pipeline</text><text aria-hidden="true" x="807" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="272">none</text><text x="807" y="140" transform="scale(.1)" fill="#fff" textLength="272">none</text></g></svg>`
	badgeSuccess = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="117" height="20" role="img" aria-label="pipeline: success"><title>pipeline: success</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="117" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="59" height="20" fill="#555"/><rect x="59" width="59" height="20" fill="#44cc11"/><rect width="117" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="431">pipeline</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="431">pipeline</text><text aria-hidden="true" x="884" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="427">success</text><text x="884" y="140" transform="scale(.1)" fill="#fff" textLength="427">success</text></g></svg>`
	badgeFailure = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="109" height="20" role="img" aria-label="pipeline: failure"><title>pipeline: failure</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="109" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="59" height="20" fill="#555"/><rect x="59" width="59" height="20" fill="#e05d44"/><rect width="109" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="431">pipeline</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="431">pipeline</text><text aria-hidden="true" x="844" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="346">failure</text><text x="844" y="140" transform="scale(.1)" fill="#fff" textLength="346">failure</text></g></svg>`
	badgeError   = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="101" height="20" role="img" aria-label="pipeline: error"><title>pipeline: error</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="101" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="59" height="20" fill="#555"/><rect x="59" width="59" height="20" fill="#9f9f9f"/><rect width="101" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="431">pipeline</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="431">pipeline</text><text aria-hidden="true" x="805" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="269">error</text><text x="805" y="140" transform="scale(.1)" fill="#fff" textLength="269">error</text></g></svg>`
	badgeStarted = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="114" height="20" role="img" aria-label="pipeline: started"><title>pipeline: started</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="114" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="59" height="20" fill="#555"/><rect x="59" width="59" height="20" fill="#dfb317"/><rect width="114" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="431">pipeline</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="431">pipeline</text><text aria-hidden="true" x="866" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="391">started</text><text x="866" y="140" transform="scale(.1)" fill="#fff" textLength="391">started</text></g></svg>`
)

// Generate an SVG badge based on a pipeline.
func TestGenerate(t *testing.T) {
	status := model.StatusDeclined
	assert.Equal(t, badgeNone, Generate("pipeline", &status))
	status = model.StatusSuccess
	assert.Equal(t, badgeSuccess, Generate("pipeline", &status))
	status = model.StatusFailure
	assert.Equal(t, badgeFailure, Generate("pipeline", &status))
	status = model.StatusError
	assert.Equal(t, badgeError, Generate("pipeline", &status))
	status = model.StatusKilled
	assert.Equal(t, badgeError, Generate("pipeline", &status))
	status = model.StatusPending
	assert.Equal(t, badgeStarted, Generate("pipeline", &status))
	status = model.StatusRunning
	assert.Equal(t, badgeStarted, Generate("pipeline", &status))
}

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
	"bytes"
	"html/template"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var (
	badgeNone    = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="82" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="82" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="49" height="20" fill="#555"/><rect x="49" width="33" height="20" fill="#9f9f9f"/><rect width="82" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="25.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="25.5" y="14">pipeline</text><text x="64.5" y="15" fill="#010101" fill-opacity=".3">none</text><text x="64.5" y="14">none</text></g></svg>`
	badgeSuccess = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="98" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="98" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="49" height="20" fill="#555"/><rect x="49" width="49" height="20" fill="#44cc11"/><rect width="98" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="25.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="25.5" y="14">pipeline</text><text x="72.5" y="15" fill="#010101" fill-opacity=".3">success</text><text x="72.5" y="14">success</text></g></svg>`
	badgeFailure = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="89" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="89" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="49" height="20" fill="#555"/><rect x="49" width="40" height="20" fill="#e05d44"/><rect width="89" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="25.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="25.5" y="14">pipeline</text><text x="68" y="15" fill="#010101" fill-opacity=".3">failure</text><text x="68" y="14">failure</text></g></svg>`
	badgeError   = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="81" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="81" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="49" height="20" fill="#555"/><rect x="49" width="32" height="20" fill="#9f9f9f"/><rect width="81" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="25.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="25.5" y="14">pipeline</text><text x="64" y="15" fill="#010101" fill-opacity=".3">error</text><text x="64" y="14">error</text></g></svg>`
	badgeStarted = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="94" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="94" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="49" height="20" fill="#555"/><rect x="49" width="45" height="20" fill="#dfb317"/><rect width="94" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="25.5" y="15" fill="#010101" fill-opacity=".3">pipeline</text><text x="25.5" y="14">pipeline</text><text x="70.5" y="15" fill="#010101" fill-opacity=".3">started</text><text x="70.5" y="14">started</text></g></svg>`
)

// Generate an SVG badge based on a pipeline.
func TestGenerate(t *testing.T) {
	status := model.StatusDeclined
	badge, err := Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeNone, badge)
	status = model.StatusSuccess
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeSuccess, badge)
	status = model.StatusFailure
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeFailure, badge)
	status = model.StatusError
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeError, badge)
	status = model.StatusKilled
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeError, badge)
	status = model.StatusPending
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeStarted, badge)
	status = model.StatusRunning
	badge, err = Generate("pipeline", &status)
	assert.NoError(t, err)
	assert.Equal(t, badgeStarted, badge)
}

func TestBadgeDrawerRender(t *testing.T) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)
	mockFontSize := 11.0
	mockDPI := 72.0

	fd, err := mustNewFontDrawer(mockFontSize, mockDPI)
	assert.NoError(t, err)

	d := &badgeDrawer{
		fd:    fd,
		tmpl:  template.Must(template.New("mock-template").Parse(mockTemplate)),
		mutex: &sync.Mutex{},
	}

	output := "XXX,YYY,#c0c0c0,15.5,29,41,26,55"

	var buf bytes.Buffer
	assert.NoError(t, d.Render("XXX", "YYY", "#c0c0c0", &buf))
	assert.Equal(t, output, buf.String())
}

func TestBadgeDrawerRenderBytes(t *testing.T) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)
	mockFontSize := 11.0
	mockDPI := 72.0

	fd, err := mustNewFontDrawer(mockFontSize, mockDPI)
	assert.NoError(t, err)

	d := &badgeDrawer{
		fd:    fd,
		tmpl:  template.Must(template.New("mock-template").Parse(mockTemplate)),
		mutex: &sync.Mutex{},
	}

	output := "XXX,YYY,#c0c0c0,15.5,29,41,26,55"

	bytes, err := d.RenderBytes("XXX", "YYY", "#c0c0c0")

	assert.NoError(t, err)
	assert.Equal(t, output, string(bytes))
}

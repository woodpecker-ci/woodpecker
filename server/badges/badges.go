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

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

var (
	badgeSuccess = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="106" height="20" role="img" aria-label="pipeline: success"><title>pipeline: success</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="106" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="53" height="20" fill="#555"/><rect x="53" width="53" height="20" fill="#44cc11"/><rect width="106" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="275" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">pipeline</text><text x="275" y="140" transform="scale(.1)" fill="#fff" textLength="430">pipeline</text><text aria-hidden="true" x="785" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">success</text><text x="785" y="140" transform="scale(.1)" fill="#fff" textLength="430">success</text></g></svg>`
	badgeFailure = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="98" height="20" role="img" aria-label="pipeline: failure"><title>pipeline: failure</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="98" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="53" height="20" fill="#555"/><rect x="53" width="45" height="20" fill="#e05d44"/><rect width="98" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="275" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">pipeline</text><text x="275" y="140" transform="scale(.1)" fill="#fff" textLength="430">pipeline</text><text aria-hidden="true" x="745" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="350">failure</text><text x="745" y="140" transform="scale(.1)" fill="#fff" textLength="350">failure</text></g></svg>`
	badgeStarted = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="102" height="20" role="img" aria-label="pipeline: started"><title>pipeline: started</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="102" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="53" height="20" fill="#555"/><rect x="53" width="49" height="20" fill="#dfb317"/><rect width="102" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="275" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">pipeline</text><text x="275" y="140" transform="scale(.1)" fill="#fff" textLength="430">pipeline</text><text aria-hidden="true" x="765" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="390">started</text><text x="765" y="140" transform="scale(.1)" fill="#fff" textLength="390">started</text></g></svg>`
	badgeError   = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="90" height="20" role="img" aria-label="pipeline: error"><title>pipeline: error</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="90" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="53" height="20" fill="#555"/><rect x="53" width="37" height="20" fill="#9f9f9f"/><rect width="90" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="275" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">pipeline</text><text x="275" y="140" transform="scale(.1)" fill="#fff" textLength="430">pipeline</text><text aria-hidden="true" x="705" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="270">error</text><text x="705" y="140" transform="scale(.1)" fill="#fff" textLength="270">error</text></g></svg>`
	badgeNone    = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="90" height="20" role="img" aria-label="pipeline: none"><title>pipeline: none</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="90" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="53" height="20" fill="#555"/><rect x="53" width="37" height="20" fill="#9f9f9f"/><rect width="90" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="275" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">pipeline</text><text x="275" y="140" transform="scale(.1)" fill="#fff" textLength="430">pipeline</text><text aria-hidden="true" x="705" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="270">none</text><text x="705" y="140" transform="scale(.1)" fill="#fff" textLength="270">none</text></g></svg>`
)

// Generate an SVG badge based on a pipeline
func Generate(pipeline *model.Pipeline) string {
	if pipeline == nil {
		return badgeNone
	}
	switch pipeline.Status {
	case model.StatusSuccess:
		return badgeSuccess
	case model.StatusFailure:
		return badgeFailure
	case model.StatusError, model.StatusKilled:
		return badgeError
	case model.StatusPending, model.StatusRunning:
		return badgeStarted
	default:
		return badgeNone
	}
}

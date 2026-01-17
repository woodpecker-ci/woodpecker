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
	_ "embed"
	"fmt"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// cspell:words Verdana

var (
	// [1] : Label : string
	// [2] : Status : string
	// [3] : totalWidth : int
	// [4] : labelWidth : int
	// [5] : statusWidth : int
	// [6] : statusColor : string
	// [7] : labelTextLength : int
	// [8] : statusTextLength : int
	// [9] : centerOfLabel : int
	// [10] : centerOfStatus : int
	badgeTemplate = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%[3]d" height="20" role="img" aria-label="%[1]s: %[2]s"><title>%[1]s: %[2]s</title><linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="%[3]d" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="%[4]d" height="20" fill="#555"/><rect x="%[4]d" width="%[4]d" height="20" fill="%[6]s"/><rect width="%[3]d" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="%[9]d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%[7]d">%[1]s</text><text x="%[9]d" y="140" transform="scale(.1)" fill="#fff" textLength="%[7]d">%[1]s</text><text aria-hidden="true" x="%[10]d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%[8]d">%[2]s</text><text x="%[10]d" y="140" transform="scale(.1)" fill="#fff" textLength="%[8]d">%[2]s</text></g></svg>`
)

var (
	//go:embed DejaVuSans.ttf
	fontData []byte

	fontFace   font.Face
	once       sync.Once
	parseError error
)

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

	// Status default label width
	badgeStatusSuccessWidth = 430
	badgeStatusFailureWidth = 350
	badgeStatusStartedWidth = 390
	badgeStatusErrorWidth   = 270
	badgeStatusNoneWidth    = 270

	// Minimal width of badge name label
	badgeNameWidthMinimum = 430

	// x-Padding of name and status labels
	badgeNamePadding   = 80
	badgeStatusPadding = 80

	// Glyph fallback width
	glyphFallbackWidth = 100
)

func getBadgeStatusColor(status *model.StatusValue) string {
	if status == nil {
		return badgeColorNone
	}

	switch *status {
	case model.StatusSuccess:
		return badgeColorSuccess
	case model.StatusFailure:
		return badgeColorFailure
	case model.StatusPending, model.StatusRunning:
		return badgeColorStarted
	case model.StatusError, model.StatusKilled:
		return badgeColorError
	default:
		return badgeColorNone
	}
}

func getBadgeStatusFontWidth(status *model.StatusValue) int {
	if status == nil {
		return badgeStatusNoneWidth
	}

	switch *status {
	case model.StatusSuccess:
		return badgeStatusSuccessWidth
	case model.StatusFailure:
		return badgeStatusFailureWidth
	case model.StatusPending, model.StatusRunning:
		return badgeStatusStartedWidth
	case model.StatusError, model.StatusKilled:
		return badgeStatusErrorWidth
	default:
		return badgeStatusNoneWidth
	}
}

func getStatusLabel(status *model.StatusValue) string {
	if status == nil {
		return badgeStatusNone
	}

	switch *status {
	case model.StatusSuccess:
		return badgeStatusSuccess
	case model.StatusFailure:
		return badgeStatusFailure
	case model.StatusPending, model.StatusRunning:
		return badgeStatusStarted
	case model.StatusError, model.StatusKilled:
		return badgeStatusError
	default:
		return badgeStatusNone
	}
}

func loadFontFace() *font.Face {
	once.Do(func() {
		tt, err := truetype.Parse(fontData)
		if err != nil {
			log.Warn().Err(err).Msg("could not initalize font for dynamic badge generation")
			return
		}

		fontFace = truetype.NewFace(tt, &truetype.Options{
			Size: 110,
		})
	})

	if parseError == nil {
		return &fontFace
	} else {
		return nil
	}
}

func getFontWidth(text string, face *font.Face) int {
	width := 0
	for _, c := range text {
		if face != nil {
			_, advance, ok := (*face).GlyphBounds(c)
			if ok {
				width += advance.Floor()
			} else {
				width += glyphFallbackWidth
			}
		} else {
			width += glyphFallbackWidth
		}
	}
	return width
}

// Generate an SVG badge based on a pipeline.
func Generate(name string, status *model.StatusValue) string {
	statusText := getStatusLabel(status)
	statusColor := getBadgeStatusColor(status)

	fontFace := loadFontFace()

	nameFontWidth := max(badgeNameWidthMinimum, getFontWidth(name, fontFace))
	statusFontWidth := getBadgeStatusFontWidth(status)

	// If no fontFace failed to load, use defaults
	if fontFace != nil {
		statusFontWidth = getFontWidth(statusText, fontFace)
	}

	// Get the x-coordinates of the centers of the labels
	centerOfName := (nameFontWidth / 2) + badgeNamePadding
	centerOfStatus := (nameFontWidth + (2 * badgeNamePadding)) + (statusFontWidth / 2) + badgeStatusPadding

	// Get the widths of the name and status labels
	nameWidth := nameFontWidth + (2 * badgeNamePadding)
	statusWidth := statusFontWidth + (2 * badgeStatusPadding)

	// Transform the x-coordinates to approximate box widths
	nameBoxWidth := nameWidth / 10
	statusBoxWidth := statusWidth / 10
	totalBoxWidth := nameBoxWidth + statusBoxWidth

	return fmt.Sprintf(badgeTemplate, name, statusText, totalBoxWidth, nameBoxWidth, statusBoxWidth, statusColor, nameFontWidth, statusFontWidth, centerOfName, centerOfStatus)
}

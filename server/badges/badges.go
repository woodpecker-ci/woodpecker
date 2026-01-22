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
	"io"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font"

	"go.woodpecker-ci.org/woodpecker/v3/server/badges/fonts"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type badge struct {
	Subject string
	Status  string
	Color   Color
	Bounds  bounds
}

type bounds struct {
	// SubjectDx is the width of subject string of the badge.
	SubjectDx float64
	SubjectX  float64
	// StatusDx is the width of status string of the badge.
	StatusDx float64
	StatusX  float64
}

func (b bounds) Dx() float64 {
	return b.SubjectDx + b.StatusDx
}

type badgeDrawer struct {
	fd    *font.Drawer
	tmpl  *template.Template
	mutex *sync.Mutex
}

func (d *badgeDrawer) Render(subject, status string, color Color, w io.Writer) error {
	d.mutex.Lock()
	subjectDx := d.measureString(subject)
	statusDx := d.measureString(status)
	d.mutex.Unlock()

	bdg := badge{
		Subject: subject,
		Status:  status,
		Color:   color,
		Bounds: bounds{
			SubjectDx: subjectDx,
			SubjectX:  subjectDx/2.0 + 1,
			StatusDx:  statusDx,
			StatusX:   subjectDx + statusDx/2.0 - 1,
		},
	}
	return d.tmpl.Execute(w, bdg)
}

func (d *badgeDrawer) RenderBytes(subject, status string, color Color) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := d.Render(subject, status, color, buf)
	return buf.Bytes(), err
}

// shield.io uses Verdana.ttf to measure text width with an extra 10px.
// As we use Vera.ttf, we have to tune this value a little.
const extraDx = 13

func (d *badgeDrawer) measureString(s string) float64 {
	SHIFT := 6
	return float64(d.fd.MeasureString(s)>>SHIFT) + extraDx
}

// Render renders a badge of the given color, with given subject and status to w.
func Render(subject, status string, color Color, w io.Writer) error {
	drawer, err := initDrawer()
	if err != nil {
		return err
	}
	return drawer.Render(subject, status, color, w)
}

// RenderBytes renders a badge of the given color, with given subject and status to bytes.
func RenderBytes(subject, status string, color Color) ([]byte, error) {
	drawer, err := initDrawer()
	if err != nil {
		return nil, err
	}
	return drawer.RenderBytes(subject, status, color)
}

const (
	dpi      = 72
	fontsize = 11
)

var (
	drawer    *badgeDrawer
	initError error
	initOnce  sync.Once
)

func initDrawer() (*badgeDrawer, error) {
	initOnce.Do(func() {
		fd, err := mustNewFontDrawer(fontsize, dpi)
		if err != nil {
			initError = err
			return
		}
		drawer = &badgeDrawer{
			fd:    fd,
			tmpl:  template.Must(template.New("flat-template").Parse(flatTemplate)),
			mutex: &sync.Mutex{},
		}
		initError = nil
	})
	return drawer, initError
}

func mustNewFontDrawer(size, dpi float64) (*font.Drawer, error) {
	ttf, err := truetype.Parse(fonts.VeraSans)
	if err != nil {
		return nil, err
	}
	return &font.Drawer{
		Face: truetype.NewFace(ttf, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}, nil
}

var (
	// Status labels.
	badgeStatusSuccess = "success"
	badgeStatusFailure = "failure"
	badgeStatusStarted = "started"
	badgeStatusError   = "error"
	badgeStatusNone    = "none"
)

func getBadgeStatus(status *model.StatusValue) (string, Color) {
	if status == nil {
		return badgeStatusNone, ColorLightgray
	}

	switch *status {
	case model.StatusSuccess:
		return badgeStatusSuccess, ColorBrightgreen
	case model.StatusFailure:
		return badgeStatusFailure, ColorRed
	case model.StatusPending, model.StatusRunning:
		return badgeStatusStarted, ColorYellow
	case model.StatusError, model.StatusKilled:
		return badgeStatusError, ColorLightgray
	default:
		return badgeStatusNone, ColorLightgray
	}
}

// Generate an SVG badge based on a pipeline.
func Generate(name string, status *model.StatusValue) (string, error) {
	label, color := getBadgeStatus(status)
	bytes, err := RenderBytes(name, label, color)
	if err != nil {
		log.Warn().Err(err).Msg("could not render badge")
		return "", err
	}
	return string(bytes), nil
}

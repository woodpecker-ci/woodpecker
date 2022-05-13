// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package ccmenue

import (
	"encoding/xml"
	"strconv"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// CCMenu displays the build status of projects on a ci server as an item in the Mac's menu bar.
// It started as part of the CruiseControl project that built the first continuous integration server.
//
// http://ccmenu.org/

type CCProjects struct {
	XMLName xml.Name   `xml:"Projects"`
	Project *CCProject `xml:"Project"`
}

type CCProject struct {
	XMLName         xml.Name `xml:"Project"`
	Name            string   `xml:"name,attr"`
	Activity        string   `xml:"activity,attr"`
	LastBuildStatus string   `xml:"lastBuildStatus,attr"`
	LastBuildLabel  string   `xml:"lastBuildLabel,attr"`
	LastBuildTime   string   `xml:"lastBuildTime,attr"`
	WebURL          string   `xml:"webUrl,attr"`
}

func New(r *model.Repo, b *model.Build, link string) *CCProjects {
	proj := &CCProject{
		Name:            r.FullName,
		WebURL:          link,
		Activity:        "Building",
		LastBuildStatus: "Unknown",
		LastBuildLabel:  "Unknown",
	}

	// if the build is not currently running then
	// we can return the latest build status.
	if b.Status != model.StatusPending &&
		b.Status != model.StatusRunning {
		proj.Activity = "Sleeping"
		proj.LastBuildTime = time.Unix(b.Started, 0).Format(time.RFC3339)
		proj.LastBuildLabel = strconv.FormatInt(b.Number, 10)
	}

	// ensure the last build Status accepts a valid
	// ccmenu enumeration
	switch b.Status {
	case model.StatusError, model.StatusKilled:
		proj.LastBuildStatus = "Exception"
	case model.StatusSuccess:
		proj.LastBuildStatus = "Success"
	case model.StatusFailure:
		proj.LastBuildStatus = "Failure"
	}

	return &CCProjects{Project: proj}
}

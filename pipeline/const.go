// Copyright 2023 Woodpecker Authors
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

const (
	ExitCodeKilled int = 137

	// Store no more than 1mb in a log-line as 4mb is the limit of a grpc message
	// and log-lines needs to be parsed by the browsers later on.
	MaxLogLineLength int = 1 * 1024 * 1024 // 1mb

	InternalLabelPrefix string = "woodpecker-ci.org"
	LabelForgeRemoteID  string = InternalLabelPrefix + "/forge-id"
	LabelRepoForgeID    string = InternalLabelPrefix + "/repo-forge-id"
	LabelRepoID         string = InternalLabelPrefix + "/repo-id"
	LabelRepoName       string = InternalLabelPrefix + "/repo-name"
	LabelRepoFullName   string = InternalLabelPrefix + "/repo-full-name"
	LabelBranch         string = InternalLabelPrefix + "/branch"
	LabelOrgID          string = InternalLabelPrefix + "/org-id"
	LabelPlatform       string = InternalLabelPrefix + "/platform"
	LabelHostname       string = InternalLabelPrefix + "/hostname"
	LabelBackend        string = InternalLabelPrefix + "/backend"
)

// ManagedLabels is a list of all labels added automatically to tasks by Woodpecker.
var ManagedLabels = map[string]bool{
	LabelForgeRemoteID: true,
	LabelRepoForgeID:   true,
	LabelRepoID:        true,
	LabelRepoName:      true,
	LabelRepoFullName:  true,
	LabelBranch:        true,
	LabelOrgID:         true,
}

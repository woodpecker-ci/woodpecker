// Copyright 2023 Woodpecker Authors
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

package utils

import "github.com/distribution/reference"

// trimImage returns the short image name without tag.
func trimImage(name string) string {
	ref, err := reference.ParseAnyReference(name)
	if err != nil {
		return name
	}
	named, err := reference.ParseNamed(ref.String())
	if err != nil {
		return name
	}
	named = reference.TrimNamed(named)
	return reference.FamiliarName(named)
}

// expandImage returns the fully qualified image name.
func expandImage(name string) string {
	ref, err := reference.ParseAnyReference(name)
	if err != nil {
		return name
	}
	named, err := reference.ParseNamed(ref.String())
	if err != nil {
		return name
	}
	named = reference.TagNameOnly(named)
	return named.String()
}

// MatchImage returns true if the image name matches
// an image in the list. Note the image tag is not used
// in the matching logic.
func MatchImage(from string, to ...string) bool {
	from = trimImage(from)
	for _, match := range to {
		if from == trimImage(match) {
			return true
		}
	}
	return false
}

// MatchHostname returns true if the image hostname
// matches the specified hostname.
func MatchHostname(image, hostname string) bool {
	ref, err := reference.ParseAnyReference(image)
	if err != nil {
		return false
	}
	named, err := reference.ParseNamed(ref.String())
	if err != nil {
		return false
	}
	if hostname == "index.docker.io" {
		hostname = "docker.io"
	}
	return reference.Domain(named) == hostname
}

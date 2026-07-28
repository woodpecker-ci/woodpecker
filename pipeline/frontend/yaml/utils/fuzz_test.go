// Copyright 2026 Woodpecker Authors
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

import (
	"testing"
)

// FuzzImageMatching exercises container image reference normalization and
// matching (privileged plugin matching, registry hostname matching) with
// untrusted image names from pipeline configs and checks that it never
// panics.
func FuzzImageMatching(f *testing.F) {
	f.Add("golang", "docker.io/library/golang:latest", "docker.io")
	f.Add("codeberg.org/woodpecker-plugins/docker-buildx", "woodpecker-plugins/docker-buildx", "codeberg.org")
	f.Add("image:tag@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "image", "index.docker.io")
	f.Add("REGISTRY.example/Repo/Image:v1", "*", "registry.example")

	f.Fuzz(func(_ *testing.T, from, to, hostname string) {
		_, _ = ParseNamed(from)
		_ = MatchImage(from, to)
		_ = MatchImageDynamic(from, to)
		_ = MatchHostname(from, hostname)
	})
}

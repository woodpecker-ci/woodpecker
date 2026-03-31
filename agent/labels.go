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

package agent

import (
	"fmt"
	"maps"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// buildFilter assembles the rpc.Filter that is passed to the server when
// polling for new workflows. It starts from a set of automatically derived
// labels (hostname, platform, backend, wildcard repo) and then overlays any
// custom labels supplied by the operator, so custom values always win.
func buildFilter(hostname, platform, backendName string, customLabels map[string]string) rpc.Filter {
	labels := map[string]string{
		pipeline.LabelFilterHostname: hostname,
		pipeline.LabelFilterPlatform: platform,
		pipeline.LabelFilterBackend:  backendName,
		pipeline.LabelFilterRepo:     "*", // allow all repos by default
	}
	maps.Copy(labels, customLabels)

	log.Debug().Any("labels", labels).Msg("agent configured with labels")

	return rpc.Filter{Labels: labels}
}

// StringSliceAddToMap parses a slice of "key=value" strings and inserts them
// into m. It is used to convert CLI flag values into the CustomLabels map.
//
// Rules:
//   - Empty strings are silently skipped (utils.StringSliceDeleteEmpty handles this).
//   - "key=value" (and "key=val=ue") are accepted; only the first "=" is the
//     separator so values may themselves contain "=".
//   - A bare "key" with no "=" returns an error.
func StringSliceAddToMap(sl []string, m map[string]string) error {
	if m == nil {
		m = make(map[string]string)
	}
	for _, v := range utils.StringSliceDeleteEmpty(sl) {
		before, after, found := strings.Cut(v, "=")
		switch {
		case before != "" && found:
			m[before] = after
		case before != "":
			return fmt.Errorf("label '%s' has no value — expected format: key=value", before)
		default:
			return fmt.Errorf("empty label string in slice")
		}
	}
	return nil
}

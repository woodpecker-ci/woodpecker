// Copyright 2024 Woodpecker Authors
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

package feedback

import (
	cicd_feedback "github.com/6543/cicd_feedback"
)

func WellKnownResponse() cicd_feedback.WellKnownResponse {
	return cicd_feedback.WellKnownResponse{
		Version:      "v0.1",
		SoftwareName: "woodpecker",
		Type:         "cicd",
	}
}

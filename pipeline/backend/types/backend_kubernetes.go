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

package types

// KubernetesBackendOptions defines all the advanced options for the kubernetes backend
type KubernetesBackendOptions struct {
	Resources          Resources         `json:"resouces,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
}

// Resources defines two maps for kubernetes resource definitions
type Resources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

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

package agent

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

type State struct {
	sync.Mutex `json:"-"`
	Polling    int             `json:"polling_count"`
	Running    int             `json:"running_count"`
	Metadata   map[string]Info `json:"running"`
}

type Info struct {
	ID       string        `json:"id"`
	Repo     string        `json:"repository"`
	Pipeline string        `json:"pipeline_number"`
	Started  time.Time     `json:"pipeline_started"`
	Timeout  time.Duration `json:"pipeline_timeout"`
}

func (s *State) Add(id string, timeout time.Duration, repo, pipeline string) {
	s.Lock()
	s.Polling--
	s.Running++
	s.Metadata[id] = Info{
		ID:       id,
		Repo:     repo,
		Pipeline: pipeline,
		Timeout:  timeout,
		Started:  time.Now().UTC(),
	}
	s.Unlock()
}

func (s *State) Done(id string) {
	s.Lock()
	s.Polling++
	s.Running--
	delete(s.Metadata, id)
	s.Unlock()
}

func (s *State) Healthy() bool {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	buf := time.Hour // 1 hour buffer
	for _, item := range s.Metadata {
		if now.After(item.Started.Add(item.Timeout).Add(buf)) {
			return false
		}
	}
	return true
}

func (s *State) WriteTo(w io.Writer) (int64, error) {
	s.Lock()
	out, _ := json.Marshal(s)
	s.Unlock()
	ret, err := w.Write(out)
	return int64(ret), err
}

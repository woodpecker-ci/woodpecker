package runner

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
	ID      string        `json:"id"`
	Repo    string        `json:"repository"`
	Build   string        `json:"build_number"`
	Started time.Time     `json:"build_started"`
	Timeout time.Duration `json:"build_timeout"`
}

func (s *State) Add(id string, timeout time.Duration, repo, build string) {
	s.Lock()
	s.Polling--
	s.Running++
	s.Metadata[id] = Info{
		ID:      id,
		Repo:    repo,
		Build:   build,
		Timeout: timeout,
		Started: time.Now().UTC(),
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

func (s *State) WriteTo(w io.Writer) (int, error) {
	s.Lock()
	out, _ := json.Marshal(s)
	s.Unlock()
	return w.Write(out)
}

package builder

import (
	"sort"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type Item struct {
	Workflow  *Workflow
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Pending   bool
	Config    *backend_types.Config
}

type Workflow struct {
	ID      int64             `json:"id"`
	PID     int               `json:"pid"`
	Name    string            `json:"name"`
	Environ map[string]string `json:"environ,omitempty"`
	AxisID  int               `json:"-"`
}

type YamlFile struct {
	Name string
	Data []byte
}

type yamlFileList []*YamlFile

func (a yamlFileList) Len() int           { return len(a) }
func (a yamlFileList) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a yamlFileList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func SortYamlFilesByName(fm []*YamlFile) []*YamlFile {
	l := yamlFileList(fm)
	sort.Sort(l)
	return l
}

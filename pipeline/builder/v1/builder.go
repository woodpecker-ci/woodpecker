package v1

import (
	"fmt"
	"sort"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type RawConfig struct {
	Name string
	Data string
}

type Pipeline struct {
	Input struct {
		RawConfigs []*RawConfig
		Metadata   *frontend.Metadata // TODO: create custom subset type excluding data of pipeline step
		// Repo         *model.Repo
		// User         *model.User
		// CurrentBuild *model.Build
		// LastBuild    *model.Build
		// Secrets      []*model.Secret
		// Registries   []*model.Registry
		// Netrc        *model.Netrc
		// Environment  map[string]string
	}
	Output struct {
		// main proc => pipeline-config => matrix-axis => pipeline-steps
		Procs  []*model.Proc
		Config *backend.Config
	}
}

func New(repo *model.Repo, user *model.User) *Pipeline {
	return &Pipeline{}
}

func (p *Pipeline) Compile(rawConfigs []*RawConfig) error {
	var items []*BuildItem

	sort.Sort(remote.ByName(rawConfigs))

	pidSequence := 1

	for _, y := range b.RawConfigs {
		// matrix axes
		axes, err := matrix.ParseString(string(y.Data))
		if err != nil {
			return err
		}
		if len(axes) == 0 {
			axes = append(axes, matrix.Axis{})
		}

		for _, axis := range axes {
			proc := &model.Proc{
				BuildID: b.Curr.ID,
				PID:     pidSequence,
				PGID:    pidSequence,
				State:   model.StatusPending,
				Environ: axis,
				Name:    sanitizePipelinePath(y.Name),
			}

			metadata := metadataFromStruct(b.Repo, b.Curr, b.Last, proc, b.Link)
			environ := b.environmentVariables(metadata, axis)

			// substitute vars
			substituted, err := b.envsubst(string(y.Data), environ)
			if err != nil {
				return nil, err
			}

			// parse yaml pipeline
			parsed, err := yaml.ParseString(substituted)
			if err != nil {
				return nil, err
			}

			// lint pipeline
			if err := linter.New(
				linter.WithTrusted(b.Repo.IsTrusted),
			).Lint(parsed); err != nil {
				return nil, err
			}

			if !parsed.Branches.Match(b.Curr.Branch) && (b.Curr.Event != model.EventDeploy && b.Curr.Event != model.EventTag) {
				proc.State = model.StatusSkipped
			}

			metadata.SetPlatform(parsed.Platform)

			ir := b.toInternalRepresentation(parsed, environ, metadata, proc.ID)

			if len(ir.Stages) == 0 {
				continue
			}

			item := &BuildItem{
				Proc:      proc,
				Config:    ir,
				Labels:    parsed.Labels,
				DependsOn: parsed.DependsOn,
				RunsOn:    parsed.RunsOn,
				Platform:  metadata.Sys.Arch,
			}
			if item.Labels == nil {
				item.Labels = map[string]string{}
			}

			items = append(items, item)
			pidSequence++
		}
	}

	items = filterItemsWithMissingDependencies(items)

	// check if at least one proc can start, if list is not empty
	procListContainsItemsToRun(items)
	if len(items) > 0 && !procListContainsItemsToRun(items) {
		return fmt.Errorf("build has no startpoint")
	}

	return nil
}

func (p *Pipeline) GetProcs() []*model.Proc {
	return nil
}

func (p *Pipeline) GetBackendConfig() *backend.Config {
	return nil
}

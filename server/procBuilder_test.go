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

package server

import (
	"fmt"
	"testing"

	"github.com/laszlocph/drone-oss-08/model"
	"github.com/laszlocph/drone-oss-08/remote"
)

func TestMultilineEnvsubst(t *testing.T) {
	b := procBuilder{
		Repo: &model.Repo{},
		Curr: &model.Build{
			Message: `aaa
bbb`,
		},
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  xxx:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
			&remote.FileMeta{Data: []byte(`
pipeline:
  build:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
		}}

	if buildItems, err := b.Build(); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(buildItems)
	}
}

func TestMultiPipeline(t *testing.T) {
	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  &model.Build{},
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  xxx:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
			&remote.FileMeta{Data: []byte(`
pipeline:
  build:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
		},
	}

	buildItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(buildItems) != 2 {
		t.Fatal("Should have generated 2 buildItems")
	}
}

func TestDependsOn(t *testing.T) {
	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  &model.Build{},
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  deploy:
    image: scratch

depends_on:
  - lint
  - test
  - build
`)},
		},
	}

	buildItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(buildItems[0].DependsOn) != 3 {
		t.Fatal("Should have 3 dependencies")
	}
	if buildItems[0].DependsOn[1] != "test" {
		t.Fatal("Should depend on test")
	}
}

func TestRunsOn(t *testing.T) {
	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  &model.Build{},
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  deploy:
    image: scratch

runs_on:
  - success
  - failure
`)},
		},
	}

	buildItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(buildItems[0].RunsOn) != 2 {
		t.Fatal("Should run on success and failure")
	}
	if buildItems[0].RunsOn[1] != "failure" {
		t.Fatal("Should run on failure")
	}
}

func TestBranchFilter(t *testing.T) {
	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  &model.Build{Branch: "dev"},
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  xxx:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
branches: master
`)},
			&remote.FileMeta{Data: []byte(`
pipeline:
  build:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
		},
	}

	buildItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(buildItems) != 2 {
		t.Fatal("Should have generated 2 buildItems")
	}
	if buildItems[0].Proc.State != model.StatusSkipped {
		t.Fatal("Should not run on dev branch")
	}
	for _, child := range buildItems[0].Proc.Children {
		if child.State != model.StatusSkipped {
			t.Fatal("Children should skipped status too")
		}
	}
	if buildItems[1].Proc.State != model.StatusPending {
		t.Fatal("Should run on dev branch")
	}
}

func TestZeroSteps(t *testing.T) {
	build := &model.Build{Branch: "dev"}

	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  build,
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
skip_clone: true
pipeline:
  build:
    when:
      branch: notdev
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
		},
	}

	buildItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(buildItems) != 0 {
		t.Fatal("Should not generate a build item if there are no steps")
	}
	if len(build.Procs) != 0 {
		t.Fatal("Should not generate a build item if there are no steps")
	}
}

func TestTree(t *testing.T) {
	build := &model.Build{}

	b := procBuilder{
		Repo:  &model.Repo{},
		Curr:  build,
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: []*remote.FileMeta{
			&remote.FileMeta{Data: []byte(`
pipeline:
  build:
    image: scratch
    yyy: ${DRONE_COMMIT_MESSAGE}
`)},
		},
	}

	_, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(build.Procs) != 3 {
		t.Fatal("Should generate three in total")
	}
	if build.Procs[1].PPID != 1 {
		t.Fatal("Clone step should be a children of the stage")
	}
	if build.Procs[2].PPID != 1 {
		t.Fatal("Build step should be a children of the stage")
	}
}

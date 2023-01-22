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

package lxc

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

var scriptTemplate = template.Must(template.New("script").Parse(`#!/bin/sh -xe
lxc-create --name="{{.Name}}" --template={{.Template}} -- --release {{.Release}} $packages
tee -a /var/lib/lxc/{{.Name}}/config <<'EOF'                                                                                                                   
security.nesting = true
lxc.cap.drop =
lxc.apparmor.profile = unconfined
EOF

mkdir /var/lib/lxc/{{.Name}}/rootfs/woodpecker
mount --bind {{.Workspace}} /var/lib/lxc/{{.Name}}/rootfs/woodpecker

mkdir /var/lib/lxc/{{.Name}}/rootfs/rundir
mount --bind {{.RunDir}} /var/lib/lxc/{{.Name}}/rootfs/rundir

lxc-start {{.Name}}
lxc-wait --name {{.Name}} --state RUNNING
lxc-attach --name {{.Name}} -- /rundir/networking.sh
lxc-attach --name {{.Name}} -- /bin/sh -c 'cd "/woodpecker/{{ .Repo }}" && /bin/sh -ex /rundir/{{ .Script }}'
`))

func writeScript(t *template.Template, config any, script string) error {
	f, err := os.Create(script)
	if err != nil {
		return err
	}
	if err := os.Chmod(script, 0o755); err != nil {
		return err
	}
	if err := t.Execute(f, config); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

type lxc struct {
	cmd       *exec.Cmd
	output    io.ReadCloser
	rundir    string
	workspace string
	name      string
}

// make sure lxc implements Engine
var _ types.Engine = &lxc{}

// New returns a new lxc Engine.
func New() types.Engine {
	return &lxc{}
}

func (e *lxc) Name() string {
	return "lxc"
}

func (e *lxc) IsAvailable() bool {
	return true
}

func (e *lxc) Load() error {
	dir, err := os.MkdirTemp("", "woodpecker-lxc-*")
	e.rundir = dir
	e.name = path.Base(dir)
	return err
}

var serviceHostnameTemplate = template.Must(template.New("hostnames").Parse(`#!/bin/sh -ex
#
# Wait until service containers get an IP and set /etc/hosts with their name
#
cat /rundir/service-alias | while read name alias ; do
  for d in $(seq 60); do
    getent hosts $name > /dev/null && break
    sleep 1
  done
  echo $(getent hosts $name) $alias
done | tee -a /etc/hosts
#
# Wait until internet connectivity is ready
#
for d in $(seq 60); do
  getent hosts wikipedia.org > /dev/null && break
  sleep 1
done
getent hosts wikipedia.org
`))

func (e *lxc) ContainerName(name string) string {
	return e.name + "-" + strings.ReplaceAll(name, "_", "")
}

func (e *lxc) Setup(ctx context.Context, config *types.Config) error {
	e.workspace = e.rundir + "/workspace"
	log.Debug().Msgf("config %d %+v", len(config.Volumes), config.Volumes[0])
	if err := writeScript(serviceHostnameTemplate, struct{}{}, e.rundir+"/networking.sh"); err != nil {
		log.Error().Err(err)
		return err
	}
	f, err := os.Create(e.rundir + "/service-alias")
	if err != nil {
		log.Error().Err(err)
		return err
	}
	for _, stage := range config.Stages {
		if stage.Alias != "services" {
			continue
		}
		for _, step := range stage.Steps {
			if _, err := f.WriteString(fmt.Sprintf("%s %s\n", e.ContainerName(step.Name), step.Alias)); err != nil {
				log.Error().Err(err)
				return err
			}
		}
	}
	return f.Close()
}

var (
	acceptable       = "^[a-zA-Z0-9]+$"
	acceptableRegexp = regexp.MustCompile(acceptable)
)

// Exec the pipeline step.
func (e *lxc) Exec(ctx context.Context, step *types.Step) error {
	var env []string
	for a, b := range step.Environment {
		env = append(env, a+"="+b)
	}
	env = append(env, "PATH="+os.Getenv("PATH"))

	defaultCloneImage := strings.Split(constant.DefaultCloneImage, ":")
	if len(defaultCloneImage) != 2 {
		err := fmt.Errorf("%s does not split in two but in %v", constant.DefaultCloneImage, defaultCloneImage)
		log.Error().Err(err)
		return err
	}
	log.Debug().Msgf("Step %+v", step)
	if strings.HasPrefix(step.Image, defaultCloneImage[0]) {
		env = append(env, "CI_WORKSPACE="+e.workspace+"/"+step.Environment["CI_REPO"])
		e.cmd = exec.CommandContext(ctx, "plugin-git")
		e.cmd.Env = env
		e.cmd.Dir = e.workspace + "/" + step.Environment["CI_REPO_OWNER"]
	} else {
		image := strings.Split(step.Image, ":")
		if len(image) != 2 {
			err := fmt.Errorf("step image %s does not split in two but in %v", step.Image, image)
			log.Error().Err(err)
			return err
		}
		for _, s := range image {
			if !acceptableRegexp.MatchString(s) {
				err := fmt.Errorf("in image name %s, %s does not match %s", step.Image, s, acceptable)
				log.Error().Err(err)
				return err
			}
		}
		template := image[0]
		release := image[1]
		log.Debug().Msgf("template %s release %s", template, release)
		script := e.rundir + "/" + step.Name
		if err := writeScript(scriptTemplate, struct {
			Name      string
			Template  string
			Release   string
			Repo      string
			Workspace string
			RunDir    string
			Script    string
		}{
			Name:      e.ContainerName(step.Name),
			Template:  template,
			Release:   release,
			Repo:      step.Environment["CI_REPO"],
			Workspace: e.workspace,
			RunDir:    e.rundir,
			Script:    "commands-" + step.Name,
		}, script); err != nil {
			log.Error().Err(err)
			return err
		}
		var command []string
		command = append(command, script)

		if err := os.WriteFile(e.rundir+"/"+"commands-"+step.Name, []byte(strings.Join(step.Commands, "\n")), 0o755); err != nil {
			log.Error().Err(err)
			return err
		}

		e.cmd = exec.CommandContext(ctx, command[0], command[1:]...)
		e.cmd.Env = env
		e.cmd.Dir = e.workspace + "/" + step.Environment["CI_REPO"]
	}

	log.Debug().Msgf("Working directory %v", e.cmd.Dir)
	err := os.MkdirAll(e.cmd.Dir, 0o700)
	if err != nil {
		return err
	}
	e.output, _ = e.cmd.StdoutPipe()
	e.cmd.Stderr = e.cmd.Stdout

	return e.cmd.Start()
}

func (e *lxc) Wait(context.Context, *types.Step) (*types.State, error) {
	err := e.cmd.Wait()
	ExitCode := 0
	if eerr, ok := err.(*exec.ExitError); ok {
		ExitCode = eerr.ExitCode()
		err = nil
	}
	return &types.State{
		Exited:   true,
		ExitCode: ExitCode,
	}, err
}

func (e *lxc) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	return e.output, nil
}

var destroyTemplate = template.Must(template.New("destroy").Parse(`#!/bin/sh -x
lxc-ls -1 --filter="^{{.Name}}" | while read container ; do
   lxc-stop --kill --name="$container"
   umount "/var/lib/lxc/$container/rootfs/woodpecker"
   umount "/var/lib/lxc/$container/rootfs/rundir"
   lxc-destroy --force --name="$container"
done
`))

func (e *lxc) Destroy(ctx context.Context, conf *types.Config) error {
	script := e.rundir + "/destroy.sh"
	if err := writeScript(destroyTemplate, struct {
		Name string
	}{
		Name: e.name,
	}, script); err != nil {
		log.Error().Err(err)
		return err
	}
	cmd := exec.CommandContext(ctx, script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg(string(output))
		return err
	}
	if len(output) > 0 {
		log.Debug().Msg(string(output))
	}
	return os.RemoveAll(e.workspace)
}

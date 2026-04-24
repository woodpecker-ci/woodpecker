// Copyright 2022 Woodpecker Authors
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

package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"codeberg.org/6543/xyaml"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/cli/lint"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/local"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// Command exports the exec command.
var Command = &cli.Command{
	Name:      "exec",
	Usage:     "execute a local pipeline",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    run,
	Flags:     slices.Concat(flags, docker.Flags, kubernetes.Flags, local.Flags),
}

var backends = []backend_types.Backend{
	kubernetes.New(),
	docker.New(),
	local.New(),
}

func run(ctx context.Context, c *cli.Command) error {
	return common.RunPipelineFunc(ctx, c, execFile, execDir)
}

func execDir(ctx context.Context, c *cli.Command, dir string) error {
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(dir))
	}
	if runtime.GOOS == "windows" && c.String("backend-engine") != "local" {
		repoPath = convertPathForWindows(repoPath)
	}

	var yamls []*builder.YamlFile
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.Mode().IsRegular() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			dat, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			yamls = append(yamls, &builder.YamlFile{Name: path, Data: dat})
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	return runExec(ctx, c, yamls, repoPath)
}

func execFile(ctx context.Context, c *cli.Command, file string) error {
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(file))
	}
	if runtime.GOOS == "windows" && c.String("backend-engine") != "local" {
		repoPath = convertPathForWindows(repoPath)
	}

	dat, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return runExec(ctx, c, []*builder.YamlFile{{Name: file, Data: dat}}, repoPath)
}

func runExec(ctx context.Context, c *cli.Command, yamls []*builder.YamlFile, repoPath string) error {
	// if we use the local backend we should signal to run at $repoPath
	if c.String("backend-engine") == "local" {
		local.CLIWorkaroundExecAtDir = repoPath
	}

	// collect secrets from flags
	var secrets []compiler.Secret
	for key, val := range c.StringMap("secrets") {
		secrets = append(secrets, compiler.Secret{Name: key, Value: val})
	}
	if secretsFile := c.String("secrets-file"); secretsFile != "" {
		fileContent, err := os.ReadFile(secretsFile)
		if err != nil {
			return err
		}
		var m map[string]string
		if err := xyaml.Unmarshal(fileContent, &m); err != nil {
			return err
		}
		for key, val := range m {
			secrets = append(secrets, compiler.Secret{Name: key, Value: val})
		}
	}

	// collect extra env vars from --env flags
	pipelineEnv := make(map[string]string)
	for _, env := range c.StringSlice("env") {
		before, after, _ := strings.Cut(env, "=")
		pipelineEnv[before] = after
	}

	privilegedPlugins := c.StringSlice("plugins-privileged")

	// NOTE: we deliberately do NOT set compiler.WithPrefix here.
	// The pipeline builder (pipeline/frontend/builder) generates a
	// unique prefix per workflow of the form wp_<ULID>_<workflowID>,
	// which becomes the workflow's docker network and volume name.
	// Passing a shared WithPrefix would override that per-workflow
	// value and cause parallel workflows to collide on the same
	// docker network/volume — the exact symptom that appeared when
	// the scheduler started running workflows concurrently.

	// build compiler options — mirrors server behavior
	compilerOpts := []compiler.Option{
		compiler.WithEscalated(privilegedPlugins...),
		compiler.WithNetworks(c.StringSlice("network")...),
		compiler.WithProxy(compiler.ProxyOptions{
			NoProxy:    c.String("backend-no-proxy"),
			HTTPProxy:  c.String("backend-http-proxy"),
			HTTPSProxy: c.String("backend-https-proxy"),
		}),
		compiler.WithLocal(c.Bool("local")),
		compiler.WithNetrc(
			c.String("netrc-username"),
			c.String("netrc-password"),
			c.String("netrc-machine"),
		),
		compiler.WithSecret(secrets...),
		compiler.WithEnviron(pipelineEnv),
	}

	// configure volumes for local execution
	volumes := c.StringSlice("volumes")
	compilerOpts = append(compilerOpts,
		compiler.WithWorkspace(
			c.String("workspace-base"),
			c.String("workspace-path"),
		),
	)
	if c.Bool("local") {
		// In local mode we bind-mount the user's repo directory into
		// each step so the step sees the working tree as-is instead
		// of a cloned copy. The per-workflow workspace volume mount
		// (<prefix>_default:<workspace-base>) is added later, after
		// the builder has assigned each workflow its own prefix —
		// see injectLocalWorkspaceMounts below.
		volumes = append(volumes,
			repoPath+":"+c.String("workspace-base")+"/"+c.String("workspace-path"),
		)
	}
	compilerOpts = append(compilerOpts, compiler.WithVolumes(volumes...))

	// build the metadata once — the CLI has a single pipeline context for all
	// workflows, so every workflow gets the same metadata.
	baseMetadata, err := metadataFromContext(ctx, c, nil)
	if err != nil {
		return fmt.Errorf("could not create metadata: %w", err)
	}

	b := builder.PipelineBuilder{
		Yamls: yamls,
		Envs:  pipelineEnv,
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  c.Bool("repo-trusted-network"),
			Volumes:  c.Bool("repo-trusted-volumes"),
			Security: c.Bool("repo-trusted-security"),
		},
		TrustedClonePlugins: constant.TrustedClonePlugins,
		PrivilegedPlugins:   privilegedPlugins,
		CompilerOptions:     compilerOpts,
		// GetWorkflowMetadata provides per-workflow metadata. In the CLI there
		// is no server context, so we derive it from the base metadata and
		// populate the workflow name/matrix from the builder.Workflow.
		GetWorkflowMetadata: func(w *builder.Workflow) metadata.Metadata {
			m := *baseMetadata
			m.Workflow = metadata.Workflow{
				Name:   w.Name,
				Number: w.PID,
				Matrix: w.Environ,
			}
			return m
		},
	}

	items, buildErr := b.Build()

	// Decide output mode up front. We need this before printing any
	// warnings: in TUI mode the warnings must be captured into the
	// model's messages ring so they render in the bottom pane,
	// rather than being smeared across the terminal right before the
	// alt-screen swap wipes them.
	useTUI := !c.Bool("no-tui") && logger.IsInteractiveTerminal()

	// preRunMessages collects pre-run diagnostic text (lint
	// warnings, "Config is valid" banners, etc.) destined for the
	// TUI messages pane. In line mode this stays empty and the
	// output goes to stdout as before.
	var preRunMessages strings.Builder

	if buildErr != nil {
		str, fmtErr := lint.FormatLintError("pipeline", buildErr, false)
		if useTUI {
			preRunMessages.WriteString(str)
		} else {
			fmt.Print(str)
		}
		if fmtErr != nil {
			return fmtErr
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("no workflows to execute (all filtered out)")
	}

	// Local mode: mount each workflow's docker volume into every
	// step's workspace path. This used to be done globally via
	// compiler.WithVolumes with a shared prefix, but with parallel
	// workflows that collided on the same docker volume name — all
	// workflows used "<shared-prefix>_default". The builder now
	// generates a per-workflow prefix, so we injecting the mount
	// here after the build gives each workflow its own volume.
	if c.Bool("local") {
		injectLocalWorkspaceMounts(items, c.String("workspace-base"))
	}

	backendCtx := context.WithValue(ctx, backend_types.CliCommand, c)
	backendEngine, err := backend.FindBackend(backendCtx, backends, c.String("backend-engine"))
	if err != nil {
		return err
	}
	if _, err = backendEngine.Load(backendCtx); err != nil {
		return err
	}

	// The pipeline context carries timeout + SIGTERM cancellation for
	// the entire DAG run. Every workflow's runtime derives its own ctx
	// from this one, so cancellation fans out to all of them at once.
	pipelineCtx, cancel := context.WithTimeout(ctx, c.Duration("timeout"))
	defer cancel()

	if useTUI {
		return runTUIMode(pipelineCtx, items, backendEngine, preRunMessages.String())
	}
	return runLineMode(pipelineCtx, items, backendEngine)
}

// runLineMode drives the scheduler with a line-oriented output path:
// per-step output goes through LineWriter to stderr, and workflow
// banners / diagnostics are rendered by handleLineModeEvent.
//
// This is the path used when --no-tui is set, when stdout is not a
// terminal (e.g. CI logs), or as a fallback when the TUI is
// unavailable.
func runLineMode(pipelineCtx context.Context, items []*builder.Item, backendEngine backend_types.Backend) error {
	pipelineCtx = utils.WithContextSigtermCallback(pipelineCtx, func() {
		fmt.Fprintln(os.Stderr, "ctrl+c received, terminating pipeline")
	})

	// Whether to emit workflow names in the per-step log prefix. With
	// a single workflow the prefix stays terse as "[step]"; with
	// multiple workflows running in parallel, interleaved output needs
	// the workflow qualifier to stay attributable.
	multiWorkflow := len(items) > 1

	// Per-workflow logger factory. The runtime calls this once per
	// step with an io.ReadCloser streaming that step's stdout+stderr;
	// we pipe each line through the workflow-aware LineWriter.
	newLogger := func(workflowName string) logging.Logger {
		return logging.Logger(func(step *backend_types.Step, rc io.ReadCloser) error {
			var lw io.WriteCloser
			if multiWorkflow {
				lw = NewWorkflowLineWriter(workflowName, step.Name, step.UUID)
			} else {
				lw = NewLineWriter(step.Name, step.UUID)
			}
			return pipeline_utils.CopyLineByLine(lw, rc, pipeline.MaxLogLineLength)
		})
	}

	// Events channel: consumed by a goroutine that turns scheduler
	// state transitions into user-visible banners and diagnostics.
	// Buffered generously so a slow terminal never back-pressures the
	// scheduler's control loop.
	events := make(chan scheduler.Event, schedulerEventBuffer)
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range events {
			handleLineModeEvent(os.Stderr, ev)
		}
	}()

	runFunc := func(runCtx context.Context, item *builder.Item) error {
		return pipeline_runtime.New(item.Config, backendEngine,
			pipeline_runtime.WithContext(runCtx),
			pipeline_runtime.WithTracer(tracing.DefaultTracer),
			pipeline_runtime.WithLogger(newLogger(item.Workflow.Name)),
			pipeline_runtime.WithDescription(map[string]string{
				"CLI": "exec",
			}),
		).Run(runCtx)
	}

	sched := scheduler.New(scheduler.Options{
		Items:  items,
		Run:    runFunc,
		Events: events,
	})

	execErr := sched.Run(pipelineCtx)
	<-eventsDone
	return execErr
}

// schedulerEventBuffer is the channel buffer size for scheduler
// events. Generous so a slow consumer (terminal, tea program) does
// not back-pressure the scheduler's control loop.
const schedulerEventBuffer = 64

// handleLineModeEvent renders a workflow-level state transition to
// the given writer for the plain (non-TUI) output path. It emits:
//
//   - a "# <name>" banner when a workflow starts running, matching
//     the legacy sequential output,
//   - a short diagnostic line when a workflow is blocked by a failed
//     dependency (so the user understands the skip),
//   - nothing for other states — per-step output and the final error
//     return already cover success/failure reporting.
func handleLineModeEvent(out io.Writer, ev scheduler.Event) {
	switch ev.State {
	case scheduler.StateRunning:
		WorkflowHeader(out, ev.Workflow)
	case scheduler.StateBlocked:
		if ev.Err != nil {
			fmt.Fprintf(out, "# %s: %s\n", ev.Workflow, ev.Err.Error())
		}
	case scheduler.StateCanceled:
		fmt.Fprintf(out, "# %s: canceled\n", ev.Workflow)
	}
}

// injectLocalWorkspaceMounts adds the per-workflow workspace volume
// binding to every step in every item. In local-mode runs (the
// default when invoking `woodpecker-cli exec` directly), steps need
// to share a named docker volume for the workspace so files written
// by one step are visible to the next; the volume itself is created
// by the backend in SetupWorkflow using the name in item.Config.Volume.
//
// Previously the CLI computed one shared prefix upfront and added
// "<prefix>_default:<workspace-base>" to compiler.WithVolumes(),
// which applied to all workflows. That worked when exec ran
// workflows sequentially but collided on the first parallel run:
// every workflow tried to create the same docker volume and docker
// network, producing "already exists" and "unknown network" errors.
//
// Now the builder emits a unique prefix per workflow (see
// pipeline/frontend/builder/builder.go). We read the per-workflow
// volume name off each item's compiled Config and inject the binding
// into every step's Volumes slice. Per-step injection matches what
// compiler.WithVolumes already does internally for the non-local
// path, so the runtime sees an identical shape either way.
func injectLocalWorkspaceMounts(items []*builder.Item, workspaceBase string) {
	for _, item := range items {
		if item == nil || item.Config == nil || item.Config.Volume == "" {
			continue
		}
		mount := item.Config.Volume + ":" + workspaceBase
		for _, stage := range item.Config.Stages {
			for _, step := range stage.Steps {
				step.Volumes = append(step.Volumes, mount)
			}
		}
	}
}

// convertPathForWindows converts a path to use slash separators
// for Windows. If the path is a Windows volume name like C:, it
// converts it to an absolute root path starting with slash (e.g.
// C: -> /c). Otherwise it just converts backslash separators to
// slashes.
func convertPathForWindows(path string) string {
	base := filepath.VolumeName(path)

	// Check if path is volume name like C:
	//nolint:mnd
	if len(base) == 2 {
		path = strings.TrimPrefix(path, base)
		base = strings.ToLower(strings.TrimSuffix(base, ":"))
		return "/" + base + filepath.ToSlash(path)
	}

	return filepath.ToSlash(path)
}

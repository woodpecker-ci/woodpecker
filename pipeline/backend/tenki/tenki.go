// Copyright 2026 Woodpecker Authors
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

// Package tenki implements a Woodpecker execution backend that runs each
// workflow inside a Tenki sandbox (an ephemeral microVM) and each step as a
// command executed within that sandbox. A single sandbox is created per
// workflow so that all of its steps share the same filesystem (the workspace),
// mirroring how the Docker backend shares a volume across step containers.
package tenki

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/TenkiCloud/tenki-sdk-go/sandbox"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const EngineName = "tenki"

// orphanReapTimeout bounds the best-effort cleanup of a sandbox left behind by
// a failed SetupWorkflow.
const orphanReapTimeout = 30 * time.Second

// tenki is the Tenki sandbox backend. A single instance handles multiple
// concurrent workflows, keyed by taskUUID.
type tenki struct {
	client    *sandbox.Client
	config    config
	workflows sync.Map // taskUUID -> *workflowState
}

// workflowState holds the sandbox backing a single workflow and the state of
// its currently running steps.
type workflowState struct {
	session   *sandbox.Session
	stepState sync.Map // step.UUID -> *stepState
}

// stepState holds the running command of a single step.
type stepState struct {
	handle *sandbox.RunHandle
	output io.ReadCloser
}

// New returns a new Tenki sandbox Backend.
func New() backend_types.Backend {
	return &tenki{}
}

func (e *tenki) Name() string {
	return EngineName
}

func (e *tenki) IsAvailable(ctx context.Context) bool {
	if c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command); ok {
		return c.String("backend-tenki-api-key") != ""
	}
	return false
}

func (e *tenki) Flags() []cli.Flag {
	return Flags
}

// Load initializes the Tenki client from the parsed CLI flags.
func (e *tenki) Load(ctx context.Context) (*backend_types.BackendInfo, error) {
	c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command)
	if !ok {
		return nil, backend_types.ErrNoCliContextFound
	}

	conf, err := configFromCli(c)
	if err != nil {
		return nil, err
	}
	e.config = conf

	opts := []sandbox.Option{sandbox.WithAuthToken(conf.apiKey)}
	if conf.endpoint != "" {
		opts = append(opts, sandbox.WithBaseURL(conf.endpoint))
	}

	client, err := sandbox.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create tenki client: %w", err)
	}
	e.client = client

	// The Go SDK's Create requires an explicit project scope (unlike the CLI,
	// which tracks a "current project"). Resolve it from the API key identity
	// when it was not configured explicitly.
	if e.config.projectID == "" {
		if err := e.resolveProject(ctx); err != nil {
			return nil, err
		}
	}

	return &backend_types.BackendInfo{
		// Tenki sandboxes run on a standard linux/amd64 base image.
		Platform: "linux/amd64",
	}, nil
}

// resolveProject fills config.projectID (and workspaceID) by inspecting the
// authenticated identity: it picks the configured workspace, or the first one,
// and that workspace's first project.
func (e *tenki) resolveProject(ctx context.Context) error {
	identity, err := e.client.WhoAmI(ctx)
	if err != nil {
		return fmt.Errorf("could not resolve tenki identity: %w", err)
	}

	workspaceID, projectID, ok := selectProject(identity.Workspaces, e.config.workspaceID)
	if !ok {
		return ErrNoProjectResolved
	}
	e.config.workspaceID = workspaceID
	e.config.projectID = projectID
	log.Debug().
		Str("workspace", workspaceID).
		Str("project", projectID).
		Msg("resolved tenki workspace/project from API key identity; set backend-tenki-project-id to override")
	return nil
}

// selectProject picks a workspace and its first project from the identity.
// When workspaceID is set, only that workspace is considered; otherwise the
// first workspace that has at least one project wins.
func selectProject(workspaces []sandbox.IdentityWorkspace, workspaceID string) (ws, project string, ok bool) {
	for _, w := range workspaces {
		if workspaceID != "" && w.ID != workspaceID {
			continue
		}
		if len(w.Projects) == 0 {
			continue
		}
		return w.ID, w.Projects[0].ID, true
	}
	return "", "", false
}

// SetupWorkflow creates one sandbox for the whole workflow. All steps of the
// workflow run inside this sandbox and therefore share its filesystem.
func (e *tenki) SetupWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	createOpts := []sandbox.CreateOption{
		sandbox.WithWaitReady(true),
		sandbox.WithMaxDuration(e.config.maxDuration),
		sandbox.WithAllowOutbound(e.config.allowOutbound),
		sandbox.WithProjectID(e.config.projectID),
		// Name and tag the sandbox so it is identifiable in the Tenki
		// dashboard/CLI (mirrors how docker/k8s name their containers/pods).
		sandbox.WithName(sandboxName(taskUUID)),
		sandbox.WithMetadata(workflowMetadata(conf, taskUUID)),
		// Intentionally no WithImage: use the standard Tenki base image and
		// avoid the (less stable) template/snapshot feature.
	}
	if e.config.workspaceID != "" {
		createOpts = append(createOpts, sandbox.WithWorkspaceID(e.config.workspaceID))
	}

	session, err := e.client.CreateAndWait(ctx, e.config.createTimeout, createOpts...)
	if err != nil {
		// CreateAndWait may have provisioned the sandbox before readiness failed
		// (or the context was canceled), dropping the handle. Reap it best-effort
		// so it does not linger until maxDuration.
		e.reapOrphanSandbox(ctx, taskUUID)
		return fmt.Errorf("could not create tenki sandbox: %w", err)
	}

	e.workflows.Store(taskUUID, &workflowState{session: session})
	return nil
}

// reapOrphanSandbox best-effort terminates a sandbox that CreateAndWait may have
// provisioned before failing (whose handle was therefore lost). It matches by
// the deterministic sandbox name so it never touches unrelated sandboxes, and
// uses its own short-lived context so it still runs when the workflow context
// was canceled.
func (e *tenki) reapOrphanSandbox(ctx context.Context, taskUUID string) {
	name := sandboxName(taskUUID)

	// Detach from the (possibly already-canceled) parent so the reap still runs,
	// while keeping its values, then bound it with our own timeout.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), orphanReapTimeout)
	defer cancel()

	sessions, err := e.client.ListProjectSandboxes(ctx, e.config.projectID)
	if err != nil {
		log.Warn().Err(err).Str("taskUUID", taskUUID).Msg("could not list sandboxes to reap possible orphan")
		return
	}
	for _, s := range sessions {
		if s.Name != name {
			continue
		}
		if err := s.CloseIfOpen(ctx); err != nil {
			log.Warn().Err(err).Str("sandbox", s.ID).Msg("could not reap orphan tenki sandbox")
			continue
		}
		log.Info().Str("taskUUID", taskUUID).Str("sandbox", s.ID).Msg("reaped orphan tenki sandbox")
	}
}

// StartStep runs the step's commands as a single shell invocation inside the
// workflow's sandbox and starts streaming its output.
func (e *tenki) StartStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	ws, err := e.getWorkflowState(taskUUID)
	if err != nil {
		return err
	}

	switch step.Type {
	case backend_types.StepTypeCommands, backend_types.StepTypeClone:
		// supported: plain command execution against the base image
	default:
		// TODO: plugins (own image) and services (long-lived + exposed port)
		// are not yet supported by this backend.
		return fmt.Errorf("%w: %s", ErrUnsupportedStepType, step.Type)
	}

	// The default clone step and plugins run via an image entrypoint rather than
	// commands. Since this backend ignores step.Image, such a step would produce
	// an empty `sh -c` that exits 0 and would falsely report success. Fail loudly
	// instead of silently doing nothing.
	if len(step.Commands) == 0 {
		return fmt.Errorf("%w: step %q (image-based clone/plugin steps are not supported)", ErrNoCommands, step.Name)
	}

	// Woodpecker points steps at a workspace directory (e.g. /woodpecker/src)
	// that does not exist in a fresh sandbox. Create it before running, since
	// the guest fails an exec whose cwd does not exist.
	if step.WorkingDir != "" {
		if err := ensureWorkspaceDir(ctx, ws.session, step.WorkingDir); err != nil {
			return fmt.Errorf("could not create working dir %q for step %q: %w", step.WorkingDir, step.Name, err)
		}
	}

	script := strings.Join(step.Commands, "\n")
	cmd := ws.session.Command([]string{"/bin/sh", "-c", script}, sandbox.RunOptions{
		Env: step.Environment,
		Dir: step.WorkingDir,
	})

	handle, err := cmd.Stream(ctx)
	if err != nil {
		return fmt.Errorf("could not start step %q: %w", step.Name, err)
	}

	// Steps never write to stdin. Close the write side so the SDK's stdin pump
	// goroutine exits instead of blocking for the lifetime of the (long-lived)
	// agent process.
	if handle.Stdin != nil {
		_ = handle.Stdin.Close()
	}

	ws.stepState.Store(step.UUID, &stepState{
		handle: handle,
		output: mergeOutput(handle),
	})
	return nil
}

// TailStep returns the merged stdout/stderr stream of the step.
func (e *tenki) TailStep(_ context.Context, step *backend_types.Step, taskUUID string) (io.ReadCloser, error) {
	ss, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		return nil, err
	}
	if ss.output == nil {
		return nil, ErrStepReaderNotFound
	}
	return ss.output, nil
}

// WaitStep blocks until the step's command exits and returns its final state.
func (e *tenki) WaitStep(ctx context.Context, step *backend_types.Step, taskUUID string) (*backend_types.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	ss, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		return nil, err
	}

	type waitResult struct {
		res *sandbox.Result
		err error
	}
	done := make(chan waitResult, 1)
	go func() {
		res, err := ss.handle.Wait()
		done <- waitResult{res: res, err: err}
	}()

	select {
	case <-ctx.Done():
		// On cancellation, signal the process and let DestroyStep/DestroyWorkflow
		// do the actual teardown. The goroutine above still delivers into the
		// buffered channel, so it does not leak.
		_ = ss.handle.Kill()
		return nil, ctx.Err()
	case r := <-done:
		if r.err != nil {
			return nil, r.err
		}
		return &backend_types.State{
			Exited:   true,
			ExitCode: int(r.res.ExitCode),
		}, nil
	}
}

// DestroyStep stops the step's command. The sandbox itself stays alive so that
// subsequent steps of the same workflow keep sharing the workspace.
func (e *tenki) DestroyStep(_ context.Context, step *backend_types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("stop step %s", step.Name)

	ss, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		if errors.Is(err, ErrStepStateNotFound) || errors.Is(err, ErrWorkflowStateNotFound) {
			return nil
		}
		return err
	}

	if ss.handle != nil {
		_ = ss.handle.Kill()
	}
	if ss.output != nil {
		_ = ss.output.Close()
	}

	if ws, err := e.getWorkflowState(taskUUID); err == nil {
		ws.stepState.Delete(step.UUID)
	}
	return nil
}

// DestroyWorkflow tears down the workflow's sandbox and any remaining steps.
func (e *tenki) DestroyWorkflow(ctx context.Context, _ *backend_types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("delete workflow environment")

	ws, err := e.getWorkflowState(taskUUID)
	if err != nil {
		if errors.Is(err, ErrWorkflowStateNotFound) {
			return nil
		}
		return err
	}

	// stop any steps still running (detached steps, or on cancellation)
	ws.stepState.Range(func(_, value any) bool {
		if ss, ok := value.(*stepState); ok && ss != nil {
			if ss.handle != nil {
				_ = ss.handle.Kill()
			}
			if ss.output != nil {
				_ = ss.output.Close()
			}
		}
		return true
	})

	if ws.session != nil {
		if err := ws.session.CloseIfOpen(ctx); err != nil {
			log.Error().Err(err).Str("taskUUID", taskUUID).Msg("could not close tenki sandbox")
		}
	}

	e.workflows.Delete(taskUUID)
	return nil
}

func (e *tenki) getWorkflowState(taskUUID string) (*workflowState, error) {
	v, ok := e.workflows.Load(taskUUID)
	if !ok {
		return nil, ErrWorkflowStateNotFound
	}
	ws, ok := v.(*workflowState)
	if !ok || ws == nil {
		return nil, fmt.Errorf("could not parse workflow state: %v", v)
	}
	return ws, nil
}

func (e *tenki) getStepState(taskUUID, stepUUID string) (*stepState, error) {
	ws, err := e.getWorkflowState(taskUUID)
	if err != nil {
		return nil, err
	}
	v, ok := ws.stepState.Load(stepUUID)
	if !ok {
		return nil, ErrStepStateNotFound
	}
	ss, ok := v.(*stepState)
	if !ok || ss == nil {
		return nil, fmt.Errorf("could not parse step state: %v", v)
	}
	return ss, nil
}

// sandboxName derives a stable, identifiable sandbox name from the workflow's
// task UUID, mirroring how the docker/kubernetes backends name their resources.
func sandboxName(taskUUID string) string {
	return "woodpecker-" + taskUUID
}

// workflowMetadata builds the metadata attached to the sandbox so it can be
// traced back to its workflow in the Tenki dashboard/CLI. It always tags the
// sandbox as Woodpecker-managed and records the task UUID, and folds in the
// workflow labels the compiler set (repo, pipeline, etc.).
func workflowMetadata(conf *backend_types.Config, taskUUID string) map[string]string {
	md := map[string]string{}

	// Workflow labels are the same across every step; copy them from the first
	// step that has any.
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			if len(step.WorkflowLabels) == 0 {
				continue
			}
			for k, v := range step.WorkflowLabels {
				md[k] = v
			}
			break
		}
		if len(md) > 0 {
			break
		}
	}

	// Our own tracing keys are authoritative and must not be shadowed by a
	// user-defined workflow label, so they are written last.
	md["managed-by"] = "woodpecker"
	md["task-uuid"] = taskUUID
	return md
}

// ensureWorkspaceDir creates the given directory inside the sandbox. Woodpecker
// uses an absolute workspace path (default /woodpecker/src) that lives at the
// filesystem root, which the non-root guest user cannot create. We first try a
// plain mkdir and, only if that fails, escalate with the base image's
// passwordless sudo and hand ownership back to the guest user. Keeping the path
// unchanged preserves Woodpecker's workspace semantics (CI_WORKSPACE, volumes,
// and any user command referencing the path all stay valid).
func ensureWorkspaceDir(ctx context.Context, sess *sandbox.Session, dir string) error {
	// Pass the directory as a positional argument ($1) instead of interpolating
	// it into the script, so the guest shell never re-parses the path (avoids
	// command injection through the workspace directory).
	const script = `set -e
if ! mkdir -p "$1" 2>/dev/null; then
  sudo mkdir -p "$1"
  sudo chown -R "$(id -u):$(id -g)" "$1"
fi`

	res, err := sess.Command([]string{"/bin/sh", "-c", script, "sh", dir}).Exec(ctx)
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("mkdir exited %d: %s", res.ExitCode, res.StderrString())
	}
	return nil
}

// mergeOutput combines the command's stdout and stderr into a single stream,
// as Woodpecker expects one log reader per step. Parallel writes to an
// io.Pipe are gated sequentially, so concurrent copies are safe.
func mergeOutput(h *sandbox.RunHandle) io.ReadCloser {
	pr, pw := io.Pipe()

	var streams []io.Reader
	for _, s := range []io.Reader{h.Stdout, h.Stderr} {
		if s != nil {
			streams = append(streams, s)
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(streams))
	for _, s := range streams {
		go func(r io.Reader) {
			defer wg.Done()
			_, _ = io.Copy(pw, r)
		}(s)
	}
	go func() {
		wg.Wait()
		_ = pw.Close()
	}()

	return pr
}

package nomad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/nomad/api"
	nomadApi "github.com/hashicorp/nomad/api"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

// TODO: clone step container failed but woodpecker thought it succeeded?
// TODO: when a pipeline is cancelled it cannot view the logs it gets "execution error context canceled" in the ui
const (
	bytesToLines     int64 = 120
	defaultTailLines int64 = 10
)

type nomadBackend struct {
	ctx          context.Context
	client       nomadApi.Client
	consulClient consulApi.Client
	vaultCilent  vaultApi.Client
	config       *config
}

type config struct {
	nomadNamespace   string   // Namespace that the agent may launch jobs in
	nomadDatacenters []string // Datacenters eligible for task placement
}

func configFromCliContext(ctx context.Context) (*config, error) {
	if ctx != nil {
		if c, ok := ctx.Value(types.CliContext).(*cli.Context); ok {
			config := config{
				nomadNamespace:   c.String("backend-nomad-namespace"),
				nomadDatacenters: strings.Split(c.String("backend-nomad-datacenters"), ","),
			}
			return &config, nil
		}
	}

	return nil, types.ErrNoCliContextFound
}

// New returns a new Engine.
func New(ctx context.Context) types.Engine {
	return &nomadBackend{
		ctx: ctx,
	}
}

func (nb *nomadBackend) Name() string {
	return "nomad"
}

// TODO: Figure out how to properly log
// TODO: Figure out how often this is called and maybe do something different
func (nb *nomadBackend) IsAvailable(ctx context.Context) bool {
	_, err := nb.client.Status().Leader()
	if err != nil {
		fmt.Println("failed to find nomad leader")
		return false
	}
	return true
}

// Load sets up the Nomad Client. DefaultConfig should handle checking most
// of the NOMAD_* env for overrides
func (nb *nomadBackend) Load(ctx context.Context) error {
	defaultConfig := nomadApi.DefaultConfig()
	client, err := nomadApi.NewClient(defaultConfig)
	if err != nil {
		return fmt.Errorf("error setting up nomad client: %s", err)
	}
	nb.client = *client
	return nil
}

func (nb *nomadBackend) SetupWorkflow(ctx context.Context, conf *types.Config, _ string) error {
	for _, stage := range conf.Stages {
		fmt.Println("stage", stage)
	}
	return nil
}

func (nb *nomadBackend) StartStep(ctx context.Context, step *types.Step, _ string) error {
	// region, err := nb.client.Agent().Region()
	name := fmt.Sprintf("%s", step.Name)
	fmt.Println("name is:", name)
	// cpu := int(step.CPUQuota)
	// shares := int(step.CPUShares)
	// mem := int(step.MemLimit)
	wo := nomadApi.WriteOptions{}

	nj := generateNomadJob(ctx, step)

	jrr, _, err := nb.client.Jobs().Register(&nj, &wo)
	if err != nil {
		return fmt.Errorf("failed to start nomad job: %s", err)
	}

	ctx = context.WithValue(ctx, "evalId", jrr.EvalID)

	return nil
}

// TODO: Should stderr and stdout be merged? Should the lines be prefixed somehow?
func (nb *nomadBackend) TailStep(ctx context.Context, step *types.Step, _ string) (io.ReadCloser, error) {
	qo := nomadApi.QueryOptions{}
	fmt.Println("Getting evals")
	fmt.Println("Getting allocs")

	var evals []*nomadApi.Evaluation
	for {
		var err error
		evals, _, err = nb.client.Jobs().Evaluations(step.Name, &qo)
		fmt.Println("FIRST EVALS", evals)
		if err != nil {
			return nil, fmt.Errorf("failed getting evals from jobs: %s", err)
		}
		if len(evals) == 0 {
			fmt.Println("will retry evals")
			time.Sleep(time.Second * 1)
		} else if len(evals) > 0 {
			fmt.Println("EVAL LEN", len(evals))
			fmt.Println("EVALS", evals)
			bsevals, _ := json.Marshal(evals)
			fmt.Println("BSEVALS", string(bsevals))
			break
		}
	}

	eid := evals[0].ID
	fmt.Println("gettingt eallocs")
	var eallocs []*nomadApi.AllocationListStub
	for {
		var err error
		eallocs, _, err = nb.client.Evaluations().Allocations(eid, &qo)
		if err != nil {
			return nil, fmt.Errorf("failed getting allocs from eval: %s", err)
		}

		if len(eallocs) == 0 {
			fmt.Println("retrying")
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}

	// For now assuming there will only be a single alloc per eval
	fmt.Println("gettingt allocId")
	allocId := eallocs[0].ID
	fmt.Println("getting alloc info")
	// For now assuming there will only be a single alloc per eval
	alloc, _, err := nb.client.Allocations().Info(allocId, &qo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve alloc info: %s", err)
	}
	fmt.Println("finished getting alloc info")
	// var offset int64 = defaultTailLines * bytesToLines
	var offset int64 = 0

	// maybe replace by...
	// Launching a go routine that checks for the status to be running or failed?
	// block with <- until its running
	for {
		fmt.Println("checking for task to be running")
		state, err := nb.taskState(alloc.ID)
		if err != nil {
			return nil, fmt.Errorf("failed checking task status: %s", err)
		}
		if state == "failed" {
			fmt.Println("task is failed checking for state in earlier for block?")
			break
		} else if state == "dead" {
			fmt.Println("Task is already done (dead?)")
			break
		} else if state == "running" {
			fmt.Println("its running")
			break
		} else if state == "pending" {
			fmt.Println("still pending")
			time.Sleep(time.Second * 2)
			continue
		} else {
			fmt.Println("THIS WAS THE STATE BEFORE PANICKING:", state)
			panic("freak out! this shouldn't have happened!")
		}
	}

	r, err := nb.followLogs(ctx, alloc, step.Name, "stdout", nomadApi.OriginStart, offset)
	if err != nil {
		fmt.Println("failed to follow logs println")
		return nil, fmt.Errorf("failed to followLogs stdout: %s", err)
	}
	re, err := nb.followLogs(ctx, alloc, step.Name, "stderr", nomadApi.OriginStart, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to followLogs stderr: %s", err)
	}

	rc, wc := io.Pipe()

	go func() {
		_, err = io.Copy(wc, re)
		if err != nil {
			return
		}
	}()

	go func() {
		defer rc.Close()
		defer wc.Close()
		_, err = io.Copy(wc, r)
		if err != nil {
			return
		}
	}()

	return rc, nil
}

func (nb nomadBackend) followLogs(ctx context.Context, alloc *nomadApi.Allocation, task, logType, origin string, offset int64) (io.ReadCloser, error) {
	cancel := make(chan struct{})
	frames, errCh := nb.client.AllocFS().Logs(alloc, true, task, logType, origin, offset, cancel, nil)
	select {
	case err := <-errCh:
		fmt.Println("errCh hit?")
		return nil, err
	default:
		fmt.Println("default for first select")
	}

	taskFinishedCh := make(chan bool)

	go func() {
		for {
			state, err := nb.taskState(alloc.ID)
			if err != nil {
				fmt.Println("error checking task status")
				return
			}
			switch state {
			case "dead":
				fmt.Println("task status is completed! (dead?)")
				taskFinishedCh <- true
				return
			case "failed":
				fmt.Println("It failed")
				taskFinishedCh <- true
				return
			case "running":
				fmt.Println("in progress...")
				fmt.Println("sending false")
				taskFinishedCh <- false
				time.Sleep(time.Second * 1)
			}

			select {
			default:
				fmt.Println("default in ch for checking task")
				time.Sleep(time.Second * 1)
			case <-ctx.Done():
				fmt.Println("hit done in ch for checking task?")
				return
			}

		}
	}()

	// Create a reader
	var r io.ReadCloser
	fmt.Println("new frame reader")
	frameReader := api.NewFrameReader(frames, errCh, cancel)
	frameReader.SetUnblockTime(500 * time.Millisecond)
	fmt.Println("about to launch go func")
	r = frameReader
	go func() {
		defer r.Close()

		for {
			fmt.Println("entering for loop in followLogs")
			select {
			default:
				fmt.Println("default case in follow logs")
				time.Sleep(time.Second * 2)
			case <-ctx.Done():
				fmt.Println("in ctx.Done() in follow logs")
				return
			case status := <-taskFinishedCh:
				fmt.Println("task status is!!:", taskFinishedCh)
				if status {
					fmt.Println("status was true so returning?:", status)
					return
				}
			}

		}
	}()

	fmt.Println("returning from follow logs")

	return r, nil
}

// from https://github.com/hashicorp/nomad/blob/72ad885a47205470ac94333640200a9cf9303df5/command/alloc_logs.go#L273
func followFile(client *nomadApi.Client, alloc *nomadApi.Allocation,
	follow bool, task, logType, origin string, offset int64,
) (io.ReadCloser, error) {
	cancel := make(chan struct{})
	frames, errCh := client.AllocFS().Logs(alloc, follow, task, logType, origin, offset, cancel, nil)
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Create a reader
	var r io.ReadCloser
	frameReader := api.NewFrameReader(frames, errCh, cancel)
	frameReader.SetUnblockTime(500 * time.Millisecond)
	r = frameReader

	go func() {
		<-signalCh

		// End the streaming
		r.Close()
	}()

	return r, nil
}

func (nb *nomadBackend) WaitStep(ctx context.Context, step *types.Step, _ string) (*types.State, error) {
	// launch go routine with ticker
	// timer? How long can woodpecker go?
	// cancel with channel
	name := step.Name
	bs := &types.State{}
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	timer := time.NewTimer(time.Minute * 1)
	defer timer.Stop()

	for {
		// Get job info
		qo := nomadApi.QueryOptions{}
		info, _, err := nb.client.Jobs().Info(name, &qo)
		if err != nil {
			return nil, fmt.Errorf("failed to get job info: %s", err)
		}

		// Check current job status and check if its complete, lost or failed
		// NOTE: JOB STATUS OF "dead" is probably ok. Need to check if alloc was complete/failed/etc
		switch *info.Status {
		case "running", "pending":
			// Leave switch and check for cancellation/timeout before repeating
			break
		case "failed", "lost":
			bs.Exited = true
			bs.ExitCode = 1
			return bs, nil
		case "completed", "dead":
			bs.Exited = true
			bs.ExitCode = 0
			return bs, nil
		default:
			panic("unexpected case")
		}

		// TODO: if it is cancelled or timed out we should make sure the batch job is cleaned up
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case <-timer.C:
			return nil, fmt.Errorf("job wait timeout")
		case <-ticker.C:
			continue
		}
	}
}

func (nb *nomadBackend) taskState(allocId string) (string, error) {
	qo := nomadApi.QueryOptions{}
	info, _, err := nb.client.Allocations().Info(allocId, &qo)
	if err != nil {
		return "", fmt.Errorf("failed to get alloc info: %s", err)
	}
	states := info.TaskStates
	if len(states) == 0 {
		return "pending", nil
	}
	for _, v := range states {
		switch v.State {
		case "pending":
			fmt.Println("alloc is pending")
			return "pending", nil
		case "running":
			fmt.Println("The alloc is running")
			return "running", nil
		case "dead":
			fmt.Println("alloc is dead (which should be good?)")
			return "dead", nil
		case "failed":
			fmt.Println("alloc is failed")
			return "failed", fmt.Errorf("alloc failed")
		default:
			fmt.Println("THE STATE IS:", v.State)
			panic("unexpected state")
		}
	}
	return "unknonwn", nil
}

func (nb *nomadBackend) DestroyWorkflow(context.Context, *types.Config, string) error {
	fmt.Println("this is where we'd tear stuff down")
	return nil
}

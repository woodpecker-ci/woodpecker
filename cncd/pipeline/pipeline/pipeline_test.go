// +build integration

package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend/docker"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend/kubernetes"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/interrupt"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/multipart"
)

func Test_Docker_01_single_step(t *testing.T) {
	run(t, "fixtures/01_single_step/pipeline.json", true)
}

func Test_Kubernetes_01_single_step(t *testing.T) {
	run(t, "fixtures/01_single_step/pipeline.json", false)
}

func Test_Docker_02_services(t *testing.T) {
	run(t, "fixtures/02_services/pipeline.json", true)
}

func Test_Kubernetes_02_services(t *testing.T) {
	run(t, "fixtures/02_services/pipeline.json", false)
}

func run(t *testing.T, fixture string, dockerEngine bool) {
	reader, err := os.Open(fixture)
	if err != nil {
		t.Errorf("Could not read pipeline %f", err)
	}
	defer reader.Close()

	config, err := Parse(reader)
	if err != nil {
		t.Errorf("Could not parse pipeline %f", err)
	}

	var defaultTracer = TraceFunc(func(state *State) error {
		if state.Process.Exited {
			if state.Process.ExitCode != 0 {
				t.Errorf("proc %q exited with status %d\n", state.Pipeline.Step.Name, state.Process.ExitCode)
			} else {
				fmt.Printf("proc %q exited with status %d\n", state.Pipeline.Step.Name, state.Process.ExitCode)
			}
		} else {
			fmt.Printf("proc %q started\n", state.Pipeline.Step.Name)
			state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "success"
			state.Pipeline.Step.Environment["CI_BUILD_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)
			if state.Pipeline.Error != nil {
				state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "failure"
			}
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	ctx = interrupt.WithContext(ctx)

	if dockerEngine {
		dockerEngine, err := docker.NewEnv()
		if err != nil {
			t.Errorf("Could not start Docker engine %f", err)
		}
		err = New(config,
			WithContext(ctx),
			WithLogger(defaultLogger),
			WithTracer(defaultTracer),
			WithEngine(dockerEngine),
		).Run()

		if err != nil {
			t.Errorf("Pipeline exited with error %v", err)
		}
	} else {
		// os.Setenv("KUBECONFIG", "/etc/rancher/k3s/k3s.yaml")
		kubernetesEngine, err := kubernetes.New("default", "example-nfs", "100Mi")
		if err != nil {
			t.Errorf("Could not start Kubernetes engine %f", err)
		}
		err = New(config,
			WithContext(ctx),
			WithLogger(defaultLogger),
			WithTracer(defaultTracer),
			WithEngine(kubernetesEngine),
		).Run()

		if err != nil {
			t.Errorf("Pipeline exited with error %v", err)
		}
	}
}

var defaultLogger = LogFunc(func(proc *backend.Step, rc multipart.Reader) error {
	part, err := rc.NextPart()
	if err != nil {
		return err
	}
	io.Copy(os.Stderr, part)
	return nil
})

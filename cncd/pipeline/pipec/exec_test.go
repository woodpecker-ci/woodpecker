package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend/docker"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend/kubernetes"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/interrupt"
)

func TestExec(t *testing.T) {
	reader, err := os.Open("../samples/sample_1/pipeline.json")
	if err != nil {
		t.Errorf("Could not read pipeline %f", err)
	}
	defer reader.Close()

	config, err := pipeline.Parse(reader)
	if err != nil {
		t.Errorf("Could not parse pipeline %f", err)
	}

	engine, err := docker.NewEnv()
	if err != nil {
		t.Errorf("Could not start Docker engine %f", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	ctx = interrupt.WithContext(ctx)

	err = pipeline.New(config,
		pipeline.WithContext(ctx),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithTracer(defaultTracer),
		pipeline.WithEngine(engine),
	).Run()

	if err != nil {
		t.Errorf("Pipeline exited with error %v", err)
	}
}

func TestKubeExec(t *testing.T) {
	reader, err := os.Open("../samples/sample_1/pipeline.json")
	if err != nil {
		t.Errorf("Could not read pipeline %f", err)
	}
	defer reader.Close()

	config, err := pipeline.Parse(reader)
	if err != nil {
		t.Errorf("Could not parse pipeline %f", err)
	}

	// os.Setenv("KUBECONFIG", "/etc/rancher/k3s/k3s.yaml")
	engine, err := kubernetes.New()
	if err != nil {
		t.Errorf("Could not start Kubernetes engine %f", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	ctx = interrupt.WithContext(ctx)

	err = pipeline.New(config,
		pipeline.WithContext(ctx),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithTracer(defaultTracer),
		pipeline.WithEngine(engine),
	).Run()

	if err != nil {
		t.Errorf("Pipeline exited with error %v", err)
	}
}

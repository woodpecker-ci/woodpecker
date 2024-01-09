// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"

	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	// To authenticate to GCP K8s clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	EngineName = "kubernetes"
)

var defaultDeleteOptions = newDefaultDeleteOptions()

type kube struct {
	ctx    context.Context
	client kubernetes.Interface
	config *config
	goos   string
}

type config struct {
	Namespace            string
	StorageClass         string
	VolumeSize           string
	StorageRwx           bool
	PodLabels            map[string]string
	PodAnnotations       map[string]string
	ImagePullSecretNames []string
	SecurityContext      SecurityContextConfig
}
type SecurityContextConfig struct {
	RunAsNonRoot bool
}

func newDefaultDeleteOptions() metav1.DeleteOptions {
	gracePeriodSeconds := int64(0) // immediately
	propagationPolicy := metav1.DeletePropagationBackground

	return metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &propagationPolicy,
	}
}

func configFromCliContext(ctx context.Context) (*config, error) {
	if ctx != nil {
		if c, ok := ctx.Value(types.CliContext).(*cli.Context); ok {
			config := config{
				Namespace:            c.String("backend-k8s-namespace"),
				StorageClass:         c.String("backend-k8s-storage-class"),
				VolumeSize:           c.String("backend-k8s-volume-size"),
				StorageRwx:           c.Bool("backend-k8s-storage-rwx"),
				PodLabels:            make(map[string]string), // just init empty map to prevent nil panic
				PodAnnotations:       make(map[string]string), // just init empty map to prevent nil panic
				ImagePullSecretNames: c.StringSlice("backend-k8s-pod-image-pull-secret-names"),
				SecurityContext: SecurityContextConfig{
					RunAsNonRoot: c.Bool("backend-k8s-secctx-nonroot"),
				},
			}
			// TODO: remove in next major
			if len(config.ImagePullSecretNames) == 1 && config.ImagePullSecretNames[0] == "regcred" {
				log.Warn().Msg("WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES is set to the default ('regcred'). It will default to empty in Woodpecker 3.0. Set it explicitly before then.")
			}
			// Unmarshal label and annotation settings here to ensure they're valid on startup
			if labels := c.String("backend-k8s-pod-labels"); labels != "" {
				if err := yaml.Unmarshal([]byte(labels), &config.PodLabels); err != nil {
					log.Error().Msgf("could not unmarshal pod labels '%s': %s", c.String("backend-k8s-pod-labels"), err)
					return nil, err
				}
			}
			if annotations := c.String("backend-k8s-pod-annotations"); annotations != "" {
				if err := yaml.Unmarshal([]byte(c.String("backend-k8s-pod-annotations")), &config.PodAnnotations); err != nil {
					log.Error().Msgf("could not unmarshal pod annotations '%s': %s", c.String("backend-k8s-pod-annotations"), err)
					return nil, err
				}
			}
			return &config, nil
		}
	}

	return nil, types.ErrNoCliContextFound
}

// New returns a new Kubernetes Backend.
func New(ctx context.Context) types.Backend {
	return &kube{
		ctx: ctx,
	}
}

func (e *kube) Name() string {
	return EngineName
}

func (e *kube) IsAvailable(context.Context) bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	return len(host) > 0
}

func (e *kube) Load(context.Context) (*types.BackendInfo, error) {
	config, err := configFromCliContext(e.ctx)
	if err != nil {
		return nil, err
	}
	e.config = config

	var kubeClient kubernetes.Interface
	_, err = rest.InClusterConfig()
	if err != nil {
		kubeClient, err = getClientOutOfCluster()
	} else {
		kubeClient, err = getClientInsideOfCluster()
	}

	if err != nil {
		return nil, err
	}

	e.client = kubeClient

	// TODO(2693): use info resp of kubeClient to define platform var
	e.goos = runtime.GOOS
	return &types.BackendInfo{
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}, nil
}

// Setup the pipeline environment.
func (e *kube) SetupWorkflow(ctx context.Context, conf *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("Setting up Kubernetes primitives")

	for _, vol := range conf.Volumes {
		_, err := startVolume(ctx, e, vol.Name)
		if err != nil {
			return err
		}
	}

	extraHosts := []types.HostAlias{}
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			if step.Type == types.StepTypeService {
				svc, err := startService(ctx, e, step)
				if err != nil {
					return err
				}
				hostAlias := types.HostAlias{Name: step.Networks[0].Aliases[0], IP: svc.Spec.ClusterIP}
				extraHosts = append(extraHosts, hostAlias)
			}
		}
	}
	log.Trace().Msgf("Adding extra hosts: %v", extraHosts)
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			step.ExtraHosts = extraHosts
		}
	}

	return nil
}

// Start the pipeline step.
func (e *kube) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	if step.Type == types.StepTypeService {
		// a service should be started by SetupWorkflow so we can ignore it
		log.Trace().Msgf("StartStep got service '%s', ignoring it.", step.Name)
		return nil
	}
	log.Trace().Str("taskUUID", taskUUID).Msgf("Starting step: %s", step.Name)
	_, err := startPod(ctx, e, step)
	return err
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *kube) WaitStep(ctx context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("Waiting for pod: %s", podName)

	finished := make(chan bool)

	podUpdated := func(old, new any) {
		pod := new.(*v1.Pod)
		if pod.Name == podName {
			if isImagePullBackOffState(pod) {
				finished <- true
			}

			switch pod.Status.Phase {
			case v1.PodSucceeded, v1.PodFailed, v1.PodUnknown:
				finished <- true
			}
		}
	}

	// TODO 5 seconds is against best practice, k3s didn't work otherwise
	si := informers.NewSharedInformerFactoryWithOptions(e.client, 5*time.Second, informers.WithNamespace(e.config.Namespace))
	if _, err := si.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: podUpdated,
		},
	); err != nil {
		return nil, err
	}

	stop := make(chan struct{})
	si.Start(stop)
	defer close(stop)

	// TODO Cancel on ctx.Done
	<-finished

	pod, err := e.client.CoreV1().Pods(e.config.Namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if isImagePullBackOffState(pod) {
		return nil, fmt.Errorf("Could not pull image for pod %s", pod.Name)
	}

	bs := &types.State{
		ExitCode:  int(pod.Status.ContainerStatuses[0].State.Terminated.ExitCode),
		Exited:    true,
		OOMKilled: false,
	}

	return bs, nil
}

// Tail the pipeline step logs.
func (e *kube) TailStep(ctx context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("Tail logs of pod: %s", podName)

	up := make(chan bool)

	podUpdated := func(old, new any) {
		pod := new.(*v1.Pod)
		if pod.Name == podName {
			switch pod.Status.Phase {
			case v1.PodRunning, v1.PodSucceeded, v1.PodFailed:
				up <- true
			}
		}
	}

	// TODO 5 seconds is against best practice, k3s didn't work otherwise
	si := informers.NewSharedInformerFactoryWithOptions(e.client, 5*time.Second, informers.WithNamespace(e.config.Namespace))
	if _, err := si.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: podUpdated,
		},
	); err != nil {
		return nil, err
	}

	stop := make(chan struct{})
	si.Start(stop)
	defer close(stop)

	<-up

	opts := &v1.PodLogOptions{
		Follow:    true,
		Container: podName,
	}

	logs, err := e.client.CoreV1().RESTClient().Get().
		Namespace(e.config.Namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		VersionedParams(opts, scheme.ParameterCodec).
		Stream(ctx)
	if err != nil {
		return nil, err
	}
	rc, wc := io.Pipe()

	go func() {
		defer logs.Close()
		defer wc.Close()
		defer rc.Close()

		_, err = io.Copy(wc, logs)
		if err != nil {
			return
		}
	}()
	return rc, nil
}

func (e *kube) DestroyStep(_ context.Context, step *types.Step, taskUUID string) error {
	if step.Type == types.StepTypeService {
		// a service should be stopped by DestroyWorkflow so we can ignore it
		log.Trace().Msgf("DestroyStep got service '%s', ignoring it.", step.Name)
		return nil
	}
	log.Trace().Str("taskUUID", taskUUID).Msgf("Stopping step: %s", step.Name)
	err := stopPod(e.ctx, e, step, defaultDeleteOptions)
	return err
}

// Destroy the pipeline environment.
func (e *kube) DestroyWorkflow(_ context.Context, conf *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("Deleting Kubernetes primitives")

	// Use noContext because the ctx sent to this function will be canceled/done in case of error or canceled by user.
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			err := stopPod(e.ctx, e, step, defaultDeleteOptions)
			if err != nil {
				return err
			}

			if step.Type == types.StepTypeService {
				err := stopService(e.ctx, e, step, defaultDeleteOptions)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, vol := range conf.Volumes {
		err := stopVolume(e.ctx, e, vol.Name, defaultDeleteOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

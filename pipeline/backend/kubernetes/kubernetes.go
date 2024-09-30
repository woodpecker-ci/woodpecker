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
	std_errs "errors"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // To authenticate to GCP K8s clusters
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

const (
	EngineName = "kubernetes"
	// TODO: 5 seconds is against best practice, k3s didn't work otherwise
	defaultResyncDuration = 5 * time.Second
)

var defaultDeleteOptions = newDefaultDeleteOptions()

type kube struct {
	client kubernetes.Interface
	config *config
	goos   string
}

type config struct {
	Namespace                   string
	StorageClass                string
	VolumeSize                  string
	StorageRwx                  bool
	PodLabels                   map[string]string
	PodLabelsAllowFromStep      bool
	PodAnnotations              map[string]string
	PodAnnotationsAllowFromStep bool
	PodNodeSelector             map[string]string
	ImagePullSecretNames        []string
	SecurityContext             SecurityContextConfig
	NativeSecretsAllowFromStep  bool
}
type SecurityContextConfig struct {
	RunAsNonRoot bool
}

func newDefaultDeleteOptions() meta_v1.DeleteOptions {
	gracePeriodSeconds := int64(0) // immediately
	propagationPolicy := meta_v1.DeletePropagationBackground

	return meta_v1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &propagationPolicy,
	}
}

func configFromCliContext(ctx context.Context) (*config, error) {
	if ctx != nil {
		if c, ok := ctx.Value(types.CliCommand).(*cli.Command); ok {
			config := config{
				Namespace:                   c.String("backend-k8s-namespace"),
				StorageClass:                c.String("backend-k8s-storage-class"),
				VolumeSize:                  c.String("backend-k8s-volume-size"),
				StorageRwx:                  c.Bool("backend-k8s-storage-rwx"),
				PodLabels:                   make(map[string]string), // just init empty map to prevent nil panic
				PodLabelsAllowFromStep:      c.Bool("backend-k8s-pod-labels-allow-from-step"),
				PodAnnotations:              make(map[string]string), // just init empty map to prevent nil panic
				PodAnnotationsAllowFromStep: c.Bool("backend-k8s-pod-annotations-allow-from-step"),
				PodNodeSelector:             make(map[string]string), // just init empty map to prevent nil panic
				ImagePullSecretNames:        c.StringSlice("backend-k8s-pod-image-pull-secret-names"),
				SecurityContext: SecurityContextConfig{
					RunAsNonRoot: c.Bool("backend-k8s-secctx-nonroot"), // cspell:words secctx nonroot
				},
				NativeSecretsAllowFromStep: c.Bool("backend-k8s-allow-native-secrets"),
			}
			// Unmarshal label and annotation settings here to ensure they're valid on startup
			if labels := c.String("backend-k8s-pod-labels"); labels != "" {
				if err := yaml.Unmarshal([]byte(labels), &config.PodLabels); err != nil {
					log.Error().Err(err).Msgf("could not unmarshal pod labels '%s'", c.String("backend-k8s-pod-labels"))
					return nil, err
				}
			}
			if annotations := c.String("backend-k8s-pod-annotations"); annotations != "" {
				if err := yaml.Unmarshal([]byte(c.String("backend-k8s-pod-annotations")), &config.PodAnnotations); err != nil {
					log.Error().Err(err).Msgf("could not unmarshal pod annotations '%s'", c.String("backend-k8s-pod-annotations"))
					return nil, err
				}
			}
			if nodeSelector := c.String("backend-k8s-pod-node-selector"); nodeSelector != "" {
				if err := yaml.Unmarshal([]byte(nodeSelector), &config.PodNodeSelector); err != nil {
					log.Error().Err(err).Msgf("could not unmarshal pod node selector '%s'", nodeSelector)
					return nil, err
				}
			}
			return &config, nil
		}
	}

	return nil, types.ErrNoCliContextFound
}

// New returns a new Kubernetes Backend.
func New() types.Backend {
	return &kube{}
}

func (e *kube) Name() string {
	return EngineName
}

func (e *kube) IsAvailable(context.Context) bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	return len(host) > 0
}

func (e *kube) Flags() []cli.Flag {
	return Flags
}

func (e *kube) Load(ctx context.Context) (*types.BackendInfo, error) {
	config, err := configFromCliContext(ctx)
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

func (e *kube) getConfig() *config {
	if e.config == nil {
		return nil
	}
	c := *e.config
	c.PodLabels = maps.Clone(e.config.PodLabels)
	c.PodAnnotations = maps.Clone(e.config.PodAnnotations)
	c.PodNodeSelector = maps.Clone(e.config.PodNodeSelector)
	c.ImagePullSecretNames = slices.Clone(e.config.ImagePullSecretNames)
	return &c
}

// SetupWorkflow sets up the pipeline environment.
func (e *kube) SetupWorkflow(ctx context.Context, conf *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("Setting up Kubernetes primitives")

	for _, vol := range conf.Volumes {
		_, err := startVolume(ctx, e, vol.Name)
		if err != nil {
			return err
		}
	}

	var extraHosts []types.HostAlias
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
	log.Trace().Msgf("adding extra hosts: %v", extraHosts)
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			step.ExtraHosts = extraHosts
		}
	}

	return nil
}

// StartStep starts the pipeline step.
func (e *kube) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	options, err := parseBackendOptions(step)
	if err != nil {
		log.Error().Err(err).Msg("could not parse backend options")
	}

	if needsRegistrySecret(step) {
		err = startRegistrySecret(ctx, e, step)
		if err != nil {
			return err
		}
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("starting step: %s", step.Name)
	_, err = startPod(ctx, e, step, options)
	return err
}

// WaitStep waits for the pipeline step to complete and returns
// the completion results.
func (e *kube) WaitStep(ctx context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("waiting for pod: %s", podName)

	finished := make(chan bool)

	podUpdated := func(_, new any) {
		pod, ok := new.(*v1.Pod)
		if !ok {
			log.Error().Msgf("could not parse pod: %v", new)
			return
		}

		if pod.Name == podName {
			if isImagePullBackOffState(pod) || isInvalidImageName(pod) {
				finished <- true
			}

			switch pod.Status.Phase {
			case v1.PodSucceeded, v1.PodFailed, v1.PodUnknown:
				finished <- true
			}
		}
	}

	si := informers.NewSharedInformerFactoryWithOptions(e.client, defaultResyncDuration, informers.WithNamespace(e.config.Namespace))
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

	// TODO: Cancel on ctx.Done
	<-finished

	pod, err := e.client.CoreV1().Pods(e.config.Namespace).Get(ctx, podName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if isImagePullBackOffState(pod) || isInvalidImageName(pod) {
		return nil, fmt.Errorf("could not pull image for pod %s", podName)
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return nil, fmt.Errorf("no container statuses found for pod %s", podName)
	}

	cs := pod.Status.ContainerStatuses[0]

	if cs.State.Terminated == nil {
		err := fmt.Errorf("no terminated state found for container %s/%s", podName, cs.Name)
		log.Error().Str("taskUUID", taskUUID).Str("pod", podName).Str("container", cs.Name).Interface("state", cs.State).Msg(err.Error())
		return nil, err
	}

	bs := &types.State{
		ExitCode:  int(cs.State.Terminated.ExitCode),
		Exited:    true,
		OOMKilled: false,
	}

	return bs, nil
}

// TailStep tails the pipeline step logs.
func (e *kube) TailStep(ctx context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of pod: %s", podName)

	up := make(chan bool)

	podUpdated := func(_, new any) {
		pod, ok := new.(*v1.Pod)
		if !ok {
			log.Error().Msgf("could not parse pod: %v", new)
			return
		}

		if pod.Name == podName {
			if isImagePullBackOffState(pod) || isInvalidImageName(pod) {
				up <- true
			}
			switch pod.Status.Phase {
			case v1.PodRunning, v1.PodSucceeded, v1.PodFailed:
				up <- true
			}
		}
	}

	si := informers.NewSharedInformerFactoryWithOptions(e.client, defaultResyncDuration, informers.WithNamespace(e.config.Namespace))
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

func (e *kube) DestroyStep(ctx context.Context, step *types.Step, taskUUID string) error {
	var errs []error
	log.Trace().Str("taskUUID", taskUUID).Msgf("Stopping step: %s", step.Name)
	if needsRegistrySecret(step) {
		err := stopRegistrySecret(ctx, e, step, defaultDeleteOptions)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := stopPod(ctx, e, step, defaultDeleteOptions)
	if err != nil {
		errs = append(errs, err)
	}
	return std_errs.Join(errs...)
}

// DestroyWorkflow destroys the pipeline environment.
func (e *kube) DestroyWorkflow(ctx context.Context, conf *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("deleting Kubernetes primitives")

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			err := stopPod(ctx, e, step, defaultDeleteOptions)
			if err != nil {
				return err
			}

			if step.Type == types.StepTypeService {
				err := stopService(ctx, e, step, defaultDeleteOptions)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, vol := range conf.Volumes {
		err := stopVolume(ctx, e, vol.Name, defaultDeleteOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

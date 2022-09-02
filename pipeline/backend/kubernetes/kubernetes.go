package kubernetes

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	// To authenticate to GCP K8s clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var noContext = context.Background()

type kube struct {
	namespace    string
	storageClass string
	volumeSize   string
	storageRwx   bool
	client       kubernetes.Interface
}

// New returns a new Kubernetes Engine.
func New(namespace, storageClass, volumeSize string, storageRwx bool) types.Engine {
	return &kube{
		namespace:    namespace,
		storageClass: storageClass,
		volumeSize:   volumeSize,
		storageRwx:   storageRwx,
	}
}

func (e *kube) Name() string {
	return "kubernetes"
}

func (e *kube) IsAvailable() bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	return len(host) > 0
}

func (e *kube) Load() error {
	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient, err = getClientOutOfCluster()
	} else {
		kubeClient, err = getClientInsideOfCluster()
	}

	if err != nil {
		return err
	}

	e.client = kubeClient

	return nil
}

// Setup the pipeline environment.
func (e *kube) Setup(ctx context.Context, conf *types.Config) error {
	log.Trace().Msgf("Setting up Kubernetes primitives")

	for _, vol := range conf.Volumes {
		pvc := PersistentVolumeClaim(e.namespace, vol.Name, e.storageClass, e.volumeSize, e.storageRwx)
		_, err := e.client.CoreV1().PersistentVolumeClaims(e.namespace).Create(ctx, pvc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	extraHosts := []string{}

	for _, stage := range conf.Stages {
		if stage.Alias == "services" {
			for _, step := range stage.Steps {
				log.Trace().Str("pod-name", podName(step)).Msgf("Creating service: %s", step.Name)
				// TODO: support ports setting
				// svc, err := Service(e.namespace, step.Name, podName(step), step.Ports)
				svc, err := Service(e.namespace, step.Name, podName(step), []string{})
				if err != nil {
					return err
				}

				svc, err = e.client.CoreV1().Services(e.namespace).Create(ctx, svc, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				extraHosts = append(extraHosts, step.Networks[0].Aliases[0]+":"+svc.Spec.ClusterIP)
			}
		}
	}

	log.Trace().Msgf("Adding extra hosts: %s", strings.Join(extraHosts, ", "))
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			step.ExtraHosts = extraHosts
		}
	}

	return nil
}

// Start the pipeline step.
func (e *kube) Exec(ctx context.Context, step *types.Step) error {
	pod := Pod(e.namespace, step)
	log.Trace().Msgf("Creating pod: %s", pod.Name)
	_, err := e.client.CoreV1().Pods(e.namespace).Create(ctx, pod, metav1.CreateOptions{})
	return err
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *kube) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	podName := podName(step)

	finished := make(chan bool)

	podUpdated := func(old, new interface{}) {
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
	si := informers.NewSharedInformerFactoryWithOptions(e.client, 5*time.Second, informers.WithNamespace(e.namespace))
	si.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: podUpdated,
		},
	)
	si.Start(wait.NeverStop)

	// TODO Cancel on ctx.Done
	<-finished

	pod, err := e.client.CoreV1().Pods(e.namespace).Get(ctx, podName, metav1.GetOptions{})
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
func (e *kube) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	podName := podName(step)

	up := make(chan bool)

	podUpdated := func(old, new interface{}) {
		pod := new.(*v1.Pod)
		if pod.Name == podName {
			switch pod.Status.Phase {
			case v1.PodRunning, v1.PodSucceeded, v1.PodFailed:
				up <- true
			}
		}
	}

	// TODO 5 seconds is against best practice, k3s didn't work otherwise
	si := informers.NewSharedInformerFactoryWithOptions(e.client, 5*time.Second, informers.WithNamespace(e.namespace))
	si.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: podUpdated,
		},
	)
	si.Start(wait.NeverStop)

	<-up

	opts := &v1.PodLogOptions{
		Follow: true,
	}

	logs, err := e.client.CoreV1().RESTClient().Get().
		Namespace(e.namespace).
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

	// rc := io.NopCloser(bytes.NewReader(e.logs.Bytes()))
	// e.logs.Reset()
	// return rc, nil
}

// Destroy the pipeline environment.
func (e *kube) Destroy(_ context.Context, conf *types.Config) error {
	gracePeriodSeconds := int64(0) // immediately
	dpb := metav1.DeletePropagationBackground

	deleteOpts := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &dpb,
	}

	// Use noContext because the ctx sent to this function will be canceled/done in case of error or canceled by user.
	// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
	// Trace log them in case the info could be useful when troubleshooting.

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			log.Trace().Msgf("Deleting pod: %s", podName(step))
			if err := e.client.CoreV1().Pods(e.namespace).Delete(noContext, podName(step), deleteOpts); err != nil {
				if errors.IsNotFound(err) {
					log.Trace().Err(err).Msgf("Unable to delete pod %s", podName(step))
				} else {
					return err
				}
			}
		}
	}

	for _, stage := range conf.Stages {
		if stage.Alias == "services" {
			for _, step := range stage.Steps {
				log.Trace().Msgf("Deleting service: %s", step.Name)
				// TODO: support ports setting
				// svc, err := Service(e.namespace, step.Name, step.Alias, step.Ports)
				svc, err := Service(e.namespace, step.Name, step.Alias, []string{})
				if err != nil {
					return err
				}
				if err := e.client.CoreV1().Services(e.namespace).Delete(noContext, svc.Name, deleteOpts); err != nil {
					if errors.IsNotFound(err) {
						log.Trace().Err(err).Msgf("Unable to service pod %s", svc.Name)
					} else {
						return err
					}
				}
			}
		}
	}

	for _, vol := range conf.Volumes {
		pvc := PersistentVolumeClaim(e.namespace, vol.Name, e.storageClass, e.volumeSize, e.storageRwx)
		err := e.client.CoreV1().PersistentVolumeClaims(e.namespace).Delete(noContext, pvc.Name, deleteOpts)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Trace().Err(err).Msgf("Unable to delete pvc %s", pvc.Name)
			} else {
				return err
			}
		}
	}

	return nil
}

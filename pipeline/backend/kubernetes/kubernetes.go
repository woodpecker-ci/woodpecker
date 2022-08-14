package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"

	v1 "k8s.io/api/core/v1"
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

type kube struct {
	logs         *bytes.Buffer // TODO remove
	namespace    string
	storageClass string
	volumeSize   string
	client       kubernetes.Interface
}

// New returns a new Kubernetes Engine.
func New(namespace, storageClass, volumeSize string) types.Engine {
	return &kube{
		logs:         new(bytes.Buffer),
		namespace:    namespace,
		storageClass: storageClass,
		volumeSize:   volumeSize,
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
	e.logs.WriteString("Setting up Kubernetes primitives\n")

	for _, vol := range conf.Volumes {
		pvc := PersistentVolumeClaim(e.namespace, vol.Name, e.storageClass, e.volumeSize)
		_, err := e.client.CoreV1().PersistentVolumeClaims(e.namespace).Create(ctx, pvc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	extraHosts := []string{}

	for _, stage := range conf.Stages {
		if stage.Alias == "services" {
			for _, step := range stage.Steps {
				e.logs.WriteString("Creating service\n")
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

	e.logs.WriteString(strings.Join(extraHosts, ", ") + "\n")
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			step.ExtraHosts = extraHosts
		}
	}

	return nil
}

// Start the pipeline step.
func (e *kube) Exec(ctx context.Context, step *types.Step) error {
	e.logs.WriteString("Creating pod\n")
	e.logs.WriteString(strings.Join(step.ExtraHosts, " ") + "\n")
	pod := Pod(e.namespace, step)
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
		defer e.logs.Reset()
		defer logs.Close()
		defer wc.Close()
		defer rc.Close()

		systemLogs := io.NopCloser(bytes.NewReader(e.logs.Bytes()))
		_, err := io.Copy(wc, systemLogs)
		if err != nil {
			return
		}
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
func (e *kube) Destroy(ctx context.Context, conf *types.Config) error {
	var gracePeriodSeconds int64 = 0 // immediately
	dpb := metav1.DeletePropagationBackground

	deleteOpts := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &dpb,
	}

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			e.logs.WriteString("Deleting pod\n")
			if err := e.client.CoreV1().Pods(e.namespace).Delete(ctx, podName(step), deleteOpts); err != nil {
				return err
			}
		}
	}

	for _, stage := range conf.Stages {
		if stage.Alias == "services" {
			for _, step := range stage.Steps {
				e.logs.WriteString("Deleting service\n")
				// TODO: support ports setting
				// svc, err := Service(e.namespace, step.Name, step.Alias, step.Ports)
				svc, err := Service(e.namespace, step.Name, step.Alias, []string{})
				if err != nil {
					return err
				}
				if err := e.client.CoreV1().Services(e.namespace).Delete(ctx, svc.Name, deleteOpts); err != nil {
					return err
				}
			}
		}
	}

	for _, vol := range conf.Volumes {
		pvc := PersistentVolumeClaim(e.namespace, vol.Name, e.storageClass, e.volumeSize)
		err := e.client.CoreV1().PersistentVolumeClaims(e.namespace).Delete(ctx, pvc.Name, deleteOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

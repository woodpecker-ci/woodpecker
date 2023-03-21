package kubernetes

import (
	"errors"
	"os"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	dnsPattern           = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	ErrDNSPatternInvalid = errors.New("name is not a valid kubernetes DNS name")
)

func dnsName(i string) (string, error) {
	res := strings.Replace(i, "_", "-", -1)

	if found := dnsPattern.FindStringIndex(res); found == nil {
		return "", ErrDNSPatternInvalid
	}

	return res, nil
}

func isImagePullBackOffState(pod *v1.Pod) bool {
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Waiting != nil {
			if containerState.State.Waiting.Reason == "ImagePullBackOff" {
				return true
			}
		}
	}

	return false
}

// getClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func getClientOutOfCluster() (kubernetes.Interface, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// getClient returns a k8s clientset to the request from inside of cluster
func getClientInsideOfCluster() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

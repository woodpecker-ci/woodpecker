package kubernetes

import (
	"log"
	"os"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var dnsPattern = regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`)

func dnsName(i string) string {
	res := strings.Replace(strings.ToLower(i), "_", "-", -1)

	match := dnsPattern.MatchString(res)
	if !match {
		log.Panic("name not valid kubernetes DNS name")
	}
	return res
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

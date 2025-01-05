// Copyright 2022 Woodpecker Authors
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

package kubernetes

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	client_cmd "k8s.io/client-go/tools/clientcmd"
)

var (
	dnsPattern              = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	dnsDisallowedCharacters = regexp.MustCompile(`[^-a-z0-9.]+`)
)

func dnsName(i string) (string, error) {
	res := strings.ToLower(strings.ReplaceAll(i, "_", "-"))

	invalidChars := dnsDisallowedCharacters.FindAllString(res, -1)
	if len(invalidChars) > 0 {
		return "", fmt.Errorf("name is not a valid kubernetes DNS name: found invalid characters '%v'", strings.Join(invalidChars, ""))
	}

	if !dnsPattern.MatchString(res) {
		return "", fmt.Errorf("name is not a valid kubernetes DNS name")
	}

	// k8s pod name limit
	const maxPodNameLength = 253
	if len(res) > maxPodNameLength {
		res = res[:maxPodNameLength]
	}

	return res, nil
}

func toDNSName(in string) (string, error) {
	lower := strings.ToLower(in)
	withoutUnderscores := strings.ReplaceAll(lower, "_", "-")
	withoutSpaces := strings.ReplaceAll(withoutUnderscores, " ", "-")
	almostDNS := dnsDisallowedCharacters.ReplaceAllString(withoutSpaces, "")
	return dnsName(almostDNS)
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

func isInvalidImageName(pod *v1.Pod) bool {
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Waiting != nil {
			if containerState.State.Waiting.Reason == "InvalidImageName" {
				return true
			}
		}
	}

	return false
}

// getClientOutOfCluster returns a k8s client set to the request from outside of cluster.
func getClientOutOfCluster() (kubernetes.Interface, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG") // cspell:words KUBECONFIG
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	// use the current context in kube config
	config, err := client_cmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// getClient returns a k8s client set to the request from inside of cluster.
func getClientInsideOfCluster() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func newBool(val bool) *bool {
	ptr := new(bool)
	*ptr = val
	return ptr
}

func newInt64(val int64) *int64 {
	ptr := new(int64)
	*ptr = val
	return ptr
}

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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"strings"

	kube_core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kube_client_cmd "k8s.io/client-go/tools/clientcmd"
)

const maxDNSLabelLen = 63

var (
	dnsDisallowedCharacters = regexp.MustCompile(`[^-.a-z0-9]+`)
	dotsAndDashes           = regexp.MustCompile(`[.-]{2,}`)
	labelDisallowedChars    = regexp.MustCompile(`[^-_.a-z0-9]+`)
	labelSeparators         = regexp.MustCompile(`[-_.]{2,}`)
	ErrDNSPatternInvalid    = errors.New("name is not a valid kubernetes DNS name")
	ErrLabelInvalid         = errors.New("value is not a valid kubernetes label value")
)

func getHostnameOrEmpty(name string) string {
	clean, _ := toDNSName(name)
	if clean == "" {
		clean = strings.ToLower(name)
	}
	clean = strings.ReplaceAll(clean, ".", "-")

	if len(clean) > maxDNSLabelLen {
		clean = clean[:maxDNSLabelLen]
	}

	clean = strings.Trim(clean, "-")

	if len(validation.IsDNS1123Label(clean)) == 0 {
		return clean
	}
	return ""
}

func toDNSName(in string) (string, error) {
	res := strings.ToLower(in)
	res = dnsDisallowedCharacters.ReplaceAllString(res, "-")
	res = dotsAndDashes.ReplaceAllStringFunc(res, func(s string) string {
		if strings.ContainsRune(s, '.') {
			return "."
		}
		return "-"
	})
	res = strings.Trim(res, "-.")

	if len(res) > validation.DNS1123SubdomainMaxLength {
		res = truncateWithHash(res, in, validation.DNS1123SubdomainMaxLength, "-.")
	}

	if res == "" || len(validation.IsDNS1123Subdomain(res)) > 0 {
		return "", ErrDNSPatternInvalid
	}

	return res, nil
}

func truncateWithHash(s, original string, maxLen int, trimChars string) string {
	hash := sha256.Sum256([]byte(original))
	hashStr := hex.EncodeToString(hash[:])[:16]
	maxBaseLength := maxLen - 1 - len(hashStr)
	truncated := strings.TrimRight(s[:maxBaseLength], trimChars)
	return truncated + "-" + hashStr
}

func toLabelValue(in string) (string, error) {
	res := strings.ToLower(in)
	res = labelDisallowedChars.ReplaceAllString(res, "-")
	res = labelSeparators.ReplaceAllString(res, "-")
	res = strings.Trim(res, "-_.")

	if len(res) > validation.LabelValueMaxLength {
		res = truncateWithHash(res, in, validation.LabelValueMaxLength, "-_.")
	}

	if len(validation.IsValidLabelValue(res)) > 0 {
		return "", ErrLabelInvalid
	}

	return res, nil
}

func isImagePullBackOffState(pod *kube_core_v1.Pod) bool {
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Waiting != nil {
			if containerState.State.Waiting.Reason == "ImagePullBackOff" {
				return true
			}
		}
	}

	return false
}

func isInvalidImageName(pod *kube_core_v1.Pod) bool {
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
	config, err := kube_client_cmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// getClientInsideOfCluster returns a k8s client set to the request from inside of cluster.
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

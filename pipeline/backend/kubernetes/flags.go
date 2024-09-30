// Copyright 2023 Woodpecker Authors
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
	"github.com/urfave/cli/v3"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_NAMESPACE"),
		Name:    "backend-k8s-namespace",
		Usage:   "backend k8s namespace",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_VOLUME_SIZE"),
		Name:    "backend-k8s-volume-size",
		Usage:   "backend k8s volume size (default 10G)",
		Value:   "10G",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_STORAGE_CLASS"),
		Name:    "backend-k8s-storage-class",
		Usage:   "backend k8s storage class",
		Value:   "",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_STORAGE_RWX"),
		Name:    "backend-k8s-storage-rwx",
		Usage:   "backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)",
		Value:   true,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_POD_LABELS"),
		Name:    "backend-k8s-pod-labels",
		Usage:   "backend k8s additional Agent-wide worker pod labels",
		Value:   "",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_POD_LABELS_ALLOW_FROM_STEP"),
		Name:    "backend-k8s-pod-labels-allow-from-step",
		Usage:   "whether to allow using labels from step's backend options",
		Value:   false,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS"),
		Name:    "backend-k8s-pod-annotations",
		Usage:   "backend k8s additional Agent-wide worker pod annotations",
		Value:   "",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_POD_NODE_SELECTOR"),
		Name:    "backend-k8s-pod-node-selector",
		Usage:   "backend k8s Agent-wide worker pod node selector",
		Value:   "",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS_ALLOW_FROM_STEP"),
		Name:    "backend-k8s-pod-annotations-allow-from-step",
		Usage:   "whether to allow using annotations from step's backend options",
		Value:   false,
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_SECCTX_NONROOT"), // cspell:words secctx nonroot
		Name:    "backend-k8s-secctx-nonroot",
		Usage:   "`run as non root` Kubernetes security context option",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES"),
		Name:    "backend-k8s-pod-image-pull-secret-names",
		Usage:   "backend k8s pull secret names for private registries",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_K8S_ALLOW_NATIVE_SECRETS"),
		Name:    "backend-k8s-allow-native-secrets",
		Usage:   "whether to allow existing Kubernetes secrets to be referenced from steps",
		Value:   false,
	},
}

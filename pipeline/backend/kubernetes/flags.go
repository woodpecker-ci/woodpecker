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
	"time"

	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_NAMESPACE"},
		Name:    "backend-k8s-namespace",
		Usage:   "backend k8s namespace",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_VOLUME_SIZE"},
		Name:    "backend-k8s-volume-size",
		Usage:   "backend k8s volume size (default 10G)",
		Value:   "10G",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_CLASS"},
		Name:    "backend-k8s-storage-class",
		Usage:   "backend k8s storage class",
		Value:   "",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_RWX"},
		Name:    "backend-k8s-storage-rwx",
		Usage:   "backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_POD_LABELS"},
		Name:    "backend-k8s-pod-labels",
		Usage:   "backend k8s additional worker pod labels",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS"},
		Name:    "backend-k8s-pod-annotations",
		Usage:   "backend k8s additional worker pod annotations",
		Value:   "",
	},
	&cli.IntFlag{
		EnvVars: []string{"WOODPECKER_CONNECT_RETRY_COUNT"},
		Name:    "connect-retry-count",
		Usage:   "number of times to retry connecting to the server",
		Value:   5,
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_CONNECT_RETRY_DELAY"},
		Name:    "connect-retry-delay",
		Usage:   "duration to wait before retrying to connect to the server",
		Value:   time.Second * 2,
	},
}

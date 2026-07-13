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

package libvirt

import (
	"github.com/urfave/cli/v3"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "backend-libvirt-uri",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_LIBVIRT_URI", "LIBVIRT_URI"),
		Usage:   "Connection URI",
		Value:   "qemu:///system",
	},
	&cli.StringFlag{
		Name:    "backend-libvirt-image-dir",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_LIBVIRT_IMG_DIR"),
		Usage:   "Directory in which libvirt disk images will be created in",
		Value:   "/var/lib/libvirt/images/",
	},
}

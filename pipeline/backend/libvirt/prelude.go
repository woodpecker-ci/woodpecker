// Copyright 2026 Julian Ospald
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
	"fmt"
	"io"
	"os"
	"os/exec"
)

func CmdViaSudo(cmd string, arg ...string) *exec.Cmd {
	if hasCommand("sudo") {
		return exec.Command("sudo", append([]string{cmd}, arg...)...)
	} else {
		return exec.Command(cmd, arg...)
	}
}

func hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func CopyFile(from string, to string, overwrite bool) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	if overwrite == false {
		_, err := os.Stat(to)
		if err == nil {
			return fmt.Errorf("File %s already exists", to)
		}
	}

	toFile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	if err != nil {
		return err
	}

	return nil
}

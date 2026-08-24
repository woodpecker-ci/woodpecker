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
	"math/rand"
	"os"
	"os/exec"
	"time"
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

// derived from https://github.com/rprata/macgo/blob/85ffbfd9b6ec3c18bde5640302adf7521ef1c2c2/generator/generator.go#L11
// MIT licensed
func GenerateRandomMACAddress() (string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var macBytes []byte

	macBytes = make([]byte, 6)

	for i := 0; i < len(macBytes); i++ {
		macBytes[i] = byte(r.Intn(256))
	}

	// we want locally administered unicast
	macBytes[0] |= 0x02
	macBytes[0] &^= 0x01

	macAddress := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		macBytes[0], macBytes[1], macBytes[2], macBytes[3], macBytes[4], macBytes[5])

	return macAddress, nil
}

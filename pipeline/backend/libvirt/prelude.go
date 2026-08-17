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

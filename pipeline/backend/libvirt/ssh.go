package libvirt

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/melbahja/goph"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/unicode"
)

func AdhocSSH(ctx context.Context, client *goph.Client, cmd string, args []string, taskUUID string, stepUUID string) error {
	sshCmd, err := client.CommandContext(ctx, cmd, args...)
	log.Debug().Msgf("Executing via ssh: %s %v", cmd, args)
	if err != nil {
		return err
	}
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	err = sshCmd.Run()

	return err
}

// This tries to terminate the command we spawned, not the SSH session itself.
// It does so, by:
//
// 1. send SIGINT... if that doesn't do it
// 2. send SIGTERM... if that doesn't do it
// 3. execute 'kill -2 <pid>'
//
// We could also try to attach to stdin and send 'Ctrl+c', but that opens up other issues (closing channel
// at the appropriate time etc.).
func (e *libvirt) TerminateSshCommand(options BackendOptions, client *goph.Client, sshCmd *goph.Cmd, guestOS string, taskUUID string, stepUUID string) error {

	pid, err := GetWoodpeckerPid(client, guestOS, stepUUID)
	if err != nil {
		return err
	}
	log.Debug().Msgf("PID: %s", pid)

	if guestOS == "windows" {
		// on windows...  we just do taskkill
		// SIGINT works, but doesn't seem to kill the whole process group
		rawCmd := fmt.Sprintf("taskkill /PID %s /F /T", pid)
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		utf16leBytes, err := encoder.Bytes([]byte(rawCmd))
		if err != nil {
			return err
		}
		base64Cmd := base64.StdEncoding.EncodeToString(utf16leBytes)
		sshCmd, err := client.Command("powershell.exe", "-noprofile", "-noninteractive", "-encodedcommand", base64Cmd)
		if err != nil {
			return err
		}
		err = sshCmd.Run()
		if err != nil {
			return err
		}

		running, err := CheckSshPid(pid, client, guestOS, stepUUID)
		if err != nil {
			return err
		}
		if running {
			return fmt.Errorf("All methods exhausted. Failed to stop SSH process!")
		}

	} else {
		// try sigint first
		log.Debug().Msg("Context canceled, sending SIGINT to remote process")
		err := sshCmd.Signal(ssh.SIGINT)
		if err != nil {
			log.Debug().Msgf("Failed to send SIGINT to remote process: %s", err)
		}

		// check if the process died
		running, err := CheckSshPid(pid, client, guestOS, stepUUID)
		if err != nil {
			return err
		}
		// try SIGTERM
		if running {
			log.Debug().Msg("SIGINT didn't work, trying SIGTERM")
			err := sshCmd.Signal(ssh.SIGTERM)
			if err != nil {
				log.Debug().Msgf("Failed to send SIGTERM to remote process: %s", err)
			}

			// check if the process died
			running, err := CheckSshPid(pid, client, guestOS, stepUUID)
			if err != nil {
				return err
			}

			if running {
				log.Debug().Msg("SIGTERM didn't work, sending kill -2 manually")

				// try kill -2 pid
				sshCmd, err := client.Command("kill", "-2", pid)
				if err != nil {
					return err
				}
				err = sshCmd.Run()
				if err != nil {
					return err
				}

				// check if the process died
				running, err := CheckSshPid(pid, client, guestOS, stepUUID)
				if err != nil {
					return err
				}
				if running {
					return fmt.Errorf("All methods exhausted. Failed to stop SSH process!")
				}
			}
		}
	}

	return nil
}

func GetWoodpeckerPid(client *goph.Client, guestOS string, stepUUID string) (string, error) {
	var sshCmd *goph.Cmd

	// assembler the per-platform command to get the running pid
	if guestOS == "windows" {
		rawCmd := fmt.Sprintf("Get-Content -Path $env:TEMP\\woodpecker_%s.pid -Raw", stepUUID)
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		utf16leBytes, err := encoder.Bytes([]byte(rawCmd))
		if err != nil {
			return "", err
		}
		base64Cmd := base64.StdEncoding.EncodeToString(utf16leBytes)

		sshCmd, err = client.Command("powershell.exe", "-noprofile", "-noninteractive", "-encodedcommand", base64Cmd)
		if err != nil {
			return "", err
		}
	} else {
		var err error
		sshCmd, err = client.Command("/bin/sh", "-c", fmt.Sprintf("'cat ${TMPDIR:-/tmp}/woodpecker_%s.pid'", stepUUID))
		if err != nil {
			return "", err
		}
	}

	// extract stdout
	bytes, err := sshCmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// (Re-)Check if the spawned process is still running with a timeout of 15s.
func CheckSshPid(pid string, client *goph.Client, guestOS string, stepUUID string) (bool, error) {
	var sshCmd *goph.Cmd
	maxBackOff, _ := time.ParseDuration("15s")

	// assemble the per-platform command to check for pid
	if guestOS == "windows" {
		rawCmd := fmt.Sprintf("Get-Process -Id %s", pid)
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		utf16leBytes, err := encoder.Bytes([]byte(rawCmd))
		if err != nil {
			return true, err
		}
		base64Cmd := base64.StdEncoding.EncodeToString(utf16leBytes)

		sshCmd, err = client.Command("powershell.exe", "-noprofile", "-noninteractive", "-encodedcommand", base64Cmd)
		if err != nil {
			return true, backoff.Permanent(fmt.Errorf("Error running command: %s", err))
		}
	} else {
		var err error
		sshCmd, err = client.Command("kill", "-0", pid)
		if err != nil {
			return true, backoff.Permanent(fmt.Errorf("Error running command: %s", err))
		}
	}

	// run with exponential backoff
	b, _ := backoff.Retry(context.Background(), func() (bool, error) {
		sshErr := sshCmd.Run()
		// if we can't figure out if it's running... assume it is
		switch sshErr.(type) {
		case *ssh.ExitMissingError:
			log.Debug().Msgf("ExitMissingError in CheckSshPid: %s", sshErr)
			return true, nil
		case *ssh.ExitError:
			log.Debug().Msgf("ExitError in CheckSshPid: %s", sshErr)
			return false, nil
		case nil:
			return true, fmt.Errorf("Process still running")
		default:
			log.Debug().Msgf("Unexpected error in CheckSshPid: %s", sshErr)
			return true, nil
		}

	}, backoff.WithMaxElapsedTime(maxBackOff))

	return b, nil
}

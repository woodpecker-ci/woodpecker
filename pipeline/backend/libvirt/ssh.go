package libvirt

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/melbahja/goph"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
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
// 3. send the 'ctrl+c' character to stdin
func (e *libvirt) TerminateSshCommand(options BackendOptions, client *goph.Client, sshCmd *goph.Cmd, guestOS string, taskUUID string, stepUUID string) error {
	log.Debug().Msg("Context canceled, sending SIGINT to remote process")
	err := sshCmd.Signal(ssh.SIGINT)
	if err != nil {
		log.Debug().Msgf("Failed to send SIGINT to remote process: %s", err)
	}

	// check if the process died
	running, err := CheckSshPid(client, guestOS, stepUUID)
	if err != nil {
		return err
	}
	if running {
		log.Debug().Msg("SIGINT didn't work, trying SIGTERM")
		err := sshCmd.Signal(ssh.SIGTERM)
		if err != nil {
			log.Debug().Msgf("Failed to send SIGTERM to remote process: %s", err)
		}

		// check if the process died
		running, err := CheckSshPid(client, guestOS, stepUUID)
		if err != nil {
			return err
		}

		if running {
			if options.SSHConfig.Tty {
				log.Debug().Msg("SIGTERM didn't work, sending Ctrl+c to stdin")
				w, ok := e.workflows.Load(taskUUID)
				if !ok {
					return fmt.Errorf("Could not find key %s for workflows", taskUUID)
				}

				p, ok := w.(*workflow).pipesStdIn.Load(stepUUID)
				if !ok {
					return fmt.Errorf("Could not find key %s for pipesStdIn", stepUUID)
				}

				p.(*pipes).pw.Write([]byte("\x03"))
				time.Sleep(time.Second * 2)
				p.(*pipes).pw.Close()
				p.(*pipes).pr.Close()

				// check if the process died
				running, err := CheckSshPid(client, guestOS, stepUUID)
				if err != nil {
					return err
				}
				if running {
					return fmt.Errorf("Failed to stop SSH process!")
				}

			} else {
				log.Debug().Msg("No tty allocated... skip sending Ctrl+c")
				return fmt.Errorf("Failed to stop SSH process!")

			}
		}

	}

	return nil
}

// (Re-)Check if the spawned process is still running with a timeout of 10s.
func CheckSshPid(client *goph.Client, guestOS string, stepUUID string) (bool, error) {
	maxBackOff, _ := time.ParseDuration("10s")

	if guestOS == "windows" {
		// TODO
		return true, nil
	} else {
		sshCmd, err := client.Command("/bin/sh", "-c", fmt.Sprintf("'cat ${TMPDIR:-/tmp}/%s.pid'", stepUUID))
		if err != nil {
			return false, err
		}
		bytes, err := sshCmd.Output()
		if err != nil {
			return false, err
		}
		pid := string(bytes)

		log.Debug().Msgf("Pid: %s", pid)

		b, _ := backoff.Retry(context.Background(), func() (bool, error) {
			sshCmd, err := client.Command("kill", "-0", pid)
			if err != nil {
				return true, backoff.Permanent(fmt.Errorf("Error running command: %s", err))
			}

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
}

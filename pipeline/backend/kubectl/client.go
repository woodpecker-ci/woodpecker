package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type KubeClient struct {
	Executable                      string        // the default executable
	Namespace                       string        // the default namespace
	Context                         string        // the default context
	RequestTimeout                  time.Duration // The kubectl request timeout
	AllowKubectlClientConfiguration bool          // If true, Allows configurations like --request-timeout
}

// Loads the client.
func (client *KubeClient) Load(ctx context.Context) error {
	err := client.LoadDefaults(ctx)
	if err != nil {
		return err
	}

	return nil
}

// The kubectl executable path.
func (client *KubeClient) GetExecutablePath() string {
	if len(client.Executable) == 0 {
		return "kubectl"
	}
	return client.Executable
}

// run a kubectl command
func (client *KubeClient) RunKubectlCommand(
	ctx context.Context, args ...interface{},
) (string, error) {
	cmnd := client.CreateKubectlCommand(ctx, args...)

	rslt, cmdErr := cmnd.Output()
	if cmdErr != nil {
		switch err := cmdErr.(type) {
		case *exec.ExitError:
			if len(err.Stderr) > 0 {
				cmdErr = errors.New(string(err.Stderr) + "(" + ")")
			}
		}

		log.Debug().Err(cmdErr).Str("Args", strings.Join(cmnd.Args, " ")).Msg(
			"kubectl command failed",
		)
		return "", cmdErr
	}
	return string(rslt), cmdErr
}

// Creates a new kubectl exec command.
func (client *KubeClient) CreateKubectlCommand(
	ctx context.Context,
	args ...interface{},
) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		client.GetExecutablePath(),
		client.ComposeKubectlCommand(args...)...,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// run in current environment.
	cmd.Env = os.Environ()

	return cmd
}

// Compose a new kubectl command from a list of args.
// The args can be either string|[]string
func (client *KubeClient) ComposeKubectlCommand(args ...interface{}) []string {
	flat := []string{}
	for _, rawAr := range args {
		switch ar := rawAr.(type) {
		case string:
			flat = append(flat, ar)
		case []string:
			flat = append(flat, ar...)
		default:
			break
		}
	}

	command := []string{}
	for _, ar := range flat {
		if len(ar) == 0 {
			continue
		}
		command = append(command, ar)
	}

	hasArg := func(markers ...string) bool {
		for _, marker := range markers {
			for _, ar := range command {
				if marker == ar {
					return true
				}
			}
		}
		return false
	}

	if client.Namespace != "" && !hasArg("--namespace", "-n") {
		command = append([]string{
			"--namespace",
			client.Namespace,
		}, command...)
	}

	if client.Context != "" && !hasArg("--context") {
		command = append([]string{
			"--context",
			client.Context,
		}, command...)
	}

	if client.AllowKubectlClientConfiguration {
		if client.RequestTimeout.Seconds() > 0 && !hasArg("--request-timeout") {
			command = append([]string{
				"--request-timeout",
				fmt.Sprintf("%ds", int(client.RequestTimeout.Seconds())),
			}, command...)
		}
	}

	return command
}

// Loads default configuration for the client.
func (client *KubeClient) LoadDefaults(ctx context.Context) error {
	if len(client.Context) == 0 {
		context, err := client.RunKubectlCommand(
			ctx, "config", "current-context",
		)
		if err == nil {
			client.Context = strings.TrimSpace(context)
		}
	}

	if len(client.Namespace) == 0 {
		namespace, err := client.RunKubectlCommand(
			ctx, "config", "view", "--minify", "--output",
			"jsonpath={.contexts[0].context.namespace}",
		)
		if err == nil {
			client.Namespace = strings.TrimSpace(namespace)
		}
	}

	return nil
}

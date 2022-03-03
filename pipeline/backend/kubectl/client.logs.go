package kubectl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
)

func (client *KubeClient) ReadResourceLogs(
	ctx context.Context,
	resourceName string,
	follow bool,
	since string,
) (*io.PipeReader, chan error) {
	reader, writer := io.Pipe()
	errChan := client.ReadResourceLogsToWriter(ctx, writer, resourceName, follow, since)
	return reader, errChan
}

// Returns the pod ip for the pod resource.
func (client *KubeClient) ReadResourceLogsToWriter(
	ctx context.Context,
	writer io.Writer,
	resourceName string,
	follow bool,
	since string,
) chan error {
	cmndArgs := []interface{}{}

	if client.AllowKubectlClientConfiguration {
		cmndArgs = append(cmndArgs,
			"--request-timeout",
			fmt.Sprintf("%ds", (60*60*24)),
		)
	}

	cmndArgs = append(cmndArgs, "logs")

	if len(since) > 0 {
		cmndArgs = append(cmndArgs, "--since", since)
	}
	if follow {
		cmndArgs = append(cmndArgs, "-f")
	}

	cmndArgs = append(cmndArgs, resourceName)

	errChan := make(chan error)
	waitForCommandToStart := WaitOnce{}
	stderr := &bytes.Buffer{}

	cmnd := client.CreateKubectlCommand(ctx, cmndArgs...)
	cmnd.Stdout = writer
	cmnd.Stderr = stderr

	stop := func(err error) {
		waitForCommandToStart.MarkComplete(err)
		if err != nil && stderr.Len() > 0 {
			err = errors.New(stderr.String() + "; " + err.Error())
		}
		errChan <- err
	}

	go func() {
		err := cmnd.Start()
		if err != nil {
			stop(err)
			return
		}
		waitForCommandToStart.MarkComplete(err)
		err = cmnd.Wait()
		stop(err)
	}()

	_ = waitForCommandToStart.Wait()

	return errChan
}

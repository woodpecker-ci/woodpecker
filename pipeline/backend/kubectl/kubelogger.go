package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
)

type KubePodLogger struct {
	Backend *KubeBackend // the kubernetes backend
	PodName string       // the name of the pod to read

	// internal properties
	cancelLogContext context.CancelFunc // cancel the executing logs context
	logContext       context.Context    // The active log context.
	lastError        error              // The last logging error
}

func (podLogger *KubePodLogger) IsRunning() bool {
	return podLogger.logContext != nil
}

func (podLogger *KubePodLogger) LastError() error {
	return podLogger.lastError
}

func (podLogger *KubePodLogger) Stop() error {
	if podLogger.IsRunning() {
		podLogger.cancelLogContext()
		podLogger.logContext = nil
		podLogger.cancelLogContext = nil
	}
	return podLogger.lastError
}

func (podLogger *KubePodLogger) Start(ctx context.Context) (*io.PipeReader, error) {
	if podLogger.IsRunning() {
		return nil, errors.New("Pod logger is running. Cannot start")
	}

	logger := podLogger.Backend.MakeLogger("").With().Str("PodName", podLogger.PodName).Logger()

	// initializing.
	podLogger.logContext, podLogger.cancelLogContext = context.WithCancel(ctx)
	podLogger.lastError = nil

	// Pipes and buffers
	logsReaer, logsWriter := io.Pipe() // logs lines output.
	rawReader, rawWriter := io.Pipe()  // logs raw writer/reader
	lineBreak := "\n"
	linesRead := false                         // if lines were read
	lineScanner := bufio.NewScanner(rawReader) // the scanner for lines

	fromStderr := func(err error, stderr string) error {
		if stderr != "" {
			err = errors.New(err.Error() + "\nstderr: " + stderr)
		}
		return err
	}

	stop := func(err error, msg string) {
		_ = logsWriter.Close()
		_ = rawWriter.Close()
		_ = podLogger.Stop()
	}

	go func() {
		for lineScanner.Scan() {
			// mark lines as read.
			linesRead = true
			_, err := logsWriter.Write(append(lineScanner.Bytes(), []byte(lineBreak)...))
			if err != nil {
				stop(err, "Error while reading lines")
			}
		}
	}()

	go func() {
		restarts := 0
		// this needs a loop since the logger here
		// may fail reading the logs.
		// in that case we are required to restart the logger.
		for {
			logsCmd := podLogger.Backend.Client.CreateKubectlCommand(
				podLogger.logContext,
				"logs",
				podLogger.PodName,
				"-f",
			)

			stdErrPipe, err := logsCmd.StderrPipe()
			if err != nil {
				stop(err, "Error creating pipe")
				break
			}

			logsCmd.Stdout = rawWriter

			err = logsCmd.Run()

			if err == context.Canceled {
				// the context was canceled.
				logger.Warn().Msg("Logger context canceled. Aborting read")
				stop(nil, "")
				break
			}

			if err != nil {
				// something else went wrong.
				stderr, _ := GetReaderContents(stdErrPipe)
				err := fromStderr(err, stderr)

				if linesRead {
					stop(err, "Error while reading logs")
					break
				}

				restarts++
				if restarts > podLogger.Backend.LogStartAttempts {
					stop(
						err,
						fmt.Sprintf(
							"Error starting log reading. Too many attempts (%d)",
							podLogger.Backend.LogStartAttempts,
						),
					)
					break
				}

				logger.Debug().Err(fromStderr(err, stderr)).Msg(
					fmt.Sprintf(
						"Failed to start logger. Retry, %d/%d",
						restarts, podLogger.Backend.LogStartAttempts,
					),
				)
				continue
			}
			// completed. Stopping
			stop(nil, "Log reading complete")
			break
		}
	}()

	return logsReaer, nil
}

func (podLogger *KubePodLogger) Done() <-chan struct{} {
	return podLogger.logContext.Done()
}

func (podLogger *KubePodLogger) Wait() error {
	if !podLogger.IsRunning() {
		return errors.New(
			"Pod logger is not running. Cannot wait",
		)
	}
	<-podLogger.logContext.Done()
	return podLogger.lastError
}

func (podLogger *KubePodLogger) ReadWithContext(ctx context.Context) (string, error) {
	reader, err := podLogger.Start(ctx)
	if err != nil {
		return "", err
	}
	err = podLogger.Wait()
	if err != nil {
		return "", err
	}

	output, err := GetReaderContents(reader)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (podLogger *KubePodLogger) Read(ctx context.Context) (string, error) {
	return podLogger.ReadWithContext(context.Background())
}

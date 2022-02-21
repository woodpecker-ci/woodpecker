package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type KubeResourceLogger struct {
	Backend      *KubeBackend // the kubernetes backend
	ResourceName string       // the name of the pod to read

	// internal properties
	cancelLogContext context.CancelFunc // cancel the executing logs context
	logContext       context.Context    // The active log context.
	lastError        error              // The last logging error
}

func (rLogger *KubeResourceLogger) IsRunning() bool {
	return rLogger.logContext != nil
}

func (rLogger *KubeResourceLogger) LastError() error {
	return rLogger.lastError
}

func (rLogger *KubeResourceLogger) Stop() error {
	if rLogger.IsRunning() {
		rLogger.cancelLogContext()
		rLogger.logContext = nil
		rLogger.cancelLogContext = nil
	}
	return rLogger.lastError
}

func (rLogger *KubeResourceLogger) Start(ctx context.Context) (*io.PipeReader, error) {
	if rLogger.IsRunning() {
		return nil, errors.New("Pod logger is running. Cannot start")
	}

	logger := rLogger.Backend.MakeLogger("").With().Str(
		"Resource", rLogger.ResourceName,
	).Logger()

	// initializing.
	rLogger.logContext, rLogger.cancelLogContext = context.WithCancel(ctx)
	rLogger.lastError = nil

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
		_ = rLogger.Stop()
		if err != nil {
			logger.Error().Err(err).Msg(msg)
		}
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
			logsCmd := rLogger.Backend.Client.CreateKubectlCommand(
				rLogger.logContext,
				"logs",
				rLogger.ResourceName,
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
				if restarts > rLogger.Backend.LogStartAttempts {
					stop(
						err,
						fmt.Sprintf(
							"Error starting log reading. Too many attempts (%d)",
							rLogger.Backend.LogStartAttempts,
						),
					)
					break
				}

				logger.Debug().Err(fromStderr(err, stderr)).Msg(
					fmt.Sprintf(
						"Failed to start logger. Retry in %.2f [second], %d/%d",
						rLogger.Backend.LogAttemptWait.Seconds(),
						restarts,
						rLogger.Backend.LogStartAttempts,
					),
				)

				// sleep before next attempt.
				time.Sleep(rLogger.Backend.LogAttemptWait)
				continue
			}
			// completed. Stopping
			stop(nil, "Log reading complete")
			break
		}
	}()

	return logsReaer, nil
}

func (rLogger *KubeResourceLogger) Done() <-chan struct{} {
	return rLogger.logContext.Done()
}

func (rLogger *KubeResourceLogger) Wait() error {
	if !rLogger.IsRunning() {
		return errors.New(
			"Pod logger is not running. Cannot wait",
		)
	}
	<-rLogger.logContext.Done()
	return rLogger.lastError
}

func (rLogger *KubeResourceLogger) ReadWithContext(ctx context.Context) (string, error) {
	reader, err := rLogger.Start(ctx)
	if err != nil {
		return "", err
	}
	err = rLogger.Wait()
	if err != nil {
		return "", err
	}

	output, err := GetReaderContents(reader)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (rLogger *KubeResourceLogger) Read(ctx context.Context) (string, error) {
	return rLogger.ReadWithContext(context.Background())
}

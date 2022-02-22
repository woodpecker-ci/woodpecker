package kubectl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type KubeResourceLogger struct {
	Backend      *KubeBackend // the kubernetes backend
	ResourceName string       // the name of the resource to read

	// internal properties
	cancelLogContext context.CancelFunc // cancel the executing logs context
	logContext       context.Context    // The active log context.
	lastError        error              // The last logging error
}

func (resLogger *KubeResourceLogger) IsRunning() bool {
	return resLogger.logContext != nil
}

func (resLogger *KubeResourceLogger) LastError() error {
	return resLogger.lastError
}

func (resLogger *KubeResourceLogger) Stop() error {
	if resLogger.IsRunning() {
		cancel := resLogger.cancelLogContext
		// clear.
		resLogger.logContext = nil
		resLogger.cancelLogContext = nil
		// cancel existing
		cancel()
	}
	return resLogger.lastError
}

func (resLogger *KubeResourceLogger) Start(ctx context.Context) (*io.PipeReader, error) {
	if resLogger.IsRunning() {
		return nil, errors.New("Resource logger is running. Cannot start")
	}

	logger := resLogger.Backend.MakeLogger("").With().Str(
		"Resource", resLogger.ResourceName,
	).Logger()

	// initializing.
	logContext, cancelLogContext := context.WithCancel(ctx)
	resLogger.logContext = logContext
	resLogger.cancelLogContext = cancelLogContext
	resLogger.lastError = nil

	// Pipes and buffers
	logsReaer, logsWriter := io.Pipe() // logs lines output.
	rawReader, rawWriter := io.Pipe()  // logs raw writer/reader
	lineBreak := "\n"
	lineScanner := bufio.NewScanner(rawReader) // the scanner for lines

	fromStderr := func(err error, stderr string) error {
		if stderr != "" {
			err = errors.New(err.Error() + "\nstderr: " + stderr)
		}
		return err
	}

	writeLine := func(line []byte) error {
		_, err := logsWriter.Write(append(line, []byte(lineBreak)...))
		return err
	}

	stop := func(err error, msg string) {
		if err != nil {
			logger.Error().Err(err).Msg(msg)
			_ = writeLine([]byte(
				fmt.Sprintf("Error reading logs from stage (%s): %s", resLogger.ResourceName, msg),
			))
			_ = writeLine([]byte(err.Error()))
		}
		_ = logsWriter.Close()
		_ = rawWriter.Close()
		_ = resLogger.Stop()
	}

	// listen for context cancel.
	go func() {
		<-logContext.Done()
		if resLogger.IsRunning() {
			err := resLogger.Stop()
			debug := logger.Debug()
			if err != nil {
				debug = debug.Err(err)
			}
			debug.Msg("Resource logger context was canceled. Resource logger stopped.")
		}
	}()

	lastLineScanned := int64(0)

	consecutiveRestarts := 0

	// listen for lines.
	go func() {
		for lineScanner.Scan() {
			// line was read. Reseting restarts and time.
			consecutiveRestarts = 0
			lastLineScanned = time.Now().Unix()

			err := writeLine(lineScanner.Bytes())
			if err != nil {
				stop(err, "Error while reading lines")
			}
		}
	}()

	go func() {
		// this needs a loop since the logging command may fail.
		for {
			// If is not running. Must return.
			sinceArg := []string{}
			if lastLineScanned > 0 {
				sinceArg = []string{
					"--since",
					fmt.Sprintf("%ds", time.Now().Unix()-lastLineScanned),
				}
			}

			logsCmd := resLogger.Backend.Client.CreateKubectlCommand(
				logContext,
				"--request-timeout",
				sinceArg,
				fmt.Sprintf("%ds", (60*60*24)),
				"logs",
				resLogger.ResourceName,
				"-f",
			)

			var stdErrBuffer bytes.Buffer

			logsCmd.Stderr = &stdErrBuffer
			logsCmd.Stdout = rawWriter

			err := logsCmd.Run()

			if err == context.Canceled {
				// the context was canceled.
				logger.Warn().Msg("Logger context canceled. Aborting read")
				stop(nil, "")
				break
			}

			if err != nil {
				// something else went wrong.
				stderr := stdErrBuffer.String()
				err := fromStderr(err, stderr)

				if !resLogger.IsRunning() {
					// terminated.
					break
				}

				consecutiveRestarts++
				if consecutiveRestarts > resLogger.Backend.CommandRetries {
					stop(
						err,
						fmt.Sprintf(
							"Error starting log reading. Too many attempts (%d)",
							resLogger.Backend.CommandRetries,
						),
					)
					break
				}

				logger.Debug().Err(fromStderr(err, stderr)).Msg(
					fmt.Sprintf(
						"Logger failed. Retry in %.2f [second], %d/%d",
						resLogger.Backend.CommandRetryWait.Seconds(),
						consecutiveRestarts,
						resLogger.Backend.CommandRetries,
					),
				)

				// sleep before next attempt.
				time.Sleep(resLogger.Backend.CommandRetryWait)
				continue
			}
			// completed. Stopping
			stop(nil, "")
			break
		}
	}()

	return logsReaer, nil
}

func (resLogger *KubeResourceLogger) Done() <-chan struct{} {
	return resLogger.logContext.Done()
}

func (resLogger *KubeResourceLogger) Wait() error {
	if !resLogger.IsRunning() {
		return errors.New(
			"Resource logger is not running",
		)
	}
	<-resLogger.logContext.Done()
	return resLogger.lastError
}

func (resLogger *KubeResourceLogger) ReadWithContext(ctx context.Context) (string, error) {
	reader, err := resLogger.Start(ctx)
	if err != nil {
		return "", err
	}
	err = resLogger.Wait()
	if err != nil {
		return "", err
	}

	output, err := GetReaderContents(reader)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (resLogger *KubeResourceLogger) Read(ctx context.Context) (string, error) {
	return resLogger.ReadWithContext(context.Background())
}

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

const LineBreak = "\n"

type KubeResourceLogger struct {
	Backend      *KubeBackend // the kubernetes backend
	ResourceName string       // the name of the resource to read

	// internal properties
	stopLogger context.CancelFunc // cancel the executing logs context
	logContext context.Context    // The active log context.
	lastError  error              // The last logging error
	isRunning  bool               // If true, the logger is running
}

func (resLogger *KubeResourceLogger) IsRunning() bool {
	return resLogger.isRunning
}

func (resLogger *KubeResourceLogger) LastError() error {
	return resLogger.lastError
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

func (resLogger *KubeResourceLogger) Stop() error {
	if resLogger.IsRunning() {
		resLogger.stopLogger()
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
	logsReaer, logsWriter := io.Pipe()         // logs lines output.
	rawReader, rawWriter := io.Pipe()          // logs raw writer/reader
	lineScanner := bufio.NewScanner(rawReader) // the scanner for lines

	fromStderr := func(err error, stderr string) error {
		if stderr != "" {
			err = errors.New(err.Error() + "\nstderr: " + stderr)
		}
		return err
	}

	writeLine := func(line []byte) error {
		_, err := logsWriter.Write(append(line, []byte(LineBreak)...))
		return err
	}

	// Main stopWithError function.
	// When the log is canceled this is called.
	stopWithError := func(err error, msg string) {
		if !resLogger.IsRunning() {
			return
		}

		resLogger.isRunning = false
		cancelLogContext()

		if err != nil {
			logger.Error().Err(err).Msg(msg)
			_ = writeLine([]byte(
				fmt.Sprintf("Logger stopped. Error reading logs from stage (%s): %s", resLogger.ResourceName, msg),
			))
			_ = writeLine([]byte(err.Error()))
		} else {
			logger.Debug().Msg("Logger stopped")
		}

		_ = logsWriter.Close()
		_ = rawWriter.Close()

		// the last error will be defined as the stop error. If any
		resLogger.lastError = err
	}

	stop := func() {
		stopWithError(nil, "")
	}

	resLogger.isRunning = true
	resLogger.logContext = logContext
	resLogger.stopLogger = stop
	resLogger.lastError = nil

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
				stopWithError(err, "Error while reading lines")
			}
		}
		logger.Debug().Msg("Logger line listener exited.")
	}()

	go func() {
		// this needs a loop since the logging command may fail.
		for {
			// If is not running. Must return.
			extraArgs := []string{}
			if lastLineScanned > 0 {
				extraArgs = append(extraArgs,
					"--since",
					fmt.Sprintf("%ds", time.Now().Unix()-lastLineScanned),
				)
			}

			if resLogger.Backend.Client.AllowClientConfiguration {
				extraArgs = append(extraArgs,
					"--request-timeout",
					fmt.Sprintf("%ds", (60*60*24)),
				)
			}

			logsCmd := resLogger.Backend.Client.CreateKubectlCommand(
				logContext,
				extraArgs,
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
				stop()
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
					stopWithError(
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

			stop()
			break
		}
	}()

	return logsReaer, nil
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

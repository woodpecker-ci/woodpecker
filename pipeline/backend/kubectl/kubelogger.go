package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

const LineBreak = "\n"

type KubeResourceLogger struct {
	Run          *KubePiplineRun // the kubernetes backend
	ResourceName string          // the name of the resource to read

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

	logger := resLogger.Run.MakeLogger(nil).With().Str(
		"Resource", resLogger.ResourceName,
	).Logger()

	// initializing.
	logContext, cancelLogContext := context.WithCancel(ctx)
	logsReaer, logsWriter := io.Pipe()         // logs lines output.
	rawReader, rawWriter := io.Pipe()          // logs raw writer/reader
	lineScanner := bufio.NewScanner(rawReader) // the scanner for lines

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
				fmt.Sprintf(
					"Logger stopped with ERROR (resource: %s). %s",
					resLogger.ResourceName,
					msg,
				),
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
			logsError := resLogger.Run.Backend.Client.ReadResourceLogsToWriter(
				logContext,
				rawWriter,
				resLogger.ResourceName,
				true,
				Triary(lastLineScanned <= 0, "",
					fmt.Sprintf("%ds", time.Now().Unix()-lastLineScanned),
				).(string),
			)

			err := <-logsError

			if err == context.Canceled {
				// the context was canceled.
				stop()
				break
			}

			if err != nil {
				if !resLogger.IsRunning() {
					// terminated.
					break
				}

				consecutiveRestarts++
				if consecutiveRestarts > resLogger.Run.Backend.CommandRetries {
					stopWithError(
						err,
						fmt.Sprintf(
							"Error starting log reading. Too many attempts (%d)",
							resLogger.Run.Backend.CommandRetries,
						),
					)
					break
				}

				logger.Debug().Err(err).Msg(
					fmt.Sprintf(
						"Logger failed. Retry in %.2f [second], %d/%d",
						resLogger.Run.Backend.CommandRetryWait.Seconds(),
						consecutiveRestarts,
						resLogger.Run.Backend.CommandRetries,
					),
				)

				// sleep before next attempt.
				time.Sleep(resLogger.Run.Backend.CommandRetryWait)
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

	output, err := ReadPipeAsString(reader)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (resLogger *KubeResourceLogger) Read(ctx context.Context) (string, error) {
	return resLogger.ReadWithContext(context.Background())
}

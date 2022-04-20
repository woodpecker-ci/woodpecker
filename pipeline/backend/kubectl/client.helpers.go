package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Get resource names from selector (with kind)
func (client *KubeClient) GetResourceNames(
	ctx context.Context,
	resourceType string,
	selector string,
) ([]string, error) {
	resourceNames := []string{}
	output, err := client.RunKubectlCommand(
		ctx,
		"get", resourceType,
		"-o",
		`custom-columns=:kind,:metadata.name`,
		"-l", selector,
	)
	if err != nil {
		return resourceNames, err
	}

	lineSplit, err := regexp.Compile(`\s+`)
	if err != nil {
		return resourceNames, err
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		lineArgs := lineSplit.Split(line, -1)
		if len(lineArgs) != 2 {
			continue
		}
		resourceNames = append(
			resourceNames,
			fmt.Sprintf("%s/%s", strings.ToLower(lineArgs[0]), lineArgs[1]),
		)
	}
	return resourceNames, nil
}

// Wait for a resource conditions to have conditions.
// resource = [kind]/[resource name]
// if kind not provided assumes pod
func (client *KubeClient) WaitForConditions(
	ctx context.Context,
	resource string,
	conditions []string,
	count int,
) chan struct {
	condition string
	err       error
} {
	resultChan := make(chan struct {
		condition string
		err       error
	})

	waitCommandArgs := client.ComposeKubectlCommand(
		"wait",
		"--timeout", fmt.Sprint(60*60*24*7)+"s",
		resource,
	)

	action := ActionContext{}
	action.OnStop = func(err error) {
		if err != nil {
			resultChan <- struct {
				condition string
				err       error
			}{
				err: err,
			}
		}
	}

	action.Start(
		ctx,
		func() {
			for _, condition := range conditions {
				waitCommand := client.CreateKubectlCommand(
					action.Context(),
					waitCommandArgs,
					"--for",
					"condition="+condition,
				)

				go func(condition string) {
					err := waitCommand.Start()
					if err != nil {
						action.Stop(err)
						return
					}
					action.MarkActionStarted()
					err = waitCommand.Wait()
					wasStopped := action.Stop(err)

					// stop and check if it was stopped
					// return if there was an error as well.
					if !wasStopped {
						return
					}

					resultChan <- struct {
						condition string
						err       error
					}{
						condition: condition,
						err:       err,
					}
				}(condition)
			}

			// wait for the action to be stopped by one of the internal functions.
			_ = action.Wait()
		},
	)

	_ = action.WaitForActionStarted()

	return resultChan
}

// Wait for events for a specific resource name regex
// By reading all events from this point in time on.
func (client *KubeClient) WaitForResourceEvents(
	ctx context.Context,
	resourceNameRegex string,
	matchEventNames []string,
	count int, // number of events to match
) chan struct {
	events []string
	err    error
} {
	resultChan := make(chan struct {
		events []string
		err    error
	})

	eventsMatched := []string{}
	action := ActionContext{}
	splitBySpaces := regexp.MustCompile(`\s+`)

	var resourceRegex *regexp.Regexp
	var err error

	addLineIfMatches := func(line string) {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			return
		}

		eventArgs := splitBySpaces.Split(line, -1)
		if len(eventArgs) < 2 {
			return
		}

		resource := eventArgs[0]
		eventName := eventArgs[1]

		if !resourceRegex.Match([]byte(resource)) {
			return
		}

		for _, name := range matchEventNames {
			if name == eventName {
				eventsMatched = append(eventsMatched, eventName)
				return
			}
		}
	}
	var eventsCmnd *exec.Cmd
	action.OnStop = func(err error) {
		if len(eventsMatched) < count {
			message := "Error, context canceled before sufficient events matched"
			if err != nil {
				message += ". " + err.Error()
			}
			err = errors.New(message)
		}

		// kill the process if currently executing
		_ = eventsCmnd.Process.Kill()
		_ = eventsCmnd.Wait()

		resultChan <- struct {
			events []string
			err    error
		}{err: err, events: eventsMatched}
	}

	action.Start(
		ctx,
		func() {
			eventsCmnd = client.CreateKubectlCommand(
				action.Context(),
				"get", "events",
				"--watch=true", "--watch-only=true",
				"--output", `custom-columns=:involvedObject.name,:reason`,
			)

			resourceRegex, err = regexp.Compile(resourceNameRegex)
			if err != nil {
				action.Stop(err)
				return
			}

			stdoutPipe, err := eventsCmnd.StdoutPipe()
			if err != nil {
				action.Stop(err)
				return
			}

			lineScanner := bufio.NewScanner(stdoutPipe)
			err = eventsCmnd.Start()
			if err != nil {
				action.Stop(err)
				return
			}

			action.MarkActionStarted()

			for lineScanner.Scan() {
				addLineIfMatches(lineScanner.Text())
				if len(eventsMatched) >= count {
					action.Stop(nil)
					break
				}
			}
		},
	)

	_ = action.WaitForActionStarted()

	return resultChan
}

// Returns the pod ip for the pod resource.
func (client *KubeClient) GetPodIP(ctx context.Context, podName string) (string, error) {
	podIP, err := client.RunKubectlCommand(
		ctx, "get", podName,
		"-o",
		"custom-columns=:status.podIP",
	)
	if err != nil {
		return "", err
	}

	podIP = strings.TrimSpace(podIP)
	if !IsIP(podIP) {
		return "", errors.New("Query returned invalid ip: " + podIP)
	}

	return podIP, nil
}

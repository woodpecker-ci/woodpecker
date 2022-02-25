package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
	waitForCommandToStart := WaitOnce{}
	resultChan := make(chan struct {
		condition string
		err       error
	})

	completed := false
	waitContext, cancel := context.WithCancel(ctx)
	waitCommand := client.ComposeKubectlCommand(
		"wait",
		"--timeout", fmt.Sprint(60*60*24*7)+"s",
		resource,
	)

	// TODO: replace this wait command with a resource get + event watch?

	for _, condition := range conditions {
		cmnd := client.CreateKubectlCommand(
			waitContext,
			waitCommand,
			"--for",
			"condition="+condition,
		)

		go func(condition string) {
			waitForCommandToStart.MarkComplete(nil)
			_, err := cmnd.Output()
			if completed {
				return
			}
			completed = true
			cancel()

			resultChan <- struct {
				condition string
				err       error
			}{
				condition: condition,
				err:       err,
			}
		}(condition)
	}

	// waiting for the first checks to start.
	_ = waitForCommandToStart.Wait()

	// returning the result chan
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
	eventsContext, eventsContextCancel := context.WithCancel(ctx)

	eventsCmnd := client.CreateKubectlCommand(
		eventsContext,
		"get", "events",
		"--watch=true", "--watch-only=true",
		"--output", `custom-columns=:involvedObject.name,:reason`,
	)

	eventsMatched := []string{}

	resultChan := make(chan struct {
		events []string
		err    error
	})

	waitForCommandToStart := WaitOnce{}

	stop := func(err error) {
		eventsContextCancel()
		waitForCommandToStart.MarkComplete(err)

		if len(eventsMatched) < count {
			message := "Error, context canceled before sufficient events matched"
			if err != nil {
				message += ". " + err.Error()
			}
			err = errors.New(message)
		}

		resultChan <- struct {
			events []string
			err    error
		}{err: err, events: eventsMatched}
	}

	splitBySpaces, err := regexp.Compile(`\s+`)
	if err != nil {
		stop(err)
		return resultChan
	}
	resourceRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		stop(err)
		return resultChan
	}

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

	go func() {
		stdoutPipe, err := eventsCmnd.StdoutPipe()
		if err != nil {
			stop(err)
			return
		}

		lineScanner := bufio.NewScanner(stdoutPipe)
		err = eventsCmnd.Start()
		if err != nil {
			stop(err)
			return
		}

		// marking command as started
		waitForCommandToStart.MarkComplete(err)

		for lineScanner.Scan() {
			addLineIfMatches(lineScanner.Text())
			if len(eventsMatched) >= count {
				stop(nil)
				return
			}
		}

		// should have not reached here unless pipe was closed
		// or context have been canceled.

		// wait for context to be done.
		<-eventsContext.Done()

		// checking status.
		stop(eventsContext.Err())
	}()

	// waiting for command to start.
	waitForCommandToStart.Wait()

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

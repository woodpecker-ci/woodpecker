package kubectl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type KubeCtlClientCoreArgs struct {
	Namespace string // the default namespace
	Context   string // the default context
}

func (clientArgs *KubeCtlClientCoreArgs) ToArgsList() []string {
	cmnd := []string{}
	if len(clientArgs.Namespace) > 0 {
		cmnd = append(cmnd, "--namespace", clientArgs.Namespace)
	}
	if len(clientArgs.Context) > 0 {
		cmnd = append(cmnd, "--context", clientArgs.Context)
	}
	return cmnd
}

func (clientArgs *KubeCtlClientCoreArgs) Merge(args KubeCtlClientCoreArgs) KubeCtlClientCoreArgs {
	return KubeCtlClientCoreArgs{
		Namespace: FirstNotEmpty(args.Namespace, clientArgs.Namespace).(string),
		Context:   FirstNotEmpty(args.Context, clientArgs.Context).(string),
	}
}

type KubeCtlClient struct {
	Executable string                // the default executable
	CoreArgs   KubeCtlClientCoreArgs // the default args
}

func (client *KubeCtlClient) GetExecutable() string {
	if len(client.Executable) == 0 {
		return "kubectl"
	}
	return client.Executable
}

// run a kubectl command
func (client *KubeCtlClient) RunKubectlCommand(
	ctx context.Context, args ...interface{},
) (string, error) {
	cmnd := client.CreateKubectlCommand(ctx, args...)
	rslt, err := cmnd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(rslt) + err.Error())
	}
	return string(rslt), err
}

// create a kubectl command context
func (client *KubeCtlClient) CreateKubectlCommand(ctx context.Context, args ...interface{}) *exec.Cmd {
	return exec.CommandContext(ctx, client.GetExecutable(), client.ComposeKubectlCommand(args...)...)
}

func (client *KubeCtlClient) ComposeKubectlCommand(args ...interface{}) []string {
	command := []string{}
	hasContext := false
	for _, ar := range args {
		switch ar.(type) {
		case string:
			if len(ar.(string)) == 0 {
				continue
			}
			if ar.(string) == "--context" {
				hasContext = true
			}
			command = append(command, ar.(string))
			break
		case []string:
			for _, part := range ar.([]string) {
				if len(part) == 0 {
					continue
				}
				command = append(command, part)
			}
			break
		default:
			continue
		}
	}

	if len(client.CoreArgs.Context) > 0 && !hasContext {
		// adding the context
		command = append([]string{"--context", client.CoreArgs.Context}, command...)
	}

	return command
}

// Loads default configuration for the client.
func (client *KubeCtlClient) LoadDefaults(ctx context.Context) error {
	if len(client.CoreArgs.Namespace) == 0 {
		namespace, err := client.RunKubectlCommand(
			ctx, "config", "view", "--minify", "--output",
			"jsonpath={..namespace}",
		)
		if err != nil {
			return err
		}
		client.CoreArgs.Namespace = strings.TrimSpace(namespace)
	}

	if len(client.CoreArgs.Context) == 0 {
		context, err := client.RunKubectlCommand(
			ctx, "config", "current-context",
		)
		if err != nil {
			return err
		}
		client.CoreArgs.Context = strings.TrimSpace(context)
	}

	return nil
}

// Get resource names from selector (with kind)
func (client *KubeCtlClient) GetResourceNames(
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

func (client *KubeCtlClient) DeployKubectlYaml(
	ctx context.Context,
	command, yaml string,
	wait bool,
) (string, error) {
	yamlFile, err := ioutil.TempFile(os.TempDir(), "wp.setup.kubectl.*.bat")
	if err != nil {
		return "", err
	}

	_, err = yamlFile.WriteString(yaml)
	if err != nil {
		return "", err
	}
	err = yamlFile.Close()
	if err != nil {
		return "", err
	}

	yamlFilename := yamlFile.Name()
	output, err := client.RunKubectlCommand(
		ctx,
		command,
		Triary(command == "delete", "--ignore-not-found=true", ""),
		Triary(wait, "--wait=true", "--wait=false"),
		"-f", yamlFilename,
	)
	removeErr := os.Remove(yamlFilename)

	if err != nil {
		return "", err
	}

	if removeErr != nil {
		return "", removeErr
	}

	return output, err
}

func (client *KubeCtlClient) WaitForConditions(
	ctx context.Context,
	resource string, conditions []string,
	count int,
) (string, error) {
	waitCommand := client.ComposeKubectlCommand(
		"wait",
		client.CoreArgs.ToArgsList(),
		"--timeout", fmt.Sprint(60*60*24*7)+"s",
		resource,
	)

	waitContext, cancel := context.WithCancel(ctx)
	completed := false

	var foundCondition string
	var waitError error

	for _, condition := range conditions {
		cmnd := client.CreateKubectlCommand(
			waitContext,
			waitCommand,
			"--for",
			"condition="+condition,
		)

		go func(condition string) {
			out, err := cmnd.CombinedOutput()
			if completed {
				return
			}
			completed = true
			foundCondition = condition
			if err != nil {
				waitError = errors.New(string(out) + "\n" + err.Error())
			}
			cancel()
		}(condition)
	}

	<-waitContext.Done()
	cancel()

	if waitError != nil {
		return foundCondition, waitError
	}
	return foundCondition, nil
}

func (client *KubeCtlClient) WaitForResourceEvents(
	ctx context.Context,
	resourceNameRegex string,
	matchEventNames []string,
	count int,
) (context.Context, error) {
	splitBySpaces, err := regexp.Compile(`\s+`)
	if err != nil {
		return nil, err
	}
	resourceRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		return nil, err
	}

	eventsContext, eventsContextCancel := context.WithCancel(ctx)

	eventsCmnd := client.CreateKubectlCommand(
		eventsContext,
		"get", "events", "--watch-only", "-o",
		`custom-columns=":involvedObject.name,:reason"`,
	)

	pr, err := eventsCmnd.StdoutPipe()
	if err != nil {
		eventsContextCancel()
		return eventsContext, err
	}

	lineScanner := bufio.NewScanner(pr)
	err = eventsCmnd.Start()

	if err != nil {
		eventsContextCancel()
		return eventsContext, err
	}

	matchLine := func(line string) bool {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			return false
		}

		eventArgs := splitBySpaces.Split(line, -1)
		if len(eventArgs) < 2 {
			return false
		}

		resource := eventArgs[0]
		eventName := eventArgs[1]

		if !resourceRegex.Match([]byte(resource)) {
			return false
		}

		matchedName := false

		for _, name := range matchEventNames {
			if name != eventName {
				matchedName = true
			}
		}

		return matchedName
	}

	go func() {
		matched := 0
		for lineScanner.Scan() {
			if matchLine(lineScanner.Text()) {
				matched++
			}

			if matched >= count {
				break
			}

			matched++
			if matched >= count {
				break
			}
		}

		eventsContextCancel()
	}()

	return eventsContext, nil
}

package kubectl

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func (backend *KubeBackend) MakeLogger(jobId string) zerolog.Logger {
	logger := log.With().Str("RunID", backend.RunID).Logger()
	if len(jobId) > 0 {
		logger = logger.With().Str("JobID", jobId).Logger()
	}
	return logger
}

// Initializes the configuration for the kube backend
// and populates the basic parameters for that config.
func (backend *KubeBackend) InitializeConfig(cfg *types.Config) error {
	backend.Config = cfg

	// resetting
	backend.SetupTemplates = []KubeTemplate{}

	// add network policy
	if backend.EnableRunNetworkPolicy {
		backend.SetupTemplates = append(backend.SetupTemplates, &KubeNetworkPolicyTemplate{
			Backend: backend,
		})
	}

	backend.PVCs = []*KubePVCTemplate{}
	backend.PVCByName = make(map[string]*KubePVCTemplate)

	for _, vol := range backend.Config.Volumes {
		pvc := &KubePVCTemplate{
			Backend: backend,
			Name:    vol.Name,
		}
		backend.PVCs = append(backend.PVCs, pvc)
		backend.PVCByName[vol.Name] = pvc
		backend.SetupTemplates = append(backend.SetupTemplates, pvc)
	}
	return nil
}

func (backend *KubeBackend) RenderSetupYaml() (string, error) {
	var templatesAsYaml []string

	for _, template := range backend.SetupTemplates {
		asYaml, err := template.Render()
		if err != nil {
			return "", err
		}
		templatesAsYaml = append(templatesAsYaml, asYaml)
	}

	return strings.Join(templatesAsYaml, "\n---\n"), nil
}

func (backend *KubeBackend) GetJobPodName(
	ctx context.Context,
	jobTemplate *KubeJobTemplate,
) ([]string, error) {
	var podNames []string
	var err error
	for {
		podNames, err = backend.Client.GetResourceNames(
			ctx,
			"pod",
			"woodpecker-job-id="+jobTemplate.JobID(),
		)

		if err != nil {
			return []string{}, err
		}

		if len(podNames) == 0 {
			continue
		}

		break
	}
	return podNames, nil
}

func (backend *KubeBackend) PopulateDetachedInfo(
	ctx context.Context,
	podName string,
	jobTemplate *KubeJobTemplate,
) error {
	podIP, err := backend.Client.RunKubectlCommand(
		ctx, "get", podName,
		"-o",
		"custom-columns=:status.podIP",
	)
	if err != nil {
		return err
	}
	jobTemplate.DetachedPodIP = strings.TrimSpace(podIP)
	return nil
}

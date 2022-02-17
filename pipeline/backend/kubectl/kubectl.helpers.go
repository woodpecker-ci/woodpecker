package kubectl

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func (backend *KubeCtlBackend) MakeLogger(jobId string) zerolog.Logger {
	logger := log.With().Str("RunID", backend.RunID).Logger()
	if len(jobId) > 0 {
		logger = logger.With().Str("JobID", jobId).Logger()
	}
	return logger
}

// Initializes the configuration for the kube backend
// and populates the basic parameters for that config.
func (backend *KubeCtlBackend) InitializeConfig(cfg *types.Config) {
	backend.Config = cfg

	// resetting
	backend.SetupTemplates = []KubeTemplate{}
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
}

func (backend *KubeCtlBackend) RenderSetupYaml() (string, error) {
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

package nomad

import (
	"context"

	nomadApi "github.com/hashicorp/nomad/api"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func generateNomadJob(ctx context.Context, step *types.Step) nomadApi.Job {
	region := "global"
	env, entry, cmd := common.GenerateContainerConf(step.Commands)
	containerConf := map[string]interface{}{}
	// If we're on the clone step then don't override it
	if step.Alias != "clone" {
		containerConf = map[string]interface{}{
			"image":      step.Image,
			"entrypoint": entry,
			"args":       cmd,
			"work_dir":   step.WorkingDir,
		}
	} else {
		containerConf = map[string]interface{}{
			"image": step.Image,
		}
	}
	for k, v := range env {
		step.Environment[k] = v
	}
	// Mount specified volumes
	// Will this work with host volumes?
	ro := false
	volName := "ci"
	private := "private"
	// Should be set based off of the retry policy in the yml file
	attempts := 0
	mounts := []*nomadApi.VolumeMount{}
	for _, v := range step.Volumes {
		dest := common.VolumeMountPath(v)
		vm := nomadApi.VolumeMount{
			Volume:          &volName,
			Destination:     &dest,
			ReadOnly:        &ro,
			PropagationMode: &private,
		}
		mounts = append(mounts, &vm)
	}
	tasks := []*nomadApi.Task{
		{
			Name:         step.Name,
			Driver:       "docker",
			Env:          step.Environment,
			Config:       containerConf,
			VolumeMounts: mounts,
			RestartPolicy: &nomadApi.RestartPolicy{
				Attempts: &attempts,
			},

			// TODO: Register a nomad service
			//Services:  []*nomadApi.Service{},
			//Resources: &nomadApi.Resources{
			//CPU:      &cpu,
			//Cores:    &shares,
			//MemoryMB: &mem,
			//},
		},
	}

	// Why is nomad launching two allocs when the first one fails?
	count := 1
	tgs := []*nomadApi.TaskGroup{
		{
			Name:  &step.Name,
			Count: &count,
			Tasks: tasks,
			RestartPolicy: &nomadApi.RestartPolicy{
				Attempts: &attempts,
			},
			// TODO: Loop over and create with unique names then feed these names into the task volume mount
			Volumes: map[string]*nomadApi.VolumeRequest{
				"ci": {
					Name:     volName,
					Type:     "host",
					Source:   volName,
					ReadOnly: ro,
					PerAlloc: false,
				},
			},
		},
	}

	ns := "default"
	dcs := []string{"dc1"}
	b := "batch"
	nj := nomadApi.Job{
		ID:          &step.Name,
		Region:      &region,
		Namespace:   &ns,
		Name:        &step.Name,
		Type:        &b,
		Datacenters: dcs,
		TaskGroups:  tgs,
		Reschedule: &nomadApi.ReschedulePolicy{
			Attempts: &attempts,
		},
		// ConsulToken: &consulToken.SecretID,
		// VaultToken:  &vaultToken,
	}

	return nj
}

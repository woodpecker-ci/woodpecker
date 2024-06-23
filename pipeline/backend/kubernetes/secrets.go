// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type nativeSecretsProcessor struct {
	config         *config
	secrets        []SecretRef
	envFromSources []v1.EnvFromSource
	envVars        []v1.EnvVar
	volumes        []v1.Volume
	mounts         []v1.VolumeMount
}

func newNativeSecretsProcessor(config *config, secrets []SecretRef) nativeSecretsProcessor {
	return nativeSecretsProcessor{
		config:  config,
		secrets: secrets,
	}
}

func (nsp *nativeSecretsProcessor) isEnabled() bool {
	return nsp.config.NativeSecretsAllowFromStep
}

func (nsp *nativeSecretsProcessor) process() error {
	if len(nsp.secrets) > 0 {
		if !nsp.isEnabled() {
			log.Debug().Msg("Secret names were defined in backend options, but secret access is disallowed by instance configuration.")
			return nil
		}
	} else {
		return nil
	}

	for _, secret := range nsp.secrets {
		switch {
		case secret.isSimple():
			simpleSecret, err := secret.toEnvFromSource()
			if err != nil {
				return err
			}
			nsp.envFromSources = append(nsp.envFromSources, simpleSecret)
		case secret.isAdvanced():
			advancedSecret, err := secret.toEnvVar()
			if err != nil {
				return err
			}
			nsp.envVars = append(nsp.envVars, advancedSecret)
		case secret.isFile():
			volume, err := secret.toVolume()
			if err != nil {
				return err
			}
			nsp.volumes = append(nsp.volumes, volume)

			mount, err := secret.toVolumeMount()
			if err != nil {
				return err
			}
			nsp.mounts = append(nsp.mounts, mount)
		}
	}

	return nil
}

func (sr SecretRef) isSimple() bool {
	return len(sr.Key) == 0 && len(sr.Target.Env) == 0 && !sr.isFile()
}

func (sr SecretRef) isAdvanced() bool {
	return (len(sr.Key) > 0 || len(sr.Target.Env) > 0) && !sr.isFile()
}

func (sr SecretRef) isFile() bool {
	return len(sr.Target.File) > 0
}

func (sr SecretRef) toEnvFromSource() (v1.EnvFromSource, error) {
	env := v1.EnvFromSource{}

	if !sr.isSimple() {
		return env, fmt.Errorf("secret '%s' is not simple reference", sr.Name)
	}

	env = v1.EnvFromSource{
		SecretRef: &v1.SecretEnvSource{
			LocalObjectReference: secretReference(sr.Name),
		},
	}

	return env, nil
}

func (sr SecretRef) toEnvVar() (v1.EnvVar, error) {
	envVar := v1.EnvVar{}

	if !sr.isAdvanced() {
		return envVar, fmt.Errorf("secret '%s' is not advanced reference", sr.Name)
	}

	envVar.ValueFrom = &v1.EnvVarSource{
		SecretKeyRef: &v1.SecretKeySelector{
			LocalObjectReference: secretReference(sr.Name),
			Key:                  sr.Key,
		},
	}

	if len(sr.Target.Env) > 0 {
		envVar.Name = sr.Target.Env
	} else {
		envVar.Name = strings.ToUpper(sr.Key)
	}

	return envVar, nil
}

func (sr SecretRef) toVolume() (v1.Volume, error) {
	var err error
	volume := v1.Volume{}

	if !sr.isFile() {
		return volume, fmt.Errorf("secret '%s' is not file reference", sr.Name)
	}

	volume.Name, err = volumeName(sr.Name)
	if err != nil {
		return volume, err
	}

	volume.Secret = &v1.SecretVolumeSource{
		SecretName: sr.Name,
	}

	return volume, nil
}

func (sr SecretRef) toVolumeMount() (v1.VolumeMount, error) {
	var err error
	mount := v1.VolumeMount{
		ReadOnly: true,
	}

	if !sr.isFile() {
		return mount, fmt.Errorf("secret '%s' is not file reference", sr.Name)
	}

	mount.Name, err = volumeName(sr.Name)
	if err != nil {
		return mount, err
	}

	mount.MountPath = sr.Target.File
	mount.SubPath = sr.Key

	return mount, nil
}

func secretsReferences(names []string) []v1.LocalObjectReference {
	secretReferences := make([]v1.LocalObjectReference, len(names))
	for i, imagePullSecretName := range names {
		secretReferences[i] = secretReference(imagePullSecretName)
	}
	return secretReferences
}

func secretReference(name string) v1.LocalObjectReference {
	return v1.LocalObjectReference{
		Name: name,
	}
}

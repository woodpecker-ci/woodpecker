package kubernetes

import (
	"strings"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend"
	// "github.com/vincent-petithory/dataurl"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(namespace string, step *backend.Step) *v1.Pod {

	var vols []v1.Volume
	var volMounts []v1.VolumeMount
	if step.WorkingDir != "" {
		for _, vol := range step.Volumes {
			vols = append(vols, v1.Volume{
				Name: volumeName(vol),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: volumeName(vol),
						ReadOnly:  false,
					},
				},
			})

			volMounts = append(volMounts, v1.VolumeMount{
				Name:      volumeName(vol),
				MountPath: volumeMountPath(step.WorkingDir),
				//MountPath: volumeMountPath(vol.Target),
			})
		}
	}

	pullPolicy := v1.PullIfNotPresent
	if step.Pull {
		pullPolicy = v1.PullAlways
	}

	command := step.Entrypoint
	args := step.Command
	envs := mapToEnvVars(step.Environment)

	if _, hasScript := step.Environment["CI_SCRIPT"]; !strings.HasSuffix(step.Name, "_clone") && hasScript {
		command = []string{"/bin/sh", "-c"}
		args = []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}
	}

	hostAliases := []v1.HostAlias{}
	for _, extraHost := range step.ExtraHosts {
		host := strings.Split(extraHost, ":")
		hostAliases = append(hostAliases, v1.HostAlias{IP: host[1], Hostnames: []string{host[0]}})
	}

	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("256Mi"),
			v1.ResourceCPU:    resource.MustParse("500m"),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("512Mi"),
			v1.ResourceCPU:    resource.MustParse("1000m"),
		},
	}

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName(step),
			Namespace: namespace,
			Labels: map[string]string{
				"step": podName(step),
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			HostAliases:   hostAliases,
			Containers: []v1.Container{{
				Name:            podName(step),
				Image:           step.Image,
				ImagePullPolicy: pullPolicy,
				Command:         command,
				Args:            args,
				WorkingDir:      step.WorkingDir,
				Env:             envs,
				VolumeMounts:    volMounts,
				Resources:       resources,
			}},
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "regcred"}},
			Volumes:          vols,
		},
	}
}

func podName(s *backend.Step) string {
	return dnsName(s.Name)
}

func mapToEnvVars(m map[string]string) []v1.EnvVar {
	var ev []v1.EnvVar
	for k, v := range m {
		ev = append(ev, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return ev
}

func volumeMountPath(i string) string {
	s := strings.Split(i, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}

// func decodeScript(snapshots []*engine.Snapshot) ([]byte, error) {
// 	var script bytes.Buffer

// 	for _, snapshot := range snapshots {
// 		if len(snapshot.Source) == 0 {
// 			continue
// 		}

// 		du, err := dataurl.DecodeString(string(snapshot.Source))
// 		if err != nil {
// 			return nil, nil
// 		}

// 		tr := tar.NewReader(bytes.NewReader(du.Data))
// 		for {
// 			hdr, err := tr.Next()
// 			if err == io.EOF {
// 				break // End of archive
// 			}
// 			if err != nil {
// 				return nil, err
// 			}

// 			if hdr.Name == "bin/_drone" {
// 				if _, err := io.Copy(&script, tr); err != nil {
// 					log.Println(err)
// 					return nil, nil
// 				}
// 			}
// 		}
// 	}

// 	return script.Bytes(), nil
// }

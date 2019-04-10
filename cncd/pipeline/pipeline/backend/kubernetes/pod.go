package kubernetes

import (
	"strings"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend"
	// "github.com/vincent-petithory/dataurl"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(namespace string, step *backend.Step) (*v1.Pod, error) {
	var vols []v1.Volume
	var volMounts []v1.VolumeMount
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

	pullPolicy := v1.PullIfNotPresent
	if step.Pull {
		pullPolicy = v1.PullAlways
	}

	command := step.Entrypoint
	args := step.Command
	envs := mapToEnvVars(step.Environment)

	if !strings.HasSuffix(step.Name, "_clone_0") {
		command = []string{"/bin/sh", "-c"}
		args = []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}
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
			Containers: []v1.Container{{
				Name:            podName(step),
				Image:           step.Image,
				ImagePullPolicy: pullPolicy,
				Command:         command,
				Args:            args,
				WorkingDir:      step.WorkingDir,
				Env:             envs,
				VolumeMounts:    volMounts,
			}},
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "regcred"}},
			Volumes:          vols,
		},
	}, nil
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

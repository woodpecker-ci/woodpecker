package kubectl

import (
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func composeValidTestKubeJobTemplate(
	backend *KubeCtlBackend,
) *KubeJobTemplate {
	return &KubeJobTemplate{
		Backend: backend,
		Step: &types.Step{
			// Only these values are supported/considered
			Name:  "lama",
			Image: "ubuntu:latest",
			Environment: map[string]string{
				"SPECIAL_CHARS": "!@##@$@$%^&%",
				"SOME_VALUE":    "a",
			},
			Pull:       true,
			Detached:   true,
			Privileged: true,
			Alias:      "a-tester", // valid for values
			WorkingDir: "/woodpecker",
			Labels: map[string]string{
				"label_a": "a",
				"label_b": "b",
			},
			Entrypoint: []string{"entrypoint"}, // should be ignored.
			Command:    []string{"echo ok", "echo ok2"},
			Volumes: []string{
				"default_volume:/woodpecker",
				"should_be_ignored:/woodpecker/src",
			},
			DNS: []string{
				"10.10.10.10",
				"10.10.10.11",
				"10.10.10.12",
			},
			DNSSearch: []string{
				"not-a-service.com",
			},
		},
	}
}

func TestTemplates(t *testing.T) {
	backend := prepareTestBackend()
	g := goblin.Goblin(t)

	g.Describe("Templates test:", func() {
		g.It("read a template file", func() {
			tmpl, err := Embedded.ReadFile("templates/step_job.yaml")
			g.Assert(err).Equal(nil)
			t.Log(string(tmpl))
		})

		g.It("render a job template (without values)", func() {
			tmpl := KubeJobTemplate{
				Backend: backend,
				Step: &types.Step{
					Name:  "lama",
					Image: "ubuntu:latest",
				},
			}
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})

		g.It("render a job template (with values)", func() {
			tmpl := composeValidTestKubeJobTemplate(backend)
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			g.Assert(strings.Contains(rslt, "entrypoint")).Equal(false)
			t.Log(rslt)
		})

		g.It("render a volume template (with values)", func() {
			tmpl := KubePVCTemplate{
				Backend:          backend,
				StorageClassName: "default",
				StorageSize:      "3Gi",
			}
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})

		g.It("render a volume template (missing values)", func() {
			tmpl := KubePVCTemplate{
				Backend: backend,
			}
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})
	})
}

package kubectl

import (
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func composeValidTestKubeJobTemplate(
	run *KubeBackendRun,
) *KubeJobTemplate {
	return &KubeJobTemplate{
		Run: run,
		Step: &types.Step{
			// Only these values are supported/considered
			Name:  "lama",
			Image: "ubuntu:latest",
			Environment: map[string]string{
				"SPECIAL_CHARS": "!@##@$@$%^&%",
				"SOME_VALUE":    "a",
			},
			Pull:       true,
			Detached:   false,
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
	run := prepareTestBackendRun()
	g := goblin.Goblin(t)

	g.Describe("Templates test:", func() {
		g.It("read a template file", func() {
			tmpl, err := Embedded.ReadFile("templates/step_job.yaml")
			g.Assert(err).Equal(nil)
			t.Log(string(tmpl))
		})

		g.It("render a job template (without values)", func() {
			tmpl := KubeJobTemplate{
				Run: run,
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
			tmpl := composeValidTestKubeJobTemplate(run)
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
				Run: run,
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
				Run: run,
			}
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})

		g.It("Render a network policy template", func() {
			tmpl := KubeNetworkPolicyTemplate{
				Run: run,
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

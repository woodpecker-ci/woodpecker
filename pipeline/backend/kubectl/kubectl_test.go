package kubectl

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func TestEngineCore(t *testing.T) {
	backend := New("kubectl", KubeCtlClientCoreArgs{}).(*KubeCtlBackend)
	g := goblin.Goblin(t)

	g.Describe("Engine core:", func() {
		g.It("get cluster info", func() {
			g.Assert(backend.Load()).Equal(nil)
		})

		g.It("read a template file", func() {
			tmpl, err := Embedded.ReadFile("templates/step_job.yaml")
			g.Assert(err).Equal(nil)
			t.Log(string(tmpl))
		})

		g.It("render a job template", func() {
			tmpl := KubeJobTemplate{
				Backend: backend,
				Step: &types.Step{
					Name:  "lama",
					Image: "artprod.dev.bloomberg.com/ubuntu20:latest",
					Environment: map[string]string{
						"test":   "asd+asd+a",
						"tester": "b",
					},
				},
			}
			rslt, err := tmpl.Render()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
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

		g.It("Render setup yaml", func() {
			rslt, err := backend.RenderSetupYaml()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})
	})
}

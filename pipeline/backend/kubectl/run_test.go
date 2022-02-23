package kubectl

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func prepareTestBackendRun() *KubePiplineRun {
	backend := New().(*KubeBackend)
	// reset a new run.
	run := backend.CreateRun()
	_ = run.InitializeConfig(&types.Config{
		Volumes: []*types.Volume{
			&(types.Volume{
				Name: "default_volume",
			}),
		},
	})

	return run
}

func TestEngineCore(t *testing.T) {
	run := prepareTestBackendRun()
	g := goblin.Goblin(t)

	g.Describe("Engine core:", func() {
		g.It("Render setup yaml", func() {
			rslt, err := run.RenderSetupYaml()
			if err != nil {
				t.Error(err)
			}
			g.Assert(err).Equal(nil)
			t.Log(rslt)
		})
	})
}

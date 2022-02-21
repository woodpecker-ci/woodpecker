package kubectl

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func prepareTestBackend() *KubeBackend {
	backend := New("kubectl", KubeCtlClientCoreArgs{}).(*KubeBackend)
	// reset a new run.
	backend.Reset()

	backend.InitializeConfig(&types.Config{
		Volumes: []*types.Volume{
			&(types.Volume{
				Name: "default_volume",
			}),
		},
	})

	return backend
}

func TestEngineCore(t *testing.T) {
	backend := prepareTestBackend()
	g := goblin.Goblin(t)

	g.Describe("Engine core:", func() {
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

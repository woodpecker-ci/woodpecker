package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/api"
	mocks_forge "go.woodpecker-ci.org/woodpecker/v2/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	mocks_manager "go.woodpecker-ci.org/woodpecker/v2/server/services/mocks"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func TestHandleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Login", func() {
		g.It("should redirect to forge login", func() {
			_manager := mocks_manager.NewManager(t)
			_forge := mocks_forge.NewForge(t)
			_store := mocks_store.NewStore(t)

			_forge.On("Login", mock.Anything, mock.Anything).Return(&model.User{}, "", nil)
			_manager.On("ForgeMain").Return(_forge, nil)
			server.Config.Services.Manager = _manager

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", _store)

			api.HandleAuth(c)

			// mockStore.AssertCalled(t, "GetPipelineList", mock.Anything, mock.Anything, mock.Anything)
			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})

		g.It("should handle the error", func() {
			_manager := mocks_manager.NewManager(t)
			_forge := mocks_forge.NewForge(t)
			_store := mocks_store.NewStore(t)

			_forge.On("Login", mock.Anything, mock.Anything).Return(&model.User{}, "", nil)
			_manager.On("ForgeMain").Return(_forge, nil)
			server.Config.Services.Manager = _manager

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", _store)
			c.Request.URL, _ = url.Parse("/authorize?error=access_denied")

			api.HandleAuth(c)

			// check if we get redirected to /login?error=access_denied
			assert.Equal(t, http.StatusFound, c.Writer.Status())
			assert.Equal(t, "/login?error=access_denied", c.Writer.Header().Get("Location"))
		})
	})
}

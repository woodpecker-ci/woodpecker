package api_test

import (
	"fmt"
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
	mocks_services "go.woodpecker-ci.org/woodpecker/v2/server/services/mocks"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func TestHandleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Login", func() {
		g.It("should handle errors from the callback", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			query := url.Values{}
			query.Set("error", "invalid_scope")
			query.Set("error_description", "The requested scope is invalid, unknown, or malformed")
			query.Set("error_uri", "https://developer.atlassian.com/cloud/jira/platform/rest/#api-group-OAuth2-ErrorHandling")

			c.Request = &http.Request{
				Header: make(http.Header),
				Method: http.MethodGet,
				URL: &url.URL{
					Path:     "/authorize",
					RawQuery: query.Encode(),
				},
			}

			api.HandleAuth(c)

			assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
			assert.Equal(t, fmt.Sprintf("/login?%s", query.Encode()), c.Writer.Header().Get("Location"))
		})

		g.It("should fail if a code was provided, but no state", func() {
			// TODO: implement
		})

		g.It("should fail if a code was provided, but the state is wrong", func() {
			// TODO: implement
		})

		g.It("should fail if a code was provided, but the state is wrong", func() {
			// TODO: implement
		})

		g.It("should redirect to forge login page", func() {
			_manager := mocks_services.NewManager(t)
			_forge := mocks_forge.NewForge(t)
			_store := mocks_store.NewStore(t)

			forgeRedirectURL := "https://my-awesome-forge.com/oauth/authorize?client_id=client-id"

			_forge.On("Login", mock.Anything, mock.Anything).Return(nil, forgeRedirectURL, nil)
			_manager.On("ForgeMain").Return(_forge, nil)
			server.Config.Services.Manager = _manager

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", _store)
			c.Request = &http.Request{
				Header: make(http.Header),
			}

			api.HandleAuth(c)

			assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
			assert.Equal(t, forgeRedirectURL, c.Writer.Header().Get("Location"))
		})

		g.It("should handle the callback and register a new user", func() {
			// TODO: implement
		})

		g.It("should handle the callback and login an existing user", func() {
			// TODO: implement
		})

		g.It("should handle the callback and deny a new user to register", func() {
			// TODO: implement
		})
	})
}

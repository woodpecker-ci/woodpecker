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
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	mocks_config_service "go.woodpecker-ci.org/woodpecker/v2/server/services/config/mocks"
	mocks_services "go.woodpecker-ci.org/woodpecker/v2/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/permissions"
	mocks_registry_service "go.woodpecker-ci.org/woodpecker/v2/server/services/registry/mocks"
	mocks_secret_service "go.woodpecker-ci.org/woodpecker/v2/server/services/secret/mocks"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

func TestHook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Hook", func() {
		g.It("should handle a correct webhook payload", func() {
			_manager := mocks_services.NewManager(t)
			_forge := mocks_forge.NewForge(t)
			_store := mocks_store.NewStore(t)
			_configService := mocks_config_service.NewService(t)
			_secretService := mocks_secret_service.NewService(t)
			_registryService := mocks_registry_service.NewService(t)
			server.Config.Services.Manager = _manager
			server.Config.Permissions.Open = true
			server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
			server.Config.Permissions.Admins = permissions.NewAdmins(nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", _store)
			user := &model.User{
				ID: 123,
			}
			repo := &model.Repo{
				ID:            123,
				ForgeRemoteID: "123",
				Owner:         "owner",
				Name:          "name",
				IsActive:      true,
				UserID:        user.ID,
				Hash:          "secret-123-this-is-a-secret",
			}
			pipeline := &model.Pipeline{
				ID:     123,
				RepoID: repo.ID,
				Event:  model.EventPush,
			}

			repoToken := token.New(token.HookToken)
			repoToken.Set("repo-id", fmt.Sprintf("%d", repo.ID))
			signedToken, err := repoToken.Sign("secret-123-this-is-a-secret")
			if err != nil {
				g.Fail(err)
			}

			header := http.Header{}
			header.Set("Authorization", fmt.Sprintf("Bearer %s", signedToken))
			c.Request = &http.Request{
				Header: header,
				URL: &url.URL{
					Scheme: "https",
				},
			}

			_manager.On("ForgeFromRepo", repo).Return(_forge, nil)
			_forge.On("Hook", mock.Anything, mock.Anything).Return(repo, pipeline, nil)
			_store.On("GetRepo", repo.ID).Return(repo, nil)
			_store.On("GetUser", user.ID).Return(user, nil)
			_store.On("UpdateRepo", repo).Return(nil)
			_store.On("CreatePipeline", mock.Anything).Return(nil)
			_manager.On("ConfigServiceFromRepo", repo).Return(_configService)
			_configService.On("Fetch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			_forge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{}, nil)
			_store.On("GetPipelineLastBefore", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			_manager.On("SecretServiceFromRepo", repo).Return(_secretService)
			_secretService.On("SecretListPipeline", repo, mock.Anything, mock.Anything).Return(nil, nil)
			_manager.On("RegistryServiceFromRepo", repo).Return(_registryService)
			_registryService.On("RegistryListPipeline", repo, mock.Anything).Return(nil, nil)
			_manager.On("EnvironmentService").Return(nil)
			_store.On("DeletePipeline", mock.Anything).Return(nil)

			api.PostHook(c)

			assert.Equal(g, http.StatusNoContent, c.Writer.Status())
			assert.Equal(g, "true", w.Header().Get("Pipeline-Filtered"))
		})
	})
}

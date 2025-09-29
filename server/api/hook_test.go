package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/api"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	config_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/config/mocks"
	services_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
)

func TestHook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_manager := services_mocks.NewMockManager(t)
	_forge := forge_mocks.NewMockForge(t)
	_store := store_mocks.NewMockStore(t)
	_configService := config_service_mocks.NewMockService(t)
	_secretService := secret_service_mocks.NewMockService(t)
	_registryService := registry_service_mocks.NewMockService(t)
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
	assert.NoError(t, err)

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

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, "true", w.Header().Get("Pipeline-Filtered"))
}

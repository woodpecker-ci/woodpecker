package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/api"
	mocks_forge "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	mocks_services "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
)

func TestHandleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user := &model.User{
		ID:            1,
		OrgID:         1,
		ForgeID:       1,
		ForgeRemoteID: "remote-id-1",
		Login:         "test",
		Email:         "test@example.com",
		Admin:         false,
	}
	org := &model.Org{
		ID:   1,
		Name: user.Login,
	}

	server.Config.Server.SessionExpires = time.Hour

	t.Run("should handle errors from the callback", func(t *testing.T) {
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
				Scheme:   "https",
				Path:     "/authorize",
				RawQuery: query.Encode(),
			},
		}

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, fmt.Sprintf("/login?%s", query.Encode()), c.Writer.Header().Get("Location"))
	})

	t.Run("should fail if the state is wrong", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)

		query := url.Values{}
		query.Set("code", "assumed_to_be_valid_code")

		wrongToken := token.New(token.OAuthStateToken)
		wrongToken.Set("forge_id", "1")
		signedWrongToken, _ := wrongToken.Sign("wrong_secret")
		query.Set("state", signedWrongToken)

		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme:   "https",
				RawQuery: query.Encode(),
			},
		}

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=invalid_state", c.Writer.Header().Get("Location"))
	})

	t.Run("should redirect to forge login page", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)

		forgeRedirectURL := ""
		_forge.On("Login", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			state, ok := args.Get(1).(*forge_types.OAuthRequest)
			if ok {
				forgeRedirectURL = fmt.Sprintf("https://my-awesome-forge.com/oauth/authorize?client_id=client-id&state=%s", state.State)
			}
		}).Return(nil, func(context.Context, *forge_types.OAuthRequest) string {
			return forgeRedirectURL
		}, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, forgeRedirectURL, c.Writer.Header().Get("Location"))
	})

	t.Run("should register a new user", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(nil, types.RecordNotExist)
		_store.On("CreateUser", mock.Anything).Return(nil)
		_store.On("OrgFindByName", user.Login).Return(nil, nil)
		_store.On("OrgCreate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_forge.On("Repos", mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should login an existing user", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(user, nil)
		_store.On("OrgGet", org.ID).Return(org, nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_forge.On("Repos", mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should deny a new user if registration is closed", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = false
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(nil, types.RecordNotExist)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=registration_closed", c.Writer.Header().Get("Location"))
	})

	t.Run("should deny a user with missing org access", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs([]string{"org1"})
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_forge.On("Teams", mock.Anything, user).Return([]*model.Team{
			{
				Login: "org2",
			},
		}, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=org_access_denied", c.Writer.Header().Get("Location"))
	})

	t.Run("should create an user org if it does not exists", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		user.OrgID = 0

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(user, nil)
		_store.On("OrgFindByName", user.Login).Return(nil, types.RecordNotExist)
		_store.On("OrgCreate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_forge.On("Repos", mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should link an user org if it has the same name as the user", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		user.OrgID = 0

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(user, nil)
		_store.On("OrgFindByName", user.Login).Return(org, nil)
		_store.On("OrgUpdate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_forge.On("Repos", mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should update an user org if the user name was changed", func(t *testing.T) {
		_manager := mocks_services.NewManager(t)
		_forge := mocks_forge.NewForge(t)
		_store := mocks_store.NewStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		org.Name = "not-the-user-name"

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserRemoteID", user.ForgeRemoteID, user.Login).Return(user, nil)
		_store.On("OrgGet", user.OrgID).Return(org, nil)
		_store.On("OrgUpdate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_forge.On("Repos", mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})
}

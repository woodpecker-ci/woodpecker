package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// GetParameter returns a repository parameter by ID.
func GetParameter(c *gin.Context) {
	repo := session.Repo(c)
	paramID, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid parameter ID")
		return
	}

	parameterService := server.Config.Services.Manager.ParameterServiceFromRepo(repo)
	parameter, err := parameterService.ParameterFindByID(repo, paramID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, parameter)
}

// PostParameter persists a parameter.
func PostParameter(c *gin.Context) {
	repo := session.Repo(c)

	in := new(model.Parameter)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing parameter. %s", err)
		return
	}
	in.RepoID = repo.ID

	if err := in.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error validating parameter. %s", err)
		return
	}

	parameterService := server.Config.Services.Manager.ParameterServiceFromRepo(repo)

	// Check if parameter with same name and branch already exists
	existing, err := parameterService.ParameterFindByNameAndBranch(repo, in.Name, in.Branch)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		handleDBError(c, err)
		return
	}
	if existing != nil && existing.ID != 0 {
		c.String(http.StatusConflict, "Parameter with name '%s' already exists for branch '%s': existing: %d, new: %d", in.Name, in.Branch, existing.ID, in.ID)
		return
	}

	err = parameterService.ParameterCreate(repo, in)
	if err != nil {
		handleDBError(c, err)
		return
	}
	parameter, err := parameterService.ParameterFind(repo, in.Name)
	c.JSON(http.StatusOK, parameter)
}

// PatchParameter updates an existing parameter by ID
func PatchParameter(c *gin.Context) {
	repo := session.Repo(c)
	paramID, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid parameter ID")
		return
	}

	in := new(model.Parameter)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing parameter. %s", err)
		return
	}
	in.RepoID = repo.ID
	in.ID = paramID

	if err := in.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error validating parameter. %s", err)
		return
	}

	parameterService := server.Config.Services.Manager.ParameterServiceFromRepo(repo)

	// Get existing parameter to check if name/branch changed
	existing, err := parameterService.ParameterFindByID(repo, paramID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// If name or branch changed, check for conflicts
	if existing.Name != in.Name || existing.Branch != in.Branch {
		conflict, err := parameterService.ParameterFindByNameAndBranch(repo, in.Name, in.Branch)
		if err != nil && !errors.Is(err, types.RecordNotExist) {
			handleDBError(c, err)
			return
		}
		if conflict != nil && conflict.ID != 0 && conflict.ID != paramID {
			c.String(http.StatusConflict, "Parameter with name '%s' already exists for branch '%s'", in.Name, in.Branch)
			return
		}
	}

	err = parameterService.ParameterUpdate(repo, in)
	if err != nil {
		handleDBError(c, err)
		return
	}
	parameter, err := parameterService.ParameterFindByID(repo, paramID)
	c.JSON(http.StatusOK, parameter)
}

// GetParameterList returns all repository parameters.
func GetParameterList(c *gin.Context) {
	repo := session.Repo(c)
	parameterService := server.Config.Services.Manager.ParameterServiceFromRepo(repo)
	list, err := parameterService.ParameterList(repo)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// DeleteParameter deletes a parameter by ID.
func DeleteParameter(c *gin.Context) {
	repo := session.Repo(c)
	paramID, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid parameter ID")
		return
	}

	parameterService := server.Config.Services.Manager.ParameterServiceFromRepo(repo)
	if err := parameterService.ParameterDeleteByID(repo, paramID); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

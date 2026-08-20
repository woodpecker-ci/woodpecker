// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// getOrCreateToken returns the token of the given type of a repo, creating it
// if the repo does not hold one yet. Repos get their tokens on activation, this
// keeps repos activated by an older version working as well.
func getOrCreateToken(s store.Store, repo *model.Repo, tokenType model.TokenType) (*model.Token, error) {
	token, err := s.TokenFind(repo, tokenType)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, types.ErrRecordNotExist) {
		return nil, err
	}

	token = model.NewToken(repo.ID, tokenType)
	if err := s.TokenCreate(token); err != nil {
		// a concurrent request won the race, its token is the valid one
		if errors.Is(err, types.ErrInsertDuplicateDetected) {
			return s.TokenFind(repo, tokenType)
		}
		return nil, err
	}

	return token, nil
}

// GetRepoToken
//
//	@Summary	Get a token of the repository
//	@Router		/repos/{repo_id}/token [get]
//	@Produce	json
//	@Success	200	{object}	Token
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		type			query	string	false	"the type of the token"	default(badge)
func GetRepoToken(c *gin.Context) {
	repo := session.Repo(c)

	tokenType := model.TokenType(c.DefaultQuery("type", string(model.TokenTypeBadge)))
	if !tokenType.Valid() {
		c.String(http.StatusBadRequest, "Invalid token type '%s'", tokenType)
		return
	}

	token, err := getOrCreateToken(store.FromContext(c), repo, tokenType)
	if err != nil {
		handleDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, token)
}

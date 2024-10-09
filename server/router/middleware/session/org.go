// Copyright 2024 Woodpecker Authors
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

package session

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func Org(c *gin.Context) *model.Org {
	v, ok := c.Get("org")
	if !ok {
		return nil
	}
	r, ok := v.(*model.Org)
	if !ok {
		return nil
	}
	return r
}

func SetOrg() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			orgID int64
			err   error
		)

		orgParam := c.Param("org_id")
		if orgParam != "" {
			orgID, err = strconv.ParseInt(orgParam, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Invalid organization ID")
				c.Abort()
				return
			}
		}

		org, err := store.FromContext(c).OrgGet(orgID)
		if err != nil && !errors.Is(err, types.RecordNotExist) {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if org == nil {
			c.String(http.StatusNotFound, "Organization not found")
			c.Abort()
			return
		}

		c.Set("org", org)
		c.Next()
	}
}

func MustOrg() gin.HandlerFunc {
	return func(c *gin.Context) {
		org := Org(c)
		switch {
		case org == nil:
			c.String(http.StatusNotFound, "Organization not loaded")
			c.Abort()
		default:
			c.Next()
		}
	}
}

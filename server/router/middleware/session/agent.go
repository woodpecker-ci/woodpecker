// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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
	"net/http"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

// AuthorizeAgent authorizes requests from agent to access the queue.
func AuthorizeAgent(c *gin.Context) {
	secret, _ := c.MustGet("agent").(string)
	if secret == "" {
		c.String(http.StatusUnauthorized, "invalid or empty token.")
		return
	}

	_, err := token.ParseRequest([]token.Type{token.AgentToken}, c.Request, func(_ *token.Token) (string, error) {
		return secret, nil
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "invalid or empty token. %s", err)
		c.Abort()
		return
	}

	c.Next()
}

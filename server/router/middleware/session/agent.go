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

	"github.com/woodpecker-ci/woodpecker/shared/token"
)

// AuthorizeAgent authorizes requests from build agents to access the queue.
func AuthorizeAgent(c *gin.Context) {
	secret, _ := c.MustGet("agent").(string)
	if secret == "" {
		c.String(http.StatusUnauthorized, "invalid or empty token.")
		return
	}

	parsed, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
		return secret, nil
	})
	switch {
	case err != nil:
		c.String(http.StatusInternalServerError, "invalid or empty token. %s", err)
		c.Abort()
	case parsed.Kind != token.AgentToken:
		c.String(http.StatusForbidden, "invalid token. please use an agent token")
		c.Abort()
	default:
		c.Next()
	}
}

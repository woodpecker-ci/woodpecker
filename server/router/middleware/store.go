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

package middleware

import (
	"github.com/urfave/cli"
	"github.com/woodpecker-ci/woodpecker/store"

	"github.com/gin-gonic/gin"
)

// Store is a middleware function that initializes the Datastore and attaches to
// the context of every http.Request.
func Store(cli *cli.Context, v store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		store.ToContext(c, v)
		c.Next()
	}
}

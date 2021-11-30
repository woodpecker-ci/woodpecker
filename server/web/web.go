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

package web

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/web"
)

// etag is an identifier for a resource version
// it lets caches determine if resource is still the same and not send it again
var etag = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))

// New returns a gin engine to serve the web frontend.
func New() *gin.Engine {
	e := gin.New()

	e.Use(setupCache)

	h := http.FileServer(web.HttpFS())
	e.GET("/favicon.svg", gin.WrapH(h))
	e.GET("/assets/*filepath", gin.WrapH(h))

	e.NoRoute(handleIndex)

	return e
}

func handleIndex(c *gin.Context) {
	rw := c.Writer
	data := web.MustLookup("index.html")
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.WriteHeader(200)
	if _, err := rw.Write(data); err != nil {
		log.Error().Err(err).Msg("can not write index.html")
	}
}

func setupCache(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
	c.Writer.Header().Del("Expires")
	c.Writer.Header().Set("ETag", etag)
}

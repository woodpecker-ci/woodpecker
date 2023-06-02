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
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/web"
)

// etag is an identifier for a resource version
// it lets caches determine if resource is still the same and not send it again
var (
	etag      = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
	indexHTML []byte
)

type prefixFS struct {
	fs     http.FileSystem
	prefix string
}

func (f *prefixFS) Open(name string) (http.File, error) {
	return f.fs.Open(strings.TrimPrefix(name, f.prefix))
}

// New returns a gin engine to serve the web frontend.
func New() (*gin.Engine, error) {
	e := gin.New()
	indexHTML = parseIndex()

	e.Use(setupCache)

	rootURL, _ := url.Parse(server.Config.Server.RootPath)
	rootPath := rootURL.Path

	httpFS, err := web.HTTPFS()
	if err != nil {
		return nil, err
	}
	h := http.FileServer(&prefixFS{httpFS, rootPath})
	e.GET(rootPath+"/favicon.svg", redirect(server.Config.Server.RootPath+"/favicons/favicon-light-default.svg", http.StatusPermanentRedirect))
	e.GET(rootPath+"/favicons/*filepath", gin.WrapH(h))
	e.GET(rootPath+"/assets/*filepath", gin.WrapH(h))

	e.NoRoute(handleIndex)

	return e, nil
}

// redirect return gin helper to redirect a request
func redirect(location string, status ...int) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		code := http.StatusFound
		if len(status) == 1 {
			code = status[0]
		}

		http.Redirect(ctx.Writer, ctx.Request, location, code)
	}
}

func handleIndex(c *gin.Context) {
	rw := c.Writer
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(indexHTML); err != nil {
		log.Error().Err(err).Msg("can not write index.html")
	}
}

func parseIndex() []byte {
	data, err := web.Lookup("index.html")
	if err != nil {
		log.Fatal().Err(err).Msg("can not find index.html")
	}
	if server.Config.Server.RootPath == "" {
		return data
	}
	return regexp.MustCompile(`/\S+\.(js|css|png|svg)`).ReplaceAll(data, []byte(server.Config.Server.RootPath+"$0"))
}

func setupCache(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
	c.Writer.Header().Del("Expires")
	c.Writer.Header().Set("ETag", etag)
}

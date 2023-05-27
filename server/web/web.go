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
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"
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

// New returns a gin engine to serve the web frontend.
func New() (*gin.Engine, error) {
	e := gin.New()
	indexHTML = parseIndex()

	e.Use(setupCache)

	httpFS, err := web.HTTPFS()
	if err != nil {
		return nil, err
	}
	h := http.FileServer(httpFS)
	e.GET("/favicon.svg", redirect("/favicons/favicon-light-default.svg", http.StatusPermanentRedirect))
	e.GET("/favicons/*filepath", gin.WrapH(h))
	e.GET("/assets/*filepath", gin.WrapH(handleCustomFilesAndAssets(h)))

	e.NoRoute(handleIndex)

	return e, nil
}

func handleCustomFilesAndAssets(assetHandler http.Handler) http.HandlerFunc {
	var serveFileOrEmptyContent = func(w http.ResponseWriter, r *http.Request, localFileName string) {
		if len(localFileName) > 0 {
			http.ServeFile(w, r, localFileName)
		} else {
			// prefer zero content over sending a 404 Not Found
			http.ServeContent(w, r, localFileName, time.Now(), bytes.NewReader([]byte{}))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/assets/custom.js" {
			serveFileOrEmptyContent(w, r, server.Config.Server.CustomJsFile)
		} else if r.RequestURI == "/assets/custom.css" {
			serveFileOrEmptyContent(w, r, server.Config.Server.CustomCssFile)
		} else {
			assetHandler.ServeHTTP(w, r)
		}
	}
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
	if server.Config.Server.RootURL == "" {
		return data
	}
	return regexp.MustCompile(`/\S+\.(js|css|png|svg)`).ReplaceAll(data, []byte(server.Config.Server.RootURL+"$0"))
}

func setupCache(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
	c.Writer.Header().Del("Expires")
	c.Writer.Header().Set("ETag", etag)
}

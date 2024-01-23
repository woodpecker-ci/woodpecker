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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/web"
)

var indexHTML []byte

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
	var err error
	indexHTML, err = parseIndex()
	if err != nil {
		return nil, err
	}

	rootPath := server.Config.Server.RootPath

	httpFS, err := web.HTTPFS()
	if err != nil {
		return nil, err
	}
	f := &prefixFS{httpFS, rootPath}
	e.GET(rootPath+"/favicon.svg", redirect(server.Config.Server.RootPath+"/favicons/favicon-light-default.svg", http.StatusPermanentRedirect))
	e.GET(rootPath+"/favicons/*filepath", serveFile(f))
	e.GET(rootPath+"/assets/*filepath", handleCustomFilesAndAssets(f))

	e.NoRoute(handleIndex)

	return e, nil
}

func handleCustomFilesAndAssets(fs *prefixFS) func(ctx *gin.Context) {
	serveFileOrEmptyContent := func(w http.ResponseWriter, r *http.Request, localFileName, fileName string) {
		if len(localFileName) > 0 {
			http.ServeFile(w, r, localFileName)
		} else {
			// prefer zero content over sending a 404 Not Found
			http.ServeContent(w, r, fileName, time.Now(), bytes.NewReader([]byte{}))
		}
	}
	return func(ctx *gin.Context) {
		switch {
		case strings.HasSuffix(ctx.Request.RequestURI, "/assets/custom.js"):
			serveFileOrEmptyContent(ctx.Writer, ctx.Request, server.Config.Server.CustomJsFile, "file.js")
		case strings.HasSuffix(ctx.Request.RequestURI, "/assets/custom.css"):
			serveFileOrEmptyContent(ctx.Writer, ctx.Request, server.Config.Server.CustomCSSFile, "file.css")
		default:
			serveFile(fs)(ctx)
		}
	}
}

func serveFile(f *prefixFS) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		file, err := f.Open(ctx.Request.URL.Path)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, fs.ErrNotExist) {
				code = http.StatusNotFound
			} else if errors.Is(err, fs.ErrPermission) {
				code = http.StatusForbidden
			}
			ctx.Status(code)
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		var mime string
		switch {
		case strings.HasSuffix(ctx.Request.URL.Path, ".js"):
			mime = "text/javascript"
		case strings.HasSuffix(ctx.Request.URL.Path, ".css"):
			mime = "text/css"
		case strings.HasSuffix(ctx.Request.URL.Path, ".png"):
			mime = "image/png"
		case strings.HasSuffix(ctx.Request.URL.Path, ".svg"):
			mime = "image/svg+xml"
		}
		ctx.Status(http.StatusOK)
		ctx.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		ctx.Writer.Header().Del("Expires")
		ctx.Writer.Header().Set("Content-Type", mime)
		if _, err := ctx.Writer.Write(replaceBytes(data)); err != nil {
			log.Error().Err(err).Msgf("cannot write %s", ctx.Request.URL.Path)
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
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(indexHTML); err != nil {
		log.Error().Err(err).Msg("cannot write index.html")
	}
}

func loadFile(path string) ([]byte, error) {
	data, err := web.Lookup(path)
	if err != nil {
		return nil, err
	}
	return replaceBytes(data), nil
}

func replaceBytes(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("/BASE_PATH"), []byte(server.Config.Server.RootPath))
}

func parseIndex() ([]byte, error) {
	data, err := loadFile("index.html")
	if err != nil {
		return nil, fmt.Errorf("cannot find index.html: %w", err)
	}
	data = bytes.ReplaceAll(data, []byte("/web-config.js"), []byte(server.Config.Server.RootPath+"/web-config.js"))
	data = bytes.ReplaceAll(data, []byte("/assets/custom.css"), []byte(server.Config.Server.RootPath+"/assets/custom.css"))
	data = bytes.ReplaceAll(data, []byte("/assets/custom.js"), []byte(server.Config.Server.RootPath+"/assets/custom.js"))
	return data, nil
}

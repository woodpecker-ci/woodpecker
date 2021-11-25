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
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/web"
)

// Endpoint provides the website endpoints.
type Endpoint interface {
	// Register registers the provider endpoints.
	Register(*gin.Engine)
}

// New returns the default website endpoint.
func New(opt ...Option) Endpoint {
	opts := new(Options)
	for _, f := range opt {
		f(opts)
	}

	return &website{
		fs:   web.HttpFS(),
		opts: opts,
		data: web.MustLookup("index.html"),
	}
}

type website struct {
	opts *Options
	fs   http.FileSystem
	data []byte
}

func (w *website) Register(mux *gin.Engine) {
	h := http.FileServer(w.fs)
	h = setupCache(h)
	mux.GET("/favicon.svg", gin.WrapH(h))
	mux.GET("/assets/*filepath", gin.WrapH(h))
	mux.NoRoute(gin.WrapF(w.handleIndex))
}

func (w *website) handleIndex(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.WriteHeader(200)
	if _, err := rw.Write(w.data); err != nil {
		log.Error().Err(err).Msg("can not write index.html")
	}
}

func setupCache(h http.Handler) http.Handler {
	data := []byte(time.Now().String())
	etag := fmt.Sprintf("%x", md5.Sum(data))

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			w.Header().Del("Expires")
			w.Header().Set("ETag", etag)
			h.ServeHTTP(w, r)
		},
	)
}

// WithUser returns a context with the current authenticated user.
func WithUser(c context.Context, user *model.User) context.Context {
	return context.WithValue(c, userKey, user)
}

type key int

const userKey key = 0

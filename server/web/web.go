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
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/web/dist"

	"github.com/dimfeld/httptreemux"
)

// Endpoint provides the website endpoints.
type Endpoint interface {
	// Register registers the provider endpoints.
	Register(*httptreemux.ContextMux)
}

// New returns the default website endpoint.
func New(opt ...Option) Endpoint {
	opts := new(Options)
	for _, f := range opt {
		f(opts)
	}

	if opts.path != "" {
		return fromPath(opts)
	}

	return &website{
		fs:      dist.New(),
		opts:    opts,
		content: dist.MustLookup("/index.html"),
	}
}

func fromPath(opts *Options) *website {
	f := filepath.Join(opts.path, "index.html")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return &website{
		fs:      http.Dir(opts.path),
		content: b,
		opts:    opts,
	}
}

type website struct {
	opts    *Options
	fs      http.FileSystem
	content []byte
}

func (w *website) Register(mux *httptreemux.ContextMux) {
	h := http.FileServer(w.fs)
	h = setupCache(h)
	mux.Handler("GET", "/favicon.svg", h)
	mux.Handler("GET", "/static/*filepath", h)
	mux.NotFoundHandler = w.handleIndex
}

func (w *website) handleIndex(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(200)
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.Write(w.content)
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

// ToUser returns a user from the context.
func ToUser(c context.Context) (*model.User, bool) {
	user, ok := c.Value(userKey).(*model.User)
	return user, ok
}

type key int

const userKey key = 0

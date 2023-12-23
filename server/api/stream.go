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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

// EventStreamSSE
//
//	@Summary	Event stream
//	@Description	event source streaming for compatibility with quic and http2
//	@Router		/stream/events [get]
//	@Produce	plain
//	@Success	200
//	@Tags			Events
func EventStreamSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	rw := c.Writer

	flusher, ok := rw.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// ping the client
	logWriteStringErr(io.WriteString(rw, ": ping\n\n"))
	flusher.Flush()

	log.Debug().Msg("user feed: connection opened")

	user := session.User(c)
	repo := map[string]bool{}
	if user != nil {
		repos, _ := store.FromContext(c).RepoList(user, false, true)
		for _, r := range repos {
			repo[r.FullName] = true
		}
	}

	eventc := make(chan []byte, 10)
	ctx, cancel := context.WithCancelCause(
		context.Background(),
	)

	defer func() {
		cancel(nil)
		close(eventc)
		log.Debug().Msg("user feed: connection closed")
	}()

	go func() {
		server.Config.Services.Pubsub.Subscribe(ctx, func(m pubsub.Message) {
			defer func() {
				obj := recover() // fix #2480 // TODO: check if it's still needed
				log.Trace().Msgf("pubsub subscribe recover return: %v", obj)
			}()
			name := m.Labels["repo"]
			priv := m.Labels["private"]
			if repo[name] || priv == "false" {
				select {
				case <-ctx.Done():
					return
				default:
					eventc <- m.Data
				}
			}
		})
		cancel(nil)
	}()

	for {
		select {
		case <-rw.CloseNotify():
			return
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 30):
			logWriteStringErr(io.WriteString(rw, ": ping\n\n"))
			flusher.Flush()
		case buf, ok := <-eventc:
			if ok {
				logWriteStringErr(io.WriteString(rw, "data: "))
				logWriteStringErr(rw.Write(buf))
				logWriteStringErr(io.WriteString(rw, "\n\n"))
				flusher.Flush()
			}
		}
	}
}

// LogStreamSSE
//
//	@Summary	Log stream
//	@Router		/stream/logs/{repo_id}/{pipeline}/{stepID} [get]
//	@Produce	plain
//	@Success	200
//	@Tags			Pipeline logs
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		pipeline	path	int		true		"the number of the pipeline"
//	@Param		stepID		path	int		true		"the step id"
func LogStreamSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	rw := c.Writer

	flusher, ok := rw.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "Streaming not supported")
		return
	}

	logWriteStringErr(io.WriteString(rw, ": ping\n\n"))
	flusher.Flush()

	_store := store.FromContext(c)
	repo := session.Repo(c)

	pipeline, err := strconv.ParseInt(c.Param("pipeline"), 10, 64)
	if err != nil {
		log.Debug().Err(err).Msg("pipeline number invalid")
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: pipeline number invalid\n\n"))
		return
	}
	pl, err := _store.GetPipelineNumber(repo, pipeline)
	if err != nil {
		log.Debug().Msgf("stream cannot get pipeline number: %v", err)
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: pipeline not found\n\n"))
		return
	}

	stepID, err := strconv.ParseInt(c.Param("stepId"), 10, 64)
	if err != nil {
		log.Debug().Err(err).Msg("step id invalid")
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: step id invalid\n\n"))
		return
	}
	step, err := _store.StepLoad(stepID)
	if err != nil {
		log.Debug().Msgf("stream cannot get step number: %v", err)
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: process not found\n\n"))
		return
	}

	if step.PipelineID != pl.ID {
		// make sure we can not read arbitrary logs by id
		err = fmt.Errorf("step with id %d is not part of repo %s", stepID, repo.FullName)
		log.Debug().Err(err).Msg("event error")
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: "+err.Error()+"\n\n"))
		return
	}

	if step.State != model.StatusRunning {
		log.Debug().Msg("step not running (anymore).")
		logWriteStringErr(io.WriteString(rw, "event: error\ndata: step not running (anymore)\n\n"))
		return
	}

	logc := make(chan []byte, 10)
	ctx, cancel := context.WithCancelCause(
		context.Background(),
	)

	log.Debug().Msgf("log stream: connection opened")

	defer func() {
		cancel(nil)
		close(logc)
		log.Debug().Msgf("log stream: connection closed")
	}()

	go func() {
		err := server.Config.Services.Logs.Tail(ctx, step.ID, func(entries ...*model.LogEntry) {
			for _, entry := range entries {
				select {
				case <-ctx.Done():
					return
				default:
					ee, _ := json.Marshal(entry)
					logc <- ee
				}
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("tail of logs failed")
		}

		logWriteStringErr(io.WriteString(rw, "event: error\ndata: eof\n\n"))

		cancel(err)
	}()

	id := 1
	last, _ := strconv.Atoi(
		c.Request.Header.Get("Last-Event-ID"),
	)
	if last != 0 {
		log.Debug().Msgf("log stream: reconnect: last-event-id: %d", last)
	}

	// retry: 10000\n

	for {
		select {
		// after 1 hour of idle (no response) end the stream.
		// this is more of a safety mechanism than anything,
		// and can be removed once the code is more mature.
		case <-time.After(time.Hour):
			return
		case <-rw.CloseNotify():
			return
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 30):
			logWriteStringErr(io.WriteString(rw, ": ping\n\n"))
			flusher.Flush()
		case buf, ok := <-logc:
			if ok {
				if id > last {
					logWriteStringErr(io.WriteString(rw, "id: "+strconv.Itoa(id)))
					logWriteStringErr(io.WriteString(rw, "\n"))
					logWriteStringErr(io.WriteString(rw, "data: "))
					logWriteStringErr(rw.Write(buf))
					logWriteStringErr(io.WriteString(rw, "\n\n"))
					flusher.Flush()
				}
				id++
			}
		}
	}
}

func logWriteStringErr(_ int, err error) {
	if err != nil {
		log.Error().Err(err).Caller(1).Msg("fail to write string")
	}
}

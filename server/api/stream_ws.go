// Copyright 2026 Woodpecker Authors
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
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// wsAcceptOptions are the options used for all WebSocket upgrades in this package.
//
// InsecureSkipVerify is set because the existing CORS / origin policy is enforced
// at the router/middleware layer, matching the behavior of the SSE endpoints which
// do not perform Origin checks themselves.
var wsAcceptOptions = &websocket.AcceptOptions{
	InsecureSkipVerify: true,
	// CompressionMode set as disabled, as payloads are small JSON messages,
	// and disabling compression avoids the per-message-deflate overhead.
	CompressionMode: websocket.CompressionDisabled,
}

// EventStreamWS
//
//	@Summary		Stream events like pipeline updates over WebSocket
//	@Description	WebSocket variant of /stream/events. Each text frame contains the
//	@Description	same JSON payload that the SSE endpoint emits in `data:` lines.
//	@Router			/stream/ws/events [get]
//	@Produce		json
//	@Success		101
//	@Tags			Events
func EventStreamWS(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, wsAcceptOptions)
	if err != nil {
		log.Debug().Err(err).Msg("user feed: websocket accept failed")
		return
	}
	// CloseNow on defer guarantees the underlying TCP connection is released
	// even if the normal close handshake did not complete.
	defer func() { _ = conn.CloseNow() }()

	log.Debug().Msg("user feed: websocket connection opened")

	user := session.User(c)
	subTopics := make(map[string]struct{})
	// subscribe to all public state changes
	subTopics[pubsub.PublicTopic] = struct{}{}
	// subscribe to all private state changes or repos the user owns
	if user != nil {
		repos, _ := store.FromContext(c).RepoList(user, false, true, nil)
		for _, r := range repos {
			subTopics[pubsub.GetRepoTopic(r)] = struct{}{}
		}
	}

	eventChan := make(chan []byte, 10)
	ctx, cancel := context.WithCancelCause(c.Request.Context())

	defer func() {
		cancel(nil)
		log.Debug().Msg("user feed: websocket connection closed")
	}()

	// Reader pump: we don't expect client messages, but we must keep reading so
	// that control frames (close, ping, pong) are handled by the library.
	// CloseRead achieves exactly that and cancels ctx when the peer disconnects.
	ctx = conn.CloseRead(ctx)

	go func() {
		err := server.Config.Services.Scheduler.Subscribe(ctx, subTopics,
			func(m pubsub.Message) {
				select {
				case <-ctx.Done():
				case eventChan <- m.Data:
				}
			})
		cancel(err)
	}()

	pingTicker := time.NewTicker(idlePingTime)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = conn.Close(websocket.StatusNormalClosure, "")
			return
		case <-pingTicker.C:
			// Bound the ping with a short deadline so a stuck client doesn't
			// block the whole loop. coder/websocket's Ping waits for the pong.
			pingCtx, pingCancel := context.WithTimeout(ctx, idlePingTime)
			if err := conn.Ping(pingCtx); err != nil {
				pingCancel()
				log.Debug().Err(err).Msg("user feed: ping failed, closing")
				return
			}
			pingCancel()
		case buf, ok := <-eventChan:
			if !ok {
				return
			}
			writeCtx, writeCancel := context.WithTimeout(ctx, idlePingTime)
			err := conn.Write(writeCtx, websocket.MessageText, buf)
			writeCancel()
			if err != nil {
				log.Debug().Err(err).Msg("user feed: write failed, closing")
				return
			}
		}
	}
}

// LogStreamWS
//
//	@Summary	Stream logs of a pipeline step over WebSocket
//	@Router		/stream/ws/logs/{repo_id}/{pipeline}/{step_id} [get]
//	@Produce	json
//	@Success	101
//	@Tags		Pipeline logs
//	@Param		repo_id		path	int	true	"the repository id"
//	@Param		pipeline	path	int	true	"the number of the pipeline"
//	@Param		step_id		path	int	true	"the step id"
func LogStreamWS(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	// Validate parameters BEFORE upgrading. Errors are returned as plain HTTP
	// responses, matching how a fetch() client expects auth/validation failures.
	pipelineNum, err := strconv.ParseInt(c.Param("pipeline"), 10, 64)
	if err != nil {
		log.Debug().Err(err).Msg("pipeline number invalid")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	pl, err := _store.GetPipelineNumber(repo, pipelineNum)
	if err != nil {
		log.Debug().Err(err).Msg("stream cannot get pipeline number")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	stepID, err := strconv.ParseInt(c.Param("step_id"), 10, 64)
	if err != nil {
		log.Debug().Err(err).Msg("step id invalid")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	step, err := _store.StepLoad(pl.ID, stepID)
	if err != nil {
		log.Debug().Err(err).Msg("stream cannot get step number")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if step.State != model.StatusPending && step.State != model.StatusRunning {
		log.Debug().Msg("step not running (anymore).")
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, wsAcceptOptions)
	if err != nil {
		log.Debug().Err(err).Msg("log stream: websocket accept failed")
		return
	}
	defer func() { _ = conn.CloseNow() }()

	log.Debug().Msg("log stream: websocket connection opened")

	logChan := make(chan []byte, 10)
	ctx, cancel := context.WithCancelCause(c.Request.Context())

	defer func() {
		cancel(nil)
		log.Debug().Msg("log stream: websocket connection closed")
	}()

	ctx = conn.CloseRead(ctx)

	if err := server.Config.Services.Logs.Open(ctx, step.ID); err != nil {
		log.Error().Err(err).Msg("log stream: open failed")
		_ = conn.Close(websocket.StatusInternalError, "can't open stream")
		return
	}

	go func() {
		batches := make(logging.LogChan, maxQueuedBatchesPerClient)

		var innerDone sync.WaitGroup
		innerDone.Add(1)
		go func() {
			defer innerDone.Done()
			for entries := range batches {
				for _, entry := range entries {
					ee, err := json.Marshal(entry)
					if err != nil {
						log.Error().Err(err).Msg("unable to serialize log entry")
						continue
					}
					select {
					case <-ctx.Done():
						return
					case logChan <- ee:
					}
				}
			}
		}()

		err := server.Config.Services.Logs.Tail(ctx, step.ID, batches)
		if err != nil {
			log.Error().Err(err).Msg("tail of logs failed")
		}

		close(batches)
		innerDone.Wait()
		cancel(err)
	}()

	pingTicker := time.NewTicker(idlePingTime)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Distinguish a clean EOF (tail completed) from an aborted client.
			// On EOF we close with a normal status; the client uses that to
			// stop reconnecting.
			if cause := context.Cause(ctx); errors.Is(cause, context.Canceled) {
				log.Debug().Msg("log stream: eof")
				_ = conn.Close(websocket.StatusNormalClosure, "eof")
				return
			}
			_ = conn.Close(websocket.StatusNormalClosure, "")
			return
		case <-pingTicker.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, idlePingTime)
			if err := conn.Ping(pingCtx); err != nil {
				pingCancel()
				log.Debug().Err(err).Msg("log stream: ping failed, closing")
				return
			}
			pingCancel()
		case buf, ok := <-logChan:
			if !ok {
				return
			}
			writeCtx, writeCancel := context.WithTimeout(ctx, idlePingTime)
			err := conn.Write(writeCtx, websocket.MessageText, buf)
			writeCancel()
			if err != nil {
				log.Debug().Err(err).Msg("log stream: write failed, closing")
				return
			}
		}
	}
}

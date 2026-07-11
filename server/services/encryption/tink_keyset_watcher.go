// Copyright 2023 Woodpecker Authors
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

package encryption

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Watch keyset file events to detect key rotations and hot reload keys.
func (svc *tinkEncryptionService) initFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf(errTemplateTinkFailedSubscribeKeysetFileChanges, err)
	}
	err = watcher.Add(svc.keysetFilePath)
	if err != nil {
		return fmt.Errorf(errTemplateTinkFailedSubscribeKeysetFileChanges, err)
	}

	svc.keysetFileWatcher = watcher
	go svc.handleFileEvents()
	return nil
}

// isKeysetChangeEvent reports whether the file event indicates new keyset
// content. Op is a bitmask and may combine several operations in one event,
// so it must be tested with Has instead of equality.
func isKeysetChangeEvent(event fsnotify.Event) bool {
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create)
}

func (svc *tinkEncryptionService) handleFileEvents() {
	for {
		select {
		case event, ok := <-svc.keysetFileWatcher.Events:
			if !ok {
				log.Fatal().Msg(errMessageTinkKeysetFileWatchFailed) //nolint:forbidigo
			}
			if isKeysetChangeEvent(event) {
				log.Warn().Msgf(logTemplateTinkKeysetFileChanged, event.Name)
				err := svc.rotate()
				if err != nil {
					log.Fatal().Err(err).Msg(errMessageFailedRotatingEncryption) //nolint:forbidigo
				}
				// the rotated service runs its own watcher; close this
				// one instead of leaking it
				if err := svc.keysetFileWatcher.Close(); err != nil {
					log.Error().Err(err).Msgf(logTemplateTinkFailedClosingKeysetFile, svc.keysetFilePath)
				}
				return
			}
		case err, ok := <-svc.keysetFileWatcher.Errors:
			if !ok {
				log.Fatal().Err(err).Msg(errMessageTinkKeysetFileWatchFailed) //nolint:forbidigo
			}
		}
	}
}

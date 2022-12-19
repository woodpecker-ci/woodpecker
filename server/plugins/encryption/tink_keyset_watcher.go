// Copyright 2022 Woodpecker Authors
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
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Watch keyset file events to detect key rotations and hot reload keys
func (svc *tinkEncryptionService) initFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msgf("Error subscribing on encryption keyset file changes")
	}
	err = watcher.Add(svc.keysetFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error subscribing on encryption keyset file changes")
	}

	svc.keysetFileWatcher = watcher
	go svc.handleFileEvents()
}

func (svc *tinkEncryptionService) handleFileEvents() {
	for {
		select {
		case event, ok := <-svc.keysetFileWatcher.Events:
			if !ok {
				log.Fatal().Msg("Error watching encryption keyset file changes")
			}
			if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Create) {
				log.Warn().Msgf("Changes detected in encryption keyset file: '%s'. Encryption service will be reloaded", event.Name)
				svc.rotate()
				return
			}
		case err, ok := <-svc.keysetFileWatcher.Errors:
			if !ok {
				log.Fatal().Err(err).Msgf("Error watching encryption keyset file changes")
			}
		}
	}
}

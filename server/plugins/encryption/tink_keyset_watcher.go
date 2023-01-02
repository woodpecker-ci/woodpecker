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
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Watch keyset file events to detect key rotations and hot reload keys
func (svc *tinkEncryptionService) initFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed subscribing on encryption keyset file changes: %w", err)
	}
	err = watcher.Add(svc.keysetFilePath)
	if err != nil {
		return fmt.Errorf("failed subscribing on encryption keyset file changes: %w", err)
	}

	svc.keysetFileWatcher = watcher
	go svc.handleFileEvents()
	return nil
}

func (svc *tinkEncryptionService) handleFileEvents() {
	for {
		select {
		case event, ok := <-svc.keysetFileWatcher.Events:
			if !ok {
				log.Fatal().Msg("failed watching encryption keyset file changes")
			}
			if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Create) {
				log.Warn().Msgf("changes detected in encryption keyset file: '%s'. Encryption service will be reloaded", event.Name)
				err := svc.rotate()
				if err != nil {
					log.Fatal().Err(err).Msgf("failed rotating TINK encryption keyset")
				}
				return
			}
		case err, ok := <-svc.keysetFileWatcher.Errors:
			if !ok {
				log.Fatal().Err(err).Msgf("failed watching encryption keyset file changes")
			}
		}
	}
}

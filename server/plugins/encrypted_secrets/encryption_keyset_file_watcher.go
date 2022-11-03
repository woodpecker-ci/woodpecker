package encrypted_secrets

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Watch keyset file events to detect key rotations and hot reload keys
func (svc *Encryption) initFileWatcher() {
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

func (svc *Encryption) handleFileEvents() {
	for {
		select {
		case event, ok := <-svc.keysetFileWatcher.Events:
			if !ok {
				log.Fatal().Msg("Error watching encryption keyset file changes")
			}
			if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Create) {
				log.Warn().Msgf("Changes detected in encryption keyset file: '%s'. Encryption service will be reloaded", event.Name)
				svc.initEncryption()
			}
		case err, ok := <-svc.keysetFileWatcher.Errors:
			if !ok {
				log.Fatal().Err(err).Msgf("Error watching encryption keyset file changes")
			}
		}
	}
}

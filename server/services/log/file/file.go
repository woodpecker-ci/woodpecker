package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/log"
)

const (
	// Add base64 overhead and space for other JSON fields (just to be safe).
	maxLineLength int = (pipeline.MaxLogLineLength/3)*4 + (64 * 1024) //nolint:mnd
)

type logStore struct {
	base string
}

func NewLogStore(base string) (log.Service, error) {
	if base == "" {
		return nil, fmt.Errorf("file storage base path is required")
	}
	if _, err := os.Stat(base); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(base, 0o600)
		if err != nil {
			return nil, err
		}
	}
	return logStore{base: base}, nil
}

func (l logStore) filePath(id int64) string {
	return filepath.Join(l.base, fmt.Sprintf("%d.json", id))
}

func (l logStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	filename := l.filePath(step.ID)
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	buf := make([]byte, 0, bufio.MaxScanTokenSize)
	s := bufio.NewScanner(file)
	s.Buffer(buf, maxLineLength)

	var entries []*model.LogEntry
	for s.Scan() {
		j := s.Text()
		if len(strings.TrimSpace(j)) == 0 {
			continue
		}
		entry := &model.LogEntry{}
		err = json.Unmarshal([]byte(j), entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (l logStore) LogAppend(step *model.Step, logEntries []*model.LogEntry) error {
	path := l.filePath(step.ID)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		logger.Error().Err(err).Msgf("could not open log file %s", path)
		return err
	}

	var bytes []byte

	for _, logEntry := range logEntries {
		if jsonLine, err := json.Marshal(logEntry); err == nil {
			bytes = append(bytes, jsonLine...)
			bytes = append(bytes, byte('\n'))
		} else {
			logger.Error().Err(err).Msg("could not convert log entry to JSON")
		}
	}

	if _, err = file.Write(bytes); err != nil {
		logger.Error().Err(err).Msg("could not write out log entries")
	}

	return file.Close()
}

func (l logStore) LogDelete(step *model.Step) error {
	return os.Remove(l.filePath(step.ID))
}

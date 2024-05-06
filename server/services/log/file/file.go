package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/log"
)

type logStore struct {
	base string
}

func NewLogStore(base string) (log.Service, error) {
	if base == "" {
		return nil, fmt.Errorf("file storage base path is required")
	}
	if _, err := os.Stat(base); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(base, 0600)
		if err != nil {
			return nil, err
		}
	}
	return logStore{base: base}, nil
}

func (l logStore) filePath(id int64) string {
	return filepath.Join(l.base, strconv.Itoa(int(id))+".json")
}

func (l logStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	file, err := os.ReadFile(l.filePath(step.ID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []*model.LogEntry
	for _, j := range bytes.Split(file, []byte("\n")) {
		if len(bytes.TrimSpace(j)) == 0 {
			continue
		}
		entry := &model.LogEntry{}
		err = json.Unmarshal(j, entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (l logStore) LogAppend(logEntry *model.LogEntry) error {
	file, err := os.OpenFile(l.filePath(logEntry.StepID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}
	_, err = file.Write(append(jsonData, byte('\n')))
	if err != nil {
		return err
	}
	return file.Close()
}

func (l logStore) LogDelete(step *model.Step) error {
	return os.Remove(l.filePath(step.ID))
}

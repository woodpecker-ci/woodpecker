package log

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

type Service interface {
	LogFind(step *model.Step) ([]*model.LogEntry, error)
	LogAppend(logEntry *model.LogEntry) error
	LogDelete(step *model.Step) error
}

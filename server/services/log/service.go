package log

import "go.woodpecker-ci.org/woodpecker/v3/server/model"

type Service interface {
	LogFind(step *model.Step) ([]*model.LogEntry, error)
	LogAppend(step *model.Step, logEntries []*model.LogEntry) error
	LogDelete(step *model.Step) error
}

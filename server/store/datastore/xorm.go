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

package datastore

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	xlog "xorm.io/xorm/log"
)

func newXORMLogger(level xlog.LogLevel) xlog.Logger {
	return &xormLogger{
		logger: log.With().Str("component", "xorm").Logger(),
		level:  level,
	}
}

// xormLogger custom log implementation for ILogger
type xormLogger struct {
	logger  zerolog.Logger
	level   xlog.LogLevel
	showSQL bool
}

// Error implement ILogger
func (x *xormLogger) Error(v ...any) {
	if x.level <= xlog.LOG_ERR {
		x.logger.Error().Msg(fmt.Sprintln(v...))
	}
}

// Errorf implement ILogger
func (x *xormLogger) Errorf(format string, v ...any) {
	if x.level <= xlog.LOG_ERR {
		x.logger.Error().Msg(fmt.Sprintf(format, v...))
	}
}

// Debug implement ILogger
func (x *xormLogger) Debug(v ...any) {
	if x.level <= xlog.LOG_DEBUG {
		x.logger.Debug().Msg(fmt.Sprintln(v...))
	}
}

// Debugf implement ILogger
func (x *xormLogger) Debugf(format string, v ...any) {
	if x.level <= xlog.LOG_DEBUG {
		x.logger.Debug().Msg(fmt.Sprintf(format, v...))
	}
}

// Info implement ILogger
func (x *xormLogger) Info(v ...any) {
	if x.level <= xlog.LOG_INFO {
		x.logger.Info().Msg(fmt.Sprintln(v...))
	}
}

// Infof implement ILogger
func (x *xormLogger) Infof(format string, v ...any) {
	if x.level <= xlog.LOG_INFO {
		x.logger.Info().Msg(fmt.Sprintf(format, v...))
	}
}

// Warn implement ILogger
func (x *xormLogger) Warn(v ...any) {
	if x.level <= xlog.LOG_WARNING {
		x.logger.Warn().Msg(fmt.Sprintln(v...))
	}
}

// Warnf implement ILogger
func (x *xormLogger) Warnf(format string, v ...any) {
	if x.level <= xlog.LOG_WARNING {
		x.logger.Warn().Msg(fmt.Sprintf(format, v...))
	}
}

// Level implement ILogger
func (x *xormLogger) Level() xlog.LogLevel {
	return xlog.LOG_INFO
}

// SetLevel implement ILogger
func (x *xormLogger) SetLevel(l xlog.LogLevel) {
	x.level = l
}

// ShowSQL implement ILogger
func (x *xormLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		x.showSQL = true
		return
	}
	x.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (x *xormLogger) IsShowSQL() bool {
	return x.showSQL
}

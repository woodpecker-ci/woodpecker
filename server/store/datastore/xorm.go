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

	"github.com/rs/zerolog/log"
	xlog "xorm.io/xorm/log"
)

func NewXORMLogger(level xlog.LogLevel) *XORMLogger {
	return &XORMLogger{
		level: level,
	}
}

// XORMLogger custom log implementation for ILogger
type XORMLogger struct {
	level   xlog.LogLevel
	showSQL bool
}

var _ xlog.Logger = &XORMLogger{}

// Error implement ILogger
func (x *XORMLogger) Error(v ...interface{}) {
	if x.level <= xlog.LOG_ERR {
		log.Error().Msg(fmt.Sprintln(v...))
	}
}

// Errorf implement ILogger
func (x *XORMLogger) Errorf(format string, v ...interface{}) {
	if x.level <= xlog.LOG_ERR {
		log.Error().Msg(fmt.Sprintf(format, v...))
	}
}

// Debug implement ILogger
func (x *XORMLogger) Debug(v ...interface{}) {
	if x.level <= xlog.LOG_DEBUG {
		log.Debug().Msg(fmt.Sprintln(v...))
	}
}

// Debugf implement ILogger
func (x *XORMLogger) Debugf(format string, v ...interface{}) {
	if x.level <= xlog.LOG_DEBUG {
		log.Debug().Msg(fmt.Sprintf(format, v...))
	}
}

// Info implement ILogger
func (x *XORMLogger) Info(v ...interface{}) {
	if x.level <= xlog.LOG_INFO {
		log.Info().Msg(fmt.Sprintln(v...))
	}
}

// Infof implement ILogger
func (x *XORMLogger) Infof(format string, v ...interface{}) {
	if x.level <= xlog.LOG_INFO {
		log.Info().Msg(fmt.Sprintf(format, v...))
	}
}

// Warn implement ILogger
func (x *XORMLogger) Warn(v ...interface{}) {
	if x.level <= xlog.LOG_WARNING {
		log.Warn().Msg(fmt.Sprintln(v...))
	}
}

// Warnf implement ILogger
func (x *XORMLogger) Warnf(format string, v ...interface{}) {
	if x.level <= xlog.LOG_WARNING {
		log.Warn().Msg(fmt.Sprintf(format, v...))
	}
}

// Level implement ILogger
func (x *XORMLogger) Level() xlog.LogLevel {
	return xlog.LOG_INFO
}

// SetLevel implement ILogger
func (x *XORMLogger) SetLevel(l xlog.LogLevel) {
	x.level = l
}

// ShowSQL implement ILogger
func (x *XORMLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		x.showSQL = true
		return
	}
	x.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (x *XORMLogger) IsShowSQL() bool {
	return x.showSQL
}

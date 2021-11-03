// Copyright 2021 Woodpecker Authors
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

package datastore_xorm

import (
	"github.com/rs/zerolog/log"
	xorm_log "xorm.io/xorm/log"
)

type LogWrapper struct {
	isShowSQL bool
}

var _ xorm_log.Logger = &LogWrapper{}

func (l LogWrapper) Debug(v ...interface{}) {
	log.Debug().Msgf("xorm: %v", v)
}

func (l LogWrapper) Debugf(format string, v ...interface{}) {
	log.Debug().Msgf("xorm: "+format, v...)
}

func (l LogWrapper) Error(v ...interface{}) {
	log.Error().Msgf("xorm: %v", v)
}

func (l LogWrapper) Errorf(format string, v ...interface{}) {
	log.Error().Msgf("xorm: "+format, v...)
}

func (l LogWrapper) Info(v ...interface{}) {
	log.Info().Msgf("xorm: %v", v)
}

func (l LogWrapper) Infof(format string, v ...interface{}) {
	log.Info().Msgf("xorm: "+format, v...)
}

func (l LogWrapper) Warn(v ...interface{}) {
	log.Warn().Msgf("xorm: %v", v)
}

func (l LogWrapper) Warnf(format string, v ...interface{}) {
	log.Warn().Msgf("xorm: "+format, v...)
}

func (l LogWrapper) Level() xorm_log.LogLevel {
	return xorm_log.LOG_INFO // tmp
}

func (_ LogWrapper) SetLevel(_ xorm_log.LogLevel) {}

func (l LogWrapper) ShowSQL(show ...bool) {
	if len(show) != 0 {
		l.isShowSQL = show[0]
	}
}

func (l LogWrapper) IsShowSQL() bool {
	return l.isShowSQL
}

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

package migration

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type xormigrateLogger struct{}

func (l *xormigrateLogger) Debug(v ...any) {
	log.Debug().Msg(fmt.Sprint(v...))
}

func (l *xormigrateLogger) Debugf(format string, v ...any) {
	log.Debug().Msgf(format, v...)
}

func (l *xormigrateLogger) Info(v ...any) {
	log.Info().Msg(fmt.Sprint(v...))
}

func (l *xormigrateLogger) Infof(format string, v ...any) {
	log.Info().Msgf(format, v...)
}

func (l *xormigrateLogger) Warn(v ...any) {
	log.Warn().Msg(fmt.Sprint(v...))
}

func (l *xormigrateLogger) Warnf(format string, v ...any) {
	log.Warn().Msgf(format, v...)
}

func (l *xormigrateLogger) Error(v ...any) {
	log.Error().Msg(fmt.Sprint(v...))
}

func (l *xormigrateLogger) Errorf(format string, v ...any) {
	log.Error().Msgf(format, v...)
}

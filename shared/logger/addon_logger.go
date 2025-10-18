// Copyright 2024 Woodpecker Authors
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

package logger

import (
	"bytes"
	"io"
	std_log "log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AddonClientLogger struct {
	Logger   zerolog.Logger
	name     string
	withArgs []any
}

// cspell:words hclog

func convertLvl(level hclog.Level) zerolog.Level {
	switch level {
	case hclog.Error:
		return zerolog.ErrorLevel
	case hclog.Warn:
		return zerolog.WarnLevel
	case hclog.Info:
		return zerolog.InfoLevel
	case hclog.Debug:
		return zerolog.DebugLevel
	case hclog.Trace:
		return zerolog.TraceLevel
	}
	return zerolog.NoLevel
}

func (c *AddonClientLogger) applyArgs(args []any) *zerolog.Logger {
	var key string
	logger := c.Logger.With()
	args = append(args, c.withArgs)
	for i, arg := range args {
		switch {
		case key != "":
			logger.Any(key, arg)
			key = ""
		case i == len(args)-1:
			logger.Any(hclog.MissingKey, arg)
		default:

			key, _ = arg.(string)
		}
	}
	l := logger.Logger()
	return &l
}

func (c *AddonClientLogger) Log(level hclog.Level, msg string, args ...any) {
	c.applyArgs(args).WithLevel(convertLvl(level)).Msg(msg)
}

func (c *AddonClientLogger) Trace(msg string, args ...any) {
	c.applyArgs(args).Trace().Msg(msg)
}

func (c *AddonClientLogger) Debug(msg string, args ...any) {
	c.applyArgs(args).Debug().Msg(msg)
}

func (c *AddonClientLogger) Info(msg string, args ...any) {
	c.applyArgs(args).Info().Msg(msg)
}

func (c *AddonClientLogger) Warn(msg string, args ...any) {
	c.applyArgs(args).Warn().Msg(msg)
}

func (c *AddonClientLogger) Error(msg string, args ...any) {
	c.applyArgs(args).Error().Msg(msg)
}

func (c *AddonClientLogger) IsTrace() bool {
	return log.Logger.GetLevel() >= zerolog.TraceLevel
}

func (c *AddonClientLogger) IsDebug() bool {
	return log.Logger.GetLevel() >= zerolog.DebugLevel
}

func (c *AddonClientLogger) IsInfo() bool {
	return log.Logger.GetLevel() >= zerolog.InfoLevel
}

func (c *AddonClientLogger) IsWarn() bool {
	return log.Logger.GetLevel() >= zerolog.WarnLevel
}

func (c *AddonClientLogger) IsError() bool {
	return log.Logger.GetLevel() >= zerolog.ErrorLevel
}

func (c *AddonClientLogger) ImpliedArgs() []any {
	return c.withArgs
}

func (c *AddonClientLogger) With(args ...any) hclog.Logger {
	return &AddonClientLogger{
		Logger:   c.Logger,
		name:     c.name,
		withArgs: args,
	}
}

func (c *AddonClientLogger) Name() string {
	return c.name
}

func (c *AddonClientLogger) Named(name string) hclog.Logger {
	curr := c.name
	if curr != "" {
		curr = c.name + "."
	}
	return c.ResetNamed(curr + name)
}

func (c *AddonClientLogger) ResetNamed(name string) hclog.Logger {
	return &AddonClientLogger{
		Logger:   c.Logger,
		name:     name,
		withArgs: c.withArgs,
	}
}

func (c *AddonClientLogger) SetLevel(level hclog.Level) {
	c.Logger = c.Logger.Level(convertLvl(level))
}

func (c *AddonClientLogger) GetLevel() hclog.Level {
	switch c.Logger.GetLevel() {
	case zerolog.ErrorLevel:
		return hclog.Error
	case zerolog.WarnLevel:
		return hclog.Warn
	case zerolog.InfoLevel:
		return hclog.Info
	case zerolog.DebugLevel:
		return hclog.Debug
	case zerolog.TraceLevel:
		return hclog.Trace
	}
	return hclog.NoLevel
}

func (c *AddonClientLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *std_log.Logger {
	return std_log.New(c.StandardWriter(opts), "", 0)
}

func (c *AddonClientLogger) StandardWriter(*hclog.StandardLoggerOptions) io.Writer {
	return ioAdapter{logger: c.Logger}
}

type ioAdapter struct {
	logger zerolog.Logger
}

func (i ioAdapter) Write(p []byte) (n int, err error) {
	str := string(bytes.TrimRight(p, " \t\n"))
	i.logger.Log().Msg(str)
	return len(p), nil
}

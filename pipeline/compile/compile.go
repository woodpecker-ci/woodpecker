// Copyright 2026 Woodpecker Authors
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

// Package compile implements the transport and the semantics of the config
// response a compile workflow emits.
//
// It has no server, store or forge dependency, so the agent, the server and
// `cli exec` all apply the same rules to the same bytes.
package compile

import "errors"

const (
	// BlockBegin and BlockEnd delimit a config response in a step's raw output.
	// The payload between them is base64 of JSON, wrapped at BlockLineWidth
	// columns so a response can never hit the log line limit.
	BlockBegin = "-----BEGIN WOODPECKER CONFIG RESPONSE-----"
	BlockEnd   = "-----END WOODPECKER CONFIG RESPONSE-----"

	// BlockLineWidth is the column at which emitters should wrap the payload.
	// The scanner does not require it, it joins whatever it finds, but an
	// emitter that respects it stays well clear of the log line limit.
	BlockLineWidth = 64

	// ResponseVersion is the only response version understood today.
	ResponseVersion = 1

	// DefaultMaxPayloadSize caps the decoded payload of one compile workflow.
	// 4 MiB of yaml is an enormous pipeline.
	DefaultMaxPayloadSize = 4 * 1024 * 1024

	// DefaultMaxConfigs caps how many configs one compile workflow may emit.
	DefaultMaxConfigs = 64
)

var (
	// ErrNoResponse is returned when a compile step emitted no response block
	// at all. That is not the same as a block carrying an empty config list,
	// which means "proceed unchanged": a generator that crashed before it
	// printed anything must not be mistaken for one that decided to do nothing.
	ErrNoResponse = errors.New("compile step produced no config response")

	// ErrMalformedBlock is returned for a block that is nested, unterminated,
	// oversized or otherwise unreadable. Malformed blocks are rejected loudly
	// rather than skipped: most backends merge stdout and stderr into one
	// stream, so a mangled block usually means the step printed something
	// alongside its response.
	ErrMalformedBlock = errors.New("malformed config response block")

	// ErrUnsupportedVersion is returned for a response whose version this
	// Woodpecker does not understand.
	ErrUnsupportedVersion = errors.New("unsupported config response version")

	// ErrDuplicateName is returned when the same config name is emitted more
	// than once. Compile workflows finish in a nondeterministic order, so
	// last-writer-wins would not be reproducible.
	ErrDuplicateName = errors.New("duplicate config name")

	// ErrRemoveUnknownConfig is returned when a response removes a config that
	// is not part of the pipeline's config set.
	ErrRemoveUnknownConfig = errors.New("cannot remove unknown config")
)

// Config is a single pipeline configuration file. It mirrors
// forge/types.FileMeta, which this package cannot import without pulling in the
// server.
type Config struct {
	Name string
	Data []byte
}

// Response is the decoded payload of a config response block. It reuses the
// shape of the http config extension response, so an existing extension service
// becomes a compile plugin by changing transport only.
type Response struct {
	Version int              `json:"version"`
	Configs []ResponseConfig `json:"configs"`
}

// ResponseConfig is one entry of a Response. An entry with empty Data removes
// the config of that name from the pipeline's config set.
type ResponseConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// ToConfigs converts the response entries into Configs.
func (r *Response) ToConfigs() []Config {
	configs := make([]Config, 0, len(r.Configs))
	for _, config := range r.Configs {
		configs = append(configs, Config{Name: config.Name, Data: []byte(config.Data)})
	}

	return configs
}

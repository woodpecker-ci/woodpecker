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

package compile

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ScanOption configures a scanner.
type ScanOption func(*scanner)

// WithMaxPayloadSize overrides the cap on the decoded payload.
func WithMaxPayloadSize(size int) ScanOption {
	return func(s *scanner) {
		s.maxPayloadSize = size
	}
}

// ScanWriter wraps a step's raw output stream and extracts the config response
// blocks from it.
//
// Every line is forwarded: ordinary lines to logs, lines belonging to a
// response block to blockLogs. Nothing is swallowed, so the response stays
// auditable and callers can route block lines to a log type the normal log view
// hides. A nil writer discards.
//
// The returned writer must sit upstream of secret masking. The masker
// substring-replaces every secret in every line, and base64's alphabet makes a
// chance collision with a short secret likely enough in a large payload that a
// masked copy cannot be trusted as the authoritative one.
//
// The result function reports what was collected and must be called only after
// the stream was fully copied. It returns ErrNoResponse when the step emitted
// no block at all, which is different from a block carrying an empty config
// list.
func ScanWriter(logs, blockLogs io.Writer, opts ...ScanOption) (src io.Writer, result func() ([]Config, error)) {
	s := &scanner{
		logs:           logs,
		blockLogs:      blockLogs,
		maxPayloadSize: DefaultMaxPayloadSize,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s, s.result
}

type scanner struct {
	mu sync.Mutex

	logs           io.Writer
	blockLogs      io.Writer
	maxPayloadSize int

	// buf holds a partial trailing line between writes.
	buf []byte
	// inBlock is true between a begin marker and its matching end marker.
	inBlock bool
	// payload accumulates the encoded lines of the block being read.
	payload strings.Builder

	configs []Config
	found   bool
	err     error
}

func (s *scanner) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buf = append(s.buf, p...)

	for {
		idx := bytes.IndexByte(s.buf, '\n')
		if idx < 0 {
			break
		}

		line := s.buf[:idx+1]
		s.buf = s.buf[idx+1:]

		if err := s.line(line); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

// close flushes a trailing line that was not newline terminated and reports an
// unterminated block. It deliberately does not implement io.Closer: the caller
// owns the underlying stream.
func (s *scanner) close() error {
	if len(s.buf) > 0 {
		line := s.buf
		s.buf = nil
		if err := s.line(line); err != nil {
			return err
		}
	}

	if s.inBlock {
		return s.fail(fmt.Errorf("%w: block was never terminated by %q", ErrMalformedBlock, BlockEnd))
	}

	return nil
}

// line routes one line and advances the block state machine. The line still
// carries its trailing newline, if it had one.
func (s *scanner) line(line []byte) error {
	trimmed := strings.TrimRight(string(line), "\r\n")

	switch {
	case trimmed == BlockBegin:
		if s.inBlock {
			return s.fail(fmt.Errorf("%w: nested %q", ErrMalformedBlock, BlockBegin))
		}
		s.inBlock = true
		s.payload.Reset()
		return s.forward(s.blockLogs, line)

	case trimmed == BlockEnd:
		if !s.inBlock {
			return s.fail(fmt.Errorf("%w: %q without a preceding %q", ErrMalformedBlock, BlockEnd, BlockBegin))
		}
		s.inBlock = false
		if err := s.forward(s.blockLogs, line); err != nil {
			return err
		}
		return s.decode(s.payload.String())

	case s.inBlock:
		// Cap while reading, not after decoding: a step that never stops
		// printing inside a block would otherwise exhaust the agent's memory
		// long before the decoded payload could be measured.
		if s.payload.Len()+len(trimmed) > base64.StdEncoding.EncodedLen(s.maxPayloadSize) {
			return s.fail(fmt.Errorf("%w: encoded payload exceeds the limit of %d decoded bytes",
				ErrMalformedBlock, s.maxPayloadSize))
		}
		s.payload.WriteString(strings.TrimSpace(trimmed))
		return s.forward(s.blockLogs, line)

	default:
		return s.forward(s.logs, line)
	}
}

func (s *scanner) forward(dst io.Writer, line []byte) error {
	if dst == nil {
		return nil
	}

	_, err := dst.Write(line)

	return err
}

// decode turns one completed block into configs and appends them to the result.
func (s *scanner) decode(payload string) error {
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return s.fail(fmt.Errorf("%w: payload is not valid base64: %w", ErrMalformedBlock, err))
	}

	if len(raw) > s.maxPayloadSize {
		return s.fail(fmt.Errorf("%w: decoded payload of %d bytes exceeds the limit of %d bytes",
			ErrMalformedBlock, len(raw), s.maxPayloadSize))
	}

	var response Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return s.fail(fmt.Errorf("%w: payload is not valid json: %w", ErrMalformedBlock, err))
	}

	if response.Version != ResponseVersion {
		return s.fail(fmt.Errorf("%w: got %d, this version of Woodpecker understands %d",
			ErrUnsupportedVersion, response.Version, ResponseVersion))
	}

	s.found = true
	s.configs = append(s.configs, response.ToConfigs()...)

	return nil
}

// fail records the first error and returns it. Recording matters because the
// copy loop may swallow the error Write returned, and result must still be able
// to report that the response is unusable.
func (s *scanner) fail(err error) error {
	if s.err == nil {
		s.err = err
	}

	return err
}

func (s *scanner) result() ([]Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.close(); s.err == nil && err != nil {
		s.err = err
	}

	switch {
	case s.err != nil:
		return nil, s.err
	case !s.found:
		return nil, ErrNoResponse
	}

	return s.configs, nil
}

// EncodeResponse renders configs as a config response block. Woodpecker itself
// only ever decodes; this exists so tests, `cli exec` and the docs share one
// definition of the wire format with the scanner.
func EncodeResponse(configs []Config) (string, error) {
	response := Response{Version: ResponseVersion, Configs: make([]ResponseConfig, 0, len(configs))}
	for _, config := range configs {
		response.Configs = append(response.Configs, ResponseConfig{Name: config.Name, Data: string(config.Data)})
	}

	raw, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("could not encode config response: %w", err)
	}

	var out strings.Builder
	out.WriteString(BlockBegin + "\n")
	for encoded := base64.StdEncoding.EncodeToString(raw); ; {
		if len(encoded) <= BlockLineWidth {
			out.WriteString(encoded + "\n")
			break
		}
		out.WriteString(encoded[:BlockLineWidth] + "\n")
		encoded = encoded[BlockLineWidth:]
	}
	out.WriteString(BlockEnd + "\n")

	return out.String(), nil
}

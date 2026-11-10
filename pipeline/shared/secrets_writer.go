// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"bytes"
	"cmp"
	"io"
	"slices"
	"strings"
)

// Strings not longer than minSecretLength are not considered secrets.
// Do not sanitize them.
const minSecretLength = 3

var secretsMask = []byte("********")

type secretsReplaceWriter struct {
	dst     io.Writer
	secrets [][]byte
}

// NewSecretsReplaceWriter wraps dst so that the given secret values are
// replaced with asterisks before being written. It is meant to wrap a
// line-oriented writer, as secrets are matched per write. Therefore each
// secret is split on newlines to handle multi-line secrets.
func NewSecretsReplaceWriter(dst io.Writer, secrets []string) io.Writer {
	w := &secretsReplaceWriter{dst: dst}
	for _, secret := range secrets {
		for part := range strings.SplitSeq(strings.TrimSpace(secret), "\n") {
			if len(part) <= minSecretLength {
				continue
			}
			w.secrets = append(w.secrets, []byte(part))
		}
	}
	// Secrets are replaced one after another, so mask the longest first.
	// Else a secret that is part of a longer one would leak the rest of it.
	slices.SortStableFunc(w.secrets, func(a, b []byte) int {
		return cmp.Compare(len(b), len(a))
	})
	return w
}

func (w *secretsReplaceWriter) Write(p []byte) (int, error) {
	out := p
	for _, secret := range w.secrets {
		// lines without a secret are passed on as is, without a copy
		if bytes.Contains(out, secret) {
			out = bytes.ReplaceAll(out, secret, secretsMask)
		}
	}
	if _, err := w.dst.Write(out); err != nil {
		return 0, err
	}
	// report the consumed input length, masking changes the written length
	return len(p), nil
}

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

package log_test

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
)

type testWriter struct {
	*sync.Mutex
	writes []string
}

func (b *testWriter) Write(p []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()
	b.writes = append(b.writes, string(p))
	return len(p), nil
}

func (b *testWriter) Close() error {
	return nil
}

func (b *testWriter) GetWrites() []string {
	b.Lock()
	defer b.Unlock()
	w := make([]string, len(b.writes))
	copy(w, b.writes)
	return w
}

func TestCopyLineByLine(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	go func() {
		err := log.CopyLineByLine(testWriter, r, 1024)
		assert.NoError(t, err)
	}()

	// wait for the goroutine to start
	time.Sleep(time.Millisecond)

	// write 4 bytes without newline
	if _, err := w.Write([]byte("1234")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writes := testWriter.GetWrites()
	assert.Lenf(t, writes, 0, "expected 0 writes, got: %v", writes)

	// write more bytes with newlines
	if _, err := w.Write([]byte("5\n678\n90")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writes = testWriter.GetWrites()
	assert.Lenf(t, writes, 2, "expected 2 writes, got: %v", writes)

	// wait for the goroutine to write the data
	time.Sleep(10 * time.Millisecond)

	writtenData := strings.Join(writes, "-")
	assert.Equal(t, "12345\n-678\n", writtenData, "unexpected writtenData: %s", writtenData)

	// closing the writer should flush the remaining data
	w.Close()

	// wait for the goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// the written data contains all the data we wrote
	writtenData = strings.Join(testWriter.GetWrites(), "-")
	assert.Equal(t, "12345\n-678\n-90", writtenData, "unexpected writtenData: %s", writtenData)
}

func TestCopyLineByLineSizeLimit(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := log.CopyLineByLine(testWriter, r, 4)
		assert.NoError(t, err)
	}()

	// wait for the goroutine to start
	time.Sleep(time.Millisecond)

	// write 4 bytes without newline
	if _, err := w.Write([]byte("12345")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writes := testWriter.GetWrites()
	assert.Lenf(t, testWriter.GetWrites(), 0, "expected 0 writes, got: %v", writes)

	// write more bytes
	if _, err := w.Write([]byte("67\n89")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// wait for writer to write
	time.Sleep(time.Millisecond)

	writes = testWriter.GetWrites()
	assert.Lenf(t, testWriter.GetWrites(), 2, "expected 2 writes, got: %v", writes)

	writes = testWriter.GetWrites()
	writtenData := strings.Join(writes, "-")
	assert.Equal(t, "1234-567\n", writtenData, "unexpected writtenData: %s", writtenData)

	// closing the writer should flush the remaining data
	w.Close()

	wg.Wait()
}

func TestStringReader(t *testing.T) {
	r := io.NopCloser(strings.NewReader("123\n4567\n890"))

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	err := log.CopyLineByLine(testWriter, r, 1024)
	assert.NoError(t, err)

	writes := testWriter.GetWrites()
	assert.Lenf(t, writes, 3, "expected 3 writes, got: %v", writes)
}

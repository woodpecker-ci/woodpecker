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

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/log"
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

	done := make(chan struct{})

	go func() {
		err := log.CopyLineByLine(testWriter, r, 1024)
		assert.NoError(t, err)
		close(done)
	}()

	// write 4 bytes without newline
	if _, err := w.Write([]byte("1234")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait until no writes have occurred (should be immediate)
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 0
	}, time.Second, 5*time.Millisecond, "expected 0 writes after first write")

	// write more bytes with newlines
	if _, err := w.Write([]byte("5\n678\n90")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait until two writes have occurred
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 2
	}, time.Second, 5*time.Millisecond, "expected 2 writes after second write")

	writes := testWriter.GetWrites()
	writtenData := strings.Join(writes, "-")
	assert.Equal(t, "12345\n-678\n", writtenData, "unexpected writtenData: %s", writtenData)

	// closing the writer should flush the remaining data
	w.Close()

	// wait for the goroutine to finish
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for goroutine to finish")
	}

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
	assert.Lenf(t, testWriter.GetWrites(), 1, "expected 1 writes, got: %v", writes)

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

func TestCopyLineByLineNewlineCharacter(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	done := make(chan struct{})

	go func() {
		err := log.CopyLineByLine(testWriter, r, 4)
		assert.NoError(t, err)
		close(done)
	}()

	// write one newline character before the maximum size of the buffer
	if _, err := w.Write([]byte("123\n45678")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait until 2 writes have occurred
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 2
	}, time.Second, 5*time.Millisecond, "expected 2 writes after first write")

	writes := testWriter.GetWrites()
	writtenData := strings.Join(writes, "-")
	assert.Equal(t, "123\n-4567", writtenData)

	// write one newline character at the beginning before the maximum size of the buffer
	if _, err := w.Write([]byte("\n123\n45678")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait until 5 writes have occurred (2 from before + 3 new ones)
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 5
	}, time.Second, 5*time.Millisecond, "expected 5 writes total after second write")

	writes = testWriter.GetWrites()
	writtenData = strings.Join(writes, "-")
	assert.Equal(t, "123\n-4567-8\n-123\n-4567", writtenData)

	// Close the writer first to signal EOF
	w.Close()

	// wait for the goroutine to finish
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for goroutine to finish")
	}

	// Verify final flush (should have "8" remaining)
	writes = testWriter.GetWrites()
	writtenData = strings.Join(writes, "-")
	assert.Equal(t, "123\n-4567-8\n-123\n-4567-8", writtenData)
}

// TestCopyLineByLineLongLine is for the long line testing to trigger the writeChunks function.
func TestCopyLineByLineLongLine(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	done := make(chan struct{})

	// max size = 10
	maxSize := 10

	go func() {
		err := log.CopyLineByLine(testWriter, r, maxSize)
		assert.NoError(t, err)
		close(done)
	}()

	// wait for the goroutine to start
	time.Sleep(time.Millisecond)

	// will trigger the writeChunks function
	if _, err := w.Write([]byte("this is a very long line\n")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait for the writer to write
	time.Sleep(time.Millisecond)

	// verify the number of writes is equal to 3
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 3
	}, time.Second, 5*time.Millisecond, "expected 3 writes after first write")

	// verify all data was written correctly
	writtenData := ""
	assert.Eventually(t, func() bool {
		writtenData = strings.Join(testWriter.GetWrites(), "-")
		return writtenData == "this is a -very long -line\n"
	}, time.Second, 5*time.Millisecond, "unexpected writtenData: %s", writtenData)

	// closing the writer should flush the remaining data
	w.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for goroutine to finish")
	}
}

// TestCopyLineByLineWriteChunks is for the writeChunks function testing.
func TestCopyLineByLineWriteChunks(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		Mutex:  &sync.Mutex{},
		writes: make([]string, 0),
	}

	done := make(chan struct{})

	// max size = 8
	maxSize := 8

	go func() {
		err := log.CopyLineByLine(testWriter, r, maxSize)
		assert.NoError(t, err)
		close(done)
	}()

	// first line: 20 chars + newline = 21 bytes (will be chunked: 8 + 8 + 5)
	// second line: 5 chars + newline = 6 bytes (normal write, no chunking)
	// third line: 16 chars + newline = 17 bytes (will be chunked: 8 + 9)
	input := "12345678901234567890\n" +
		"short\n" +
		"abcdefghijklmnop\n"

	if _, err := w.Write([]byte(input)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify the number of writes is equal to 7
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == 7
	}, time.Second, 5*time.Millisecond, "expected 7 writes after first write")

	// verify all data was written correctly
	writtenData := ""
	assert.Eventually(t, func() bool {
		writtenData = strings.Join(testWriter.GetWrites(), "")
		return writtenData == input
	}, time.Second, 5*time.Millisecond, "unexpected writtenData: %s", writtenData)

	// verify the number of writes
	expectedWrites := 7
	assert.Eventually(t, func() bool {
		return len(testWriter.GetWrites()) == expectedWrites
	}, time.Second, 5*time.Millisecond, "expected %d writes, got %d: %v", expectedWrites, len(testWriter.GetWrites()), testWriter.GetWrites())

	writes := testWriter.GetWrites()
	// verify first line chunks
	assert.Equal(t, "12345678", writes[0], "first chunk of first line")
	assert.Equal(t, "90123456", writes[1], "second chunk of first line")
	assert.Equal(t, "7890\n", writes[2], "third chunk of first line")

	// verify second line (not chunked)
	assert.Equal(t, "short\n", writes[3], "second line should not be chunked")

	// verify third line chunks
	assert.Equal(t, "abcdefgh", writes[4], "first chunk of third line")
	assert.Equal(t, "ijklmnop", writes[5], "second chunk of third line")
	assert.Equal(t, "\n", writes[6], "third chunk of third line (just newline)")

	// closing the writer should flush the remaining data
	w.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for goroutine to finish")
	}
}

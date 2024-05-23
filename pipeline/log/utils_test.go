package log_test

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
)

type testWriter struct {
	writes []string
}

func (b *testWriter) Write(p []byte) (n int, err error) {
	b.writes = append(b.writes, string(p))
	return len(p), nil
}

func (b *testWriter) Close() error {
	return nil
}

func TestCopyLineByLine(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
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

	assert.Lenf(t, testWriter.writes, 0, "expected 0 writes, got: %v", testWriter.writes)

	// write more bytes with newlines
	if _, err := w.Write([]byte("5\n678\n90")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait for the goroutine to write the data
	time.Sleep(time.Millisecond)

	assert.Lenf(t, testWriter.writes, 2, "expected 2 writes, got: %v", testWriter.writes)

	writtenData := strings.Join(testWriter.writes, "-")
	assert.Equal(t, "12345\n-678\n", writtenData, "unexpected writtenData: %s", writtenData)

	// closing the writer should flush the remaining data
	w.Close()

	// wait for the goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// the written data contains all the data we wrote
	writtenData = strings.Join(testWriter.writes, "-")
	assert.Equal(t, "12345\n-678\n-90", writtenData, "unexpected writtenData: %s", writtenData)
}

func TestCopyLineByLineSizeLimit(t *testing.T) {
	r, w := io.Pipe()

	testWriter := &testWriter{
		writes: make([]string, 0),
	}

	go func() {
		err := log.CopyLineByLine(testWriter, r, 4)
		assert.NoError(t, err)
	}()

	// wait for the goroutine to start
	time.Sleep(time.Millisecond)

	// write 4 bytes without newline
	if _, err := w.Write([]byte("12345")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Lenf(t, testWriter.writes, 0, "expected 0 writes, got: %v", testWriter.writes)

	// write more bytes
	if _, err := w.Write([]byte("67\n89")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wait for the goroutine to write the data
	time.Sleep(time.Millisecond)

	assert.Lenf(t, testWriter.writes, 2, "expected 2 writes, got: %v", testWriter.writes)

	writtenData := strings.Join(testWriter.writes, "-")
	assert.Equal(t, "1234-567\n", writtenData, "unexpected writtenData: %s", writtenData)
}

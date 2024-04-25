package logger_test

import (
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v2/agent/logger"
)

type testBuffer struct {
	buf     []byte
	flushes int
}

func (b *testBuffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	b.flushes++
	return len(p), nil
}

func TestFlushAfterSize(t *testing.T) {
	bufSize := 4
	bufTime := 10 * time.Minute // using a high value to avoid the timer to trigger

	testBuffer := &testBuffer{
		buf:     make([]byte, 0),
		flushes: 0,
	}
	logBuffer := logger.NewLogBuffer(testBuffer, bufSize, bufTime)
	defer logBuffer.Close()

	// Write 4 bytes (exact buffer size, so fill buffer)
	if _, err := logBuffer.Write([]byte("123")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(testBuffer.buf) != "" {
		t.Fatalf("expected 0 bytes, got %s", testBuffer.buf)
	}

	// Write 4 more bytes (buffer should be flushed once)
	if _, err := logBuffer.Write([]byte("4567")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(testBuffer.buf) != "1234" {
		t.Fatalf("expected 1234, got %s", testBuffer.buf)
	}

	// Write 2 more bytes (buffer should be flushed again)
	if _, err := logBuffer.Write([]byte("89")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if the buffer is flushed
	if testBuffer.flushes != 2 {
		t.Fatalf("expected 2 flushes, got %d", testBuffer.flushes)
	}

	if string(testBuffer.buf) != "12345678" {
		t.Fatalf("expected 12345678, got %s", testBuffer.buf)
	}
}

func TestFlushAfterTime(t *testing.T) {
	bufSize := 1024 // using a high value to avoid the buffer to be flushed by size
	bufTime := 10 * time.Millisecond

	testBuffer := &testBuffer{
		buf: make([]byte, 0),
	}

	logBuffer := logger.NewLogBuffer(testBuffer, bufSize, bufTime)
	defer logBuffer.Close()

	// Write 4 bytes
	if _, err := logBuffer.Write([]byte("1234")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if the buffer is empty
	if len(testBuffer.buf) != 0 {
		t.Fatalf("expected 0 bytes, got %d", len(testBuffer.buf))
	}

	// Wait for the buffer to be flushed
	time.Sleep(20 * time.Millisecond)

	// Check if the buffer is flushed
	if len(testBuffer.buf) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(testBuffer.buf))
	}
}

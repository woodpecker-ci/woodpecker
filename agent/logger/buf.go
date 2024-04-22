package logger

import (
	"bufio"
	"io"
	"time"
)

type LogBuffer struct {
	buffer        *bufio.Writer
	flushInterval time.Duration
	timer         *time.Timer
	flushChan     chan struct{}
	closeChan     chan struct{}
}

func NewLogBuffer(writer io.Writer, bufferSize int, flushInterval time.Duration) *LogBuffer {
	lb := &LogBuffer{
		buffer:        bufio.NewWriterSize(writer, bufferSize),
		flushInterval: flushInterval,
		timer:         time.NewTimer(flushInterval),
		flushChan:     make(chan struct{}, 1),
	}
	go lb.start()
	return lb
}

func (lb *LogBuffer) Write(data []byte) (int, error) {
	n, err := lb.buffer.Write(data)
	if err != nil {
		return n, err
	}

	// Reset timer since there's new activity
	if !lb.timer.Stop() {
		<-lb.timer.C
	}
	lb.timer.Reset(lb.flushInterval)

	return n, nil
}

func (lb *LogBuffer) start() {
	for {
		if !lb.waitForFlush() {
			break
		}
	}
}

func (lb *LogBuffer) waitForFlush() bool {
	lb.timer.Stop()
	defer lb.timer.Reset(lb.flushInterval)

	// Wait for either a timeout or a manual flush signal
	select {
	case <-lb.timer.C:
		// Time limit reached, flush the buffer
		lb.buffer.Flush()
	case <-lb.flushChan:
		// Manual flush signal received
		lb.buffer.Flush()
	case <-lb.closeChan:
		// Close signal received
		return false
	}

	return true
}

func (lb *LogBuffer) Flush() error {
	lb.flushChan <- struct{}{}
	return nil
}

func (lb *LogBuffer) Close() error {
	lb.timer.Stop()
	close(lb.flushChan)
	return lb.buffer.Flush()
}

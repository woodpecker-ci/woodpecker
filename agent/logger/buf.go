package logger

import (
	"bufio"
	"io"
	"sync"
	"time"
)

type LogBuffer struct {
	*sync.Mutex
	buffer        *bufio.Writer
	flushInterval time.Duration
	timer         *time.Timer
	closeChan     chan struct{}
}

func NewLogBuffer(writer io.Writer, bufferSize int, flushInterval time.Duration) *LogBuffer {
	lb := &LogBuffer{
		Mutex:         &sync.Mutex{},
		buffer:        bufio.NewWriterSize(writer, bufferSize),
		flushInterval: flushInterval,
		timer:         time.NewTimer(flushInterval),
		closeChan:     make(chan struct{}),
	}
	go lb.start()
	return lb
}

func (lb *LogBuffer) Write(data []byte) (int, error) {
	n, err := lb.buffer.Write(data)
	if err != nil {
		return n, err
	}

	// reset timer since there's new activity
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
	// wait for either a timeout or a manual flush signal
	select {
	case <-lb.timer.C:
		// time limit reached, flush the buffer
		lb.Lock()
		defer lb.Unlock()
		err := lb.buffer.Flush()
		if err != nil {
			return false
		}
	case <-lb.closeChan:
		// close signal received
		return false
	}

	return true
}

func (lb *LogBuffer) Flush() error {
	lb.Lock()
	defer lb.Unlock()
	return lb.buffer.Flush()
}

func (lb *LogBuffer) Close() error {
	lb.timer.Stop()
	close(lb.closeChan)
	return lb.buffer.Flush()
}

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

package log

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

func writeChunks(dst io.Writer, data []byte, size int) error {
	if len(data) <= size {
		_, err := dst.Write(data)
		return err
	}

	for len(data) > size {
		if _, err := dst.Write(data[:size]); err != nil {
			return err
		}
		data = data[size:]
	}

	if len(data) > 0 {
		_, err := dst.Write(data)
		return err
	}

	return nil
}

func CopyLineByLine(dst io.Writer, src io.Reader, maxSize int) error {
	r := bufio.NewReader(src)
	// buffer to cache
	var buf []byte
	// buffer to read
	readBuf := make([]byte, maxSize)

	for {
		n, err := r.Read(readBuf)

		// handle the data first
		if n > 0 {
			// if it has data, cache into the buffer
			buf = append(buf, readBuf[:n]...)

		processBuffer:
			for len(buf) > 0 {
				// find the index to anchor the new line
				idx := bytes.IndexByte(buf, '\n')
				switch {
				case idx >= 0:
					// found the new line, write to the dst
					lineEnd := idx + 1
					if lineEnd > maxSize {
						if wErr := writeChunks(dst, buf[:lineEnd], maxSize); wErr != nil {
							return wErr
						}
					} else {
						if _, wErr := dst.Write(buf[:lineEnd]); wErr != nil {
							return wErr
						}
					}
					// remove the line written from the buffer
					buf = buf[lineEnd:]
				case len(buf) >= maxSize:
					if _, wErr := dst.Write(buf[:maxSize]); wErr != nil {
						return wErr
					}
					buf = buf[maxSize:]
				default:
					// no newline found and buffer not full, read more data
					break processBuffer
				}
			}
		}

		// and then if it is EOF, write the remaining data and break the loop
		if errors.Is(err, io.EOF) {
			if len(buf) == 0 {
				break
			}
			if _, wErr := dst.Write(buf); wErr != nil {
				return wErr
			}
			break
		}

		if err != nil {
			return err
		}
	}
	return nil
}

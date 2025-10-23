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

	for {
		// read until newline or maxSize, whichever comes first
		line, err := r.ReadBytes('\n')

		// if line is longer than maxSize, we need to chunk it
		if len(line) > maxSize {
			// write in chunks of maxSize
			if werr := writeChunks(dst, line, maxSize); werr != nil {
				return werr
			}
		} else if len(line) > 0 {
			// line fits in maxSize, write it directly
			if _, werr := dst.Write(line); werr != nil {
				return werr
			}
		}

		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}

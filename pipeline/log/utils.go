package log

import (
	"bufio"
	"io"
)

func writeChunks(dst io.WriteCloser, data []byte, size int) (int, error) {
	if len(data) <= size {
		return dst.Write(data)
	}

	for len(data) > size {
		if _, err := dst.Write(data[:size]); err != nil {
			return 0, err
		}
		data = data[size:]
	}

	if len(data) > 0 {
		return dst.Write(data)
	}

	return 0, nil
}

func CopyLineByLine(dst io.WriteCloser, src io.Reader, maxSize int) error {
	r := bufio.NewReader(src)
	defer dst.Close()

	for {
		// TODO: read til newline or maxSize directly
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			_, err = writeChunks(dst, line, maxSize)
			return err
		} else if err != nil {
			return err
		}

		_, err = writeChunks(dst, line, maxSize)
		if err != nil {
			return err
		}
	}
}

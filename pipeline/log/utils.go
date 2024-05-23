package log

import (
	"bufio"
	"io"
)

func CopyLineByLine(dst io.WriteCloser, src io.Reader) error {
	r := bufio.NewReader(src)
	defer dst.Close()
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			if len(line) > 0 {
				dst.Write(line)
			}

			return nil
		} else if err != nil {
			return err
		}

		dst.Write(line)
	}
}

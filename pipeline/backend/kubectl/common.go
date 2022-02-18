package kubectl

import (
	"bytes"
	"io"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func CreateRandomID(
	n int,
) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetReaderContents(reader io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func FirstNotEmpty(args ...interface{}) interface{} {
	for _, arg := range args {
		switch arg.(type) {
		case string:
			if len(args) > 0 {
				return arg
			}
		default:
			return arg
		}
	}
	return nil
}

// select between two options.
func Triary(
	condition bool,
	a interface{},
	b interface{},
) interface{} {
	if condition {
		return a
	}
	return b
}

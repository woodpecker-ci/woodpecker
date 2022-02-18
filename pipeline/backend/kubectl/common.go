package kubectl

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"strings"
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

// Returns an env if exists otherwise returns default value.
func getEnv(name string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(name)
	return Triary(exists, val, defaultValue)
}

// Returns an env if exists otherwise returns default value.
// NOTE: prepends WOODPECKER_KUBECTL_ to name, and uppercase the name.
func getWPKEnv(name string, defaultValue interface{}) interface{} {
	name = strings.ToUpper("WOODPECKER_KUBECTL_" + name)
	return getEnv(name, defaultValue)
}

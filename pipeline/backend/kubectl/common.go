package kubectl

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Creates a time seeded random id.
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

// Converts a string into a kubernetes valid FQDN name.
func ToKubernetesValidName(name string, maxChars int) string {
	name = strings.ToLower(name)

	// cleanup chars
	re, _ := regexp.Compile("[^a-z0-9]+")
	name = string(re.ReplaceAll([]byte(name), []byte("-")))

	if maxChars > -1 && len(name) > maxChars {
		name = name[len(name)-maxChars:]
	}

	// cleanup starters and enders
	re, _ = regexp.Compile("^-+|-+$")
	name = string(re.ReplaceAll([]byte(name), []byte("")))

	return name
}

// Returns the string representation of a pipe content.
func ReadPipeAsString(reader io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Selects the first non empty interface.
func FirstNotEmpty(args ...interface{}) interface{} {
	for _, arg := range args {
		switch arg.(type) {
		case string:
			if len(args) == 0 {
				continue
			}
			return arg
		default:
			if arg == nil {
				continue
			}
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

var IsIPRegex *regexp.Regexp

func IsIP(ip string) bool {
	if IsIPRegex == nil {
		IsIPRegex = regexp.MustCompile(`((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))`)
	}
	return IsIPRegex.Match([]byte(ip))
}

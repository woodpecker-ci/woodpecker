package meddler

import (
	"strings"
	"unicode"
)

// MapperFunc signature. Argument is field name, return value is database column.
type MapperFunc func(in string) string

// Mapper defines the function to transform struct field names into database columns.
// Default is strings.TrimSpace, basically a no-op
var Mapper MapperFunc = strings.TrimSpace

// LowerCase returns a lowercased version of the input string
func LowerCase(in string) string {
	return strings.ToLower(in)
}

// SnakeCase returns a snake_cased version of the input string
func SnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

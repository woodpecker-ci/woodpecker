package template

import "bytes"

// hexdump is a template function that creates a hux dump
// similar to xxd -i.
func hexdump(v interface{}) string {
	var data []byte
	switch vv := v.(type) {
	case []byte:
		data = vv
	case string:
		data = []byte(vv)
	default:
		return ""
	}
	var buf bytes.Buffer
	for i, b := range data {
		dst := make([]byte, 4)
		src := []byte{b}
		encode(dst, src, ldigits)
		buf.Write(dst)

		buf.WriteString(",")
		if (i+1)%cols == 0 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// default number of columns
const cols = 12

// hex lookup table for hex encoding
const (
	ldigits = "0123456789abcdef"
	udigits = "0123456789ABCDEF"
)

func encode(dst, src []byte, hextable string) {
	dst[0] = '0'
	dst[1] = 'x'
	for i, v := range src {
		dst[i+1*2] = hextable[v>>4]
		dst[i+1*2+1] = hextable[v&0x0f]
	}
}

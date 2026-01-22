package fonts

import (
	_ "embed"
)

// VeraSans is vera.ttf font inlined to the bytes slice.
//
//go:embed vera.ttf
var VeraSans []byte

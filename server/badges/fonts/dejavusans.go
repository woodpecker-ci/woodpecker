// Copyright 2026 Woodpecker Authors
// SPDX-License-Identifier: MIT.

package fonts

import (
	_ "embed"
)

// DejaVuSans is DejaVuSans.ttf font inlined to the bytes slice.
//
//go:embed DejaVuSans.ttf
var DejaVuSans []byte

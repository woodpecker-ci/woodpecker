// Copyright 2023 The narqo/go-badge Authors. All rights reserved.
// SPDX-License-Identifier: MIT.

package fonts

import (
	_ "embed"
)

// DejaVuSans is DejaVuSans.ttf font inlined to the bytes slice.
//
//go:embed DejaVuSans.ttf
var DejaVuSans []byte

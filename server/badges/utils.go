// Copyright 2026 Woodpecker Authors
// Copyright 2023 The narqo/go-badge Authors. All rights reserved.
// SPDX-License-Identifier: MIT.

package badges

import (
	"regexp"
	"strings"
)

// https://github.com/badges/shields/blob/b6be37d277b64a90bde98ca10446c9d433a56681/badge-maker/lib/xml.js#L7
func stripXMLWhitespace(xml string) string {
	return strings.TrimSpace(regexp.MustCompile(`<\s+`).ReplaceAllString(regexp.MustCompile(`>\s+`).ReplaceAllString(xml, ">"), "<"))
}

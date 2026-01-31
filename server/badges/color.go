// Copyright 2026 Woodpecker Authors
// Copyright 2023 The narqo/go-badge Authors. All rights reserved.
// SPDX-License-Identifier: MIT.

package badges

// Color represents color of the badge.
type Color string

// ColorScheme contains named colors that could be used to render the badge.
var ColorScheme = map[string]string{
	"LightGreen":  "#44cc11",
	"Green":       "#97ca00",
	"Yellow":      "#dfb317",
	"YellowGreen": "#a4a61d",
	"Orange":      "#fe7d37",
	"Red":         "#e05d44",
	"Blue":        "#007ec6",
	"Grey":        "#555555",
	"Gray":        "#555555",
	"LightGrey":   "#9f9f9f",
	"LightGray":   "#9f9f9f",
}

// Standard colors.
const (
	ColorLightGreen  = Color("LightGreen")
	ColorGreen       = Color("Green")
	ColorYellow      = Color("Yellow")
	ColorYellowGreen = Color("YellowGreen")
	ColorOrange      = Color("Orange")
	ColorRed         = Color("Red")
	ColorBlue        = Color("Blue")
	ColorGrey        = Color("Grey")
	ColorGray        = Color("Gray")
	ColorLightGrey   = Color("LightGrey")
	ColorLightGray   = Color("LightGray")
)

func (c Color) String() string {
	color, ok := ColorScheme[string(c)]
	if ok {
		return color
	} else {
		return string(c)
	}
}

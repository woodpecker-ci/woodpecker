// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tui

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
)

// Color palette. Intentionally minimal and terminal-friendly: all
// colors are drawn from the standard 16-color ANSI range so they
// adapt to the user's terminal theme rather than clashing with it.
// A theming pass can come later; v1 stays neutral.
var (
	colorAccent = lipgloss.Color("6") // cyan
	colorMuted  = lipgloss.Color("8") // bright black / gray
)

// selectedRowStyle highlights the tree row under the cursor when
// the tree has focus. Reverse video works across every terminal that
// supports ANSI at all, including ones without truecolor.
var selectedRowStyle = lipgloss.NewStyle().Reverse(true)

// paneTitleStyle renders the short title bar at the top of each
// pane ("logs", "messages"). Reusing an accent foreground + bold
// underline keeps the label distinct from the pane border without
// adding a second color.
var paneTitleStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true).
	Underline(true)

// footerStyle is the keybind hint strip at the bottom of the view.
var footerStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Faint(true)

// paneStyle returns the border style for a pane. Focused panes get
// the accent color; unfocused panes get a muted border so the focus
// indicator is unambiguous without stealing too much attention.
func paneStyle(focused bool) lipgloss.Style {
	color := colorMuted
	if focused {
		color = colorAccent
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)
}

// Status markers. Two render modes: emoji (large and unambiguous on
// modern terminals) and colored squares (a single BMP block tinted
// per state, which stays distinguishable on an old ANSI terminal or
// the Linux console where emoji are unavailable or mis-sized). The
// mode is auto-detected once; see detectGlyphMode.
//
// Emoji are deliberately full-color status faces so pending vs running
// is obvious at a glance — the original "·" vs "●" was too subtle.
const (
	emojiSuccess  = "✅"
	emojiFailure  = "❌"
	emojiSkipped  = "⏭️"
	emojiBlocked  = "⏸️"
	emojiCanceled = "✖️"
	emojiRunning  = "🔵"
	emojiPending  = "⚪"
)

// squareGlyph is the fallback marker; color carries the meaning.
const squareGlyph = "■"

var (
	colorSuccess  = lipgloss.Color("2") // green
	colorFailure  = lipgloss.Color("1") // red
	colorSkipped  = lipgloss.Color("8") // gray
	colorBlocked  = lipgloss.Color("3") // yellow
	colorCanceled = lipgloss.Color("5") // magenta
	colorRunning  = lipgloss.Color("4") // blue
	colorPending  = lipgloss.Color("7") // white
)

type glyphMode int

const (
	glyphModeSquares glyphMode = iota
	glyphModeEmoji
)

// activeGlyphMode resolves the render mode once per process.
var activeGlyphMode = sync.OnceValue(detectGlyphMode)

// detectGlyphMode picks emoji vs colored squares. An explicit
// WOODPECKER_EXEC_TUI_GLYPHS=emoji|squares override wins; otherwise it
// auto-detects and falls back to squares whenever emoji are unlikely
// to render correctly.
func detectGlyphMode() glyphMode {
	switch strings.ToLower(os.Getenv("WOODPECKER_EXEC_TUI_GLYPHS")) {
	case "emoji":
		return glyphModeEmoji
	case "squares", "square", "ascii":
		return glyphModeSquares
	}
	if supportsEmoji() {
		return glyphModeEmoji
	}
	return glyphModeSquares
}

// supportsEmoji is a best-effort heuristic: emoji need a UTF-8 locale
// plus a terminal modern enough to render them at width 2. There is no
// reliable query for this, so we infer from the usual environment
// signals and err toward the safe squares fallback.
func supportsEmoji() bool {
	if !localeIsUTF8() {
		return false
	}
	if ct := strings.ToLower(os.Getenv("COLORTERM")); ct == "truecolor" || ct == "24bit" {
		return true
	}
	if os.Getenv("TERM_PROGRAM") != "" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	for _, hint := range []string{"kitty", "alacritty", "wezterm", "ghostty", "contour", "256color"} {
		if strings.Contains(term, hint) {
			return true
		}
	}
	return false
}

// localeIsUTF8 reports whether the active locale is UTF-8, checking the
// POSIX precedence LC_ALL > LC_CTYPE > LANG.
func localeIsUTF8() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := strings.ToLower(os.Getenv(key))
		if v == "" {
			continue
		}
		return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
	}
	return false
}

// glyphFor renders a status marker in the active mode: the emoji as-is,
// or the square tinted with the state color.
func glyphFor(emoji string, c color.Color) string {
	if activeGlyphMode() == glyphModeEmoji {
		return emoji
	}
	return lipgloss.NewStyle().Foreground(c).Render(squareGlyph)
}

// stateGlyph returns the status marker for a workflow state. Used by
// the tree renderer and the placeholder view.
func stateGlyph(s scheduler.State) string {
	switch s {
	case scheduler.StateSuccess:
		return glyphFor(emojiSuccess, colorSuccess)
	case scheduler.StateFailure:
		return glyphFor(emojiFailure, colorFailure)
	case scheduler.StateBlocked:
		return glyphFor(emojiBlocked, colorBlocked)
	case scheduler.StateCanceled:
		return glyphFor(emojiCanceled, colorCanceled)
	case scheduler.StateRunning:
		return glyphFor(emojiRunning, colorRunning)
	}
	return glyphFor(emojiPending, colorPending)
}

// placeholderHeaderWidth is the width of the horizontal rule in the
// skeleton placeholder view. Replaced by lipgloss-aware sizing in
// the full layout (chunk 5).
const placeholderHeaderWidth = 40

// placeholderView is the bare-bones view used until the full tree +
// log + debug layout lands. It renders one line per workflow with
// state glyph, name, and (if running or finished) a short summary.
// Enough to verify the wiring end-to-end without committing to a
// visual design yet.
func placeholderView(m *Model) string {
	var b strings.Builder

	fmt.Fprintln(&b, "Woodpecker exec")
	fmt.Fprintln(&b, strings.Repeat("─", placeholderHeaderWidth))

	for _, wf := range m.workflows {
		fmt.Fprintf(&b, "  %s %s", stateGlyph(wf.state), wf.name)
		if wf.err != nil {
			fmt.Fprintf(&b, "  (%s)", wf.err.Error())
		}
		fmt.Fprintln(&b)

		if wf.expanded {
			for _, s := range wf.steps {
				fmt.Fprintf(&b, "      %s %s\n", stepGlyph(s), s.name)
			}
		}
	}

	fmt.Fprintln(&b)
	if m.canceling {
		fmt.Fprintln(&b, "canceling…")
	}
	if m.done {
		if m.doneErr != nil {
			fmt.Fprintf(&b, "finished with error: %s\n", m.doneErr.Error())
		} else {
			fmt.Fprintln(&b, "finished successfully")
		}
	}
	fmt.Fprintln(&b, "q / ctrl-c: quit")

	return b.String()
}

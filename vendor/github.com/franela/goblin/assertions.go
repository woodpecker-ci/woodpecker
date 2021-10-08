package goblin

import (
	"fmt"
	"reflect"
	"strings"
)

// Assertion represents a fact stated about a source object. It contains the
// source object and function to call
type Assertion struct {
	src  interface{}
	fail func(interface{})
}

func objectsAreEqual(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	if reflect.DeepEqual(a, b) {
		return true
	}

	if fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b) {
		return true
	}

	return false
}

// Format series of messages provided to an assertion.  Separate messages from
// the preamble of assertion with a comma and concatenate messages using spaces.
// Messages that are purely whitespace will be wrapped with square brackets, so
// the developer can glean that something was actually reported in a message.
func formatMessages(messages ...interface{}) string {
	// Concatenate messages together.
	var fm strings.Builder
	for _, message := range messages {
		fm.WriteString(" ")

		// Format message then wrap with square brackets if only
		// whitespace.
		m := fmt.Sprintf("%v", message)
		if strings.TrimSpace(m) == "" {
			m = fmt.Sprintf("[%s]", m)
		}
		fm.WriteString(m)
	}

	if fm.Len() == 0 {
		return ""
	}
	return "," + fm.String()
}

// Eql is a shorthand alias of Equal for convenience
func (a *Assertion) Eql(dst interface{}, messages ...interface{}) {
	a.Equal(dst, messages)
}

// Equal takes a destination object and asserts that a source object and
// destination object are equal to one another. It will fail the assertion and
// print a corresponding message if the objects are not equivalent.
func (a *Assertion) Equal(dst interface{}, messages ...interface{}) {
	if !objectsAreEqual(a.src, dst) {
		a.fail(fmt.Sprintf("%#v %s %#v%s", a.src, "does not equal", dst,
			formatMessages(messages...)))
	}
}

// IsTrue asserts that a source is equal to true. Optional messages can be
// provided for inclusion in the displayed message if the assertion fails. It
// will fail the assertion if the source does not resolve to true.
func (a *Assertion) IsTrue(messages ...interface{}) {
	if !objectsAreEqual(a.src, true) {
		a.fail(fmt.Sprintf("%v %s%s", a.src,
			"expected false to be truthy",
			formatMessages(messages...)))
	}
}

// IsFalse asserts that a source is equal to false. Optional messages can be
// provided for inclusion in the displayed message if the assertion fails. It
// will fail the assertion if the source does not resolve to false.
func (a *Assertion) IsFalse(messages ...interface{}) {
	if !objectsAreEqual(a.src, false) {
		a.fail(fmt.Sprintf("%v %s%s", a.src,
			"expected true to be falsey",
			formatMessages(messages...)))
	}
}

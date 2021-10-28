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

// isNil returns whether a.src is nil or not.
func (a *Assertion) isNil() bool {
	if !objectsAreEqual(a.src, nil) {
		specialKinds := []reflect.Kind{
			reflect.Slice, reflect.Chan,
			reflect.Map, reflect.Ptr,
			reflect.Interface, reflect.Func,
		}
		t := reflect.TypeOf(a.src).Kind()
		for _, kind := range specialKinds {
			if t == kind {
				return reflect.ValueOf(a.src).IsNil()
			}
		}
		return false
	}
	return true
}

// IsNil asserts that source is nil.
func (a *Assertion) IsNil(messages ...interface{}) {
	if !a.isNil() {
		message := fmt.Sprintf("%v %s%v", a.src, "expected to be nil", formatMessages(messages...))
		a.fail(message)
	}
}

// IsNotNil asserts that source is not nil.
func (a *Assertion) IsNotNil(messages ...interface{}) {
	if a.isNil() {
		message := fmt.Sprintf("%v %s%v", a.src, "is nil", formatMessages(messages...))
		a.fail(message)
	}
}

// IsZero asserts that source is a zero value for its respective type.
// If it is a structure, for example, all of its fields must have their
// respective zero value: "" for strings, 0 for int, etc. Slices, arrays
// and maps are only considered zero if they are nil. To check if these
// type of values are empty or not, use the len() from the data source
// with IsZero(). Example: g.Assert(len(list)).IsZero().
func (a *Assertion) IsZero(messages ...interface{}) {
	valueOf := reflect.ValueOf(a.src)

	if !valueOf.IsZero() {
		message := fmt.Sprintf("%#v %s%v", a.src, "is not a zero value", formatMessages(messages...))

		a.fail(message)
	}
}

// IsNotZero asserts the contrary of IsZero.
func (a *Assertion) IsNotZero(messages ...interface{}) {
	valueOf := reflect.ValueOf(a.src)

	if valueOf.IsZero() {
		message := fmt.Sprintf("%#v %s%v", a.src, "is a zero value", formatMessages(messages...))
		a.fail(message)
	}
}

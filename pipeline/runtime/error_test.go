package runtime

import (
	"testing"
)

func TestExitError(t *testing.T) {
	err := ExitError{
		Name: "build",
		Code: 255,
	}
	got, want := err.Error(), "build : exit code 255"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

func TestOomError(t *testing.T) {
	err := OomError{
		Name: "build",
	}
	got, want := err.Error(), "build : received oom kill"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

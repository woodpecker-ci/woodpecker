package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/pipeline/errors"
)

func TestGetPipelineErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title    string
		err      error
		expected []*pipeline_errors.PipelineError
	}{
		{
			title:    "nil error",
			err:      nil,
			expected: nil,
		},
		{
			title: "warning",
			err: &pipeline_errors.PipelineError{
				IsWarning: true,
			},
			expected: []*pipeline_errors.PipelineError{
				{
					IsWarning: true,
				},
			},
		},
		{
			title: "pipeline error",
			err: &pipeline_errors.PipelineError{
				IsWarning: false,
			},
			expected: []*pipeline_errors.PipelineError{
				{
					IsWarning: false,
				},
			},
		},
		{
			title: "multiple warnings",
			err: multierr.Combine(
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
			),
			expected: []*pipeline_errors.PipelineError{
				{
					IsWarning: true,
				},
				{
					IsWarning: true,
				},
			},
		},
		{
			title: "multiple errors and warnings",
			err: multierr.Combine(
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
				&pipeline_errors.PipelineError{
					IsWarning: false,
				},
				errors.New("some error"),
			),
			expected: []*pipeline_errors.PipelineError{
				{
					IsWarning: true,
				},
				{
					IsWarning: false,
				},
				{
					Type:      pipeline_errors.PipelineErrorTypeGeneric,
					IsWarning: false,
					Message:   "some error",
				},
			},
		},
	}

	for _, test := range tests {
		assert.Equalf(t, pipeline_errors.GetPipelineErrors(test.err), test.expected, test.title)
	}
}

func TestHasBlockingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title    string
		err      error
		expected bool
	}{
		{
			title:    "nil error",
			err:      nil,
			expected: false,
		},
		{
			title: "warning",
			err: &pipeline_errors.PipelineError{
				IsWarning: true,
			},
			expected: false,
		},
		{
			title: "pipeline error",
			err: &pipeline_errors.PipelineError{
				IsWarning: false,
			},
			expected: true,
		},
		{
			title: "multiple warnings",
			err: multierr.Combine(
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
			),
			expected: false,
		},
		{
			title: "multiple errors and warnings",
			err: multierr.Combine(
				&pipeline_errors.PipelineError{
					IsWarning: true,
				},
				&pipeline_errors.PipelineError{
					IsWarning: false,
				},
				errors.New("some error"),
			),
			expected: true,
		},
	}

	for _, test := range tests {
		if pipeline_errors.HasBlockingErrors(test.err) != test.expected {
			t.Error("Should only return true if there are blocking errors")
		}
	}
}

func TestExitError(t *testing.T) {
	err := pipeline_errors.ExitError{
		Name: "build",
		Code: 255,
	}
	got, want := err.Error(), "build : exit code 255"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

func TestOomError(t *testing.T) {
	err := pipeline_errors.OomError{
		Name: "build",
	}
	got, want := err.Error(), "build : received oom kill"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

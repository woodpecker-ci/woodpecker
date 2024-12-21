package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors/types"
)

func TestGetPipelineErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title    string
		err      error
		expected []*types.PipelineError
	}{
		{
			title:    "nil error",
			err:      nil,
			expected: nil,
		},
		{
			title: "warning",
			err: &types.PipelineError{
				IsWarning: true,
			},
			expected: []*types.PipelineError{
				{
					IsWarning: true,
				},
			},
		},
		{
			title: "pipeline error",
			err: &types.PipelineError{
				IsWarning: false,
			},
			expected: []*types.PipelineError{
				{
					IsWarning: false,
				},
			},
		},
		{
			title: "multiple warnings",
			err: multierr.Combine(
				&types.PipelineError{
					IsWarning: true,
				},
				&types.PipelineError{
					IsWarning: true,
				},
			),
			expected: []*types.PipelineError{
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
				&types.PipelineError{
					IsWarning: true,
				},
				&types.PipelineError{
					IsWarning: false,
				},
				errors.New("some error"),
			),
			expected: []*types.PipelineError{
				{
					IsWarning: true,
				},
				{
					IsWarning: false,
				},
				{
					Type:      types.PipelineErrorTypeGeneric,
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
			err: &types.PipelineError{
				IsWarning: true,
			},
			expected: false,
		},
		{
			title: "pipeline error",
			err: &types.PipelineError{
				IsWarning: false,
			},
			expected: true,
		},
		{
			title: "multiple warnings",
			err: multierr.Combine(
				&types.PipelineError{
					IsWarning: true,
				},
				&types.PipelineError{
					IsWarning: true,
				},
			),
			expected: false,
		},
		{
			title: "multiple errors and warnings",
			err: multierr.Combine(
				&types.PipelineError{
					IsWarning: true,
				},
				&types.PipelineError{
					IsWarning: false,
				},
				errors.New("some error"),
			),
			expected: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, pipeline_errors.HasBlockingErrors(test.err))
	}
}

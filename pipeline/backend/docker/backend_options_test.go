package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func Test_parseBackendOptions(t *testing.T) {
	tests := []struct {
		name    string
		step    *backend.Step
		want    BackendOptions
		wantErr bool
	}{
		{
			name: "nil options",
			step: &backend.Step{BackendOptions: nil},
			want: BackendOptions{},
		},
		{
			name: "empty options",
			step: &backend.Step{BackendOptions: map[string]any{}},
			want: BackendOptions{},
		},
		{
			name: "with user option",
			step: &backend.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"user": "1000:1000",
				},
			}},
			want: BackendOptions{User: "1000:1000"},
		},
		{
			name:    "invalid backend options",
			step:    &backend.Step{BackendOptions: map[string]any{"docker": "invalid"}},
			want:    BackendOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBackendOptions(tt.step)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

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

package kubernetes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestToDnsName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "underscores to dashes",
			in:   "wp_01he8bebctabr3kgk0qj36d2me_0_services_0",
			want: "wp-01he8bebctabr3kgk0qj36d2me-0-services-0",
		},
		{
			name: "mixed case with dots and dashes",
			in:   "a.0-AA",
			want: "a.0-aa",
		},
		{
			name: "long valid fqdn unchanged",
			in:   "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local",
			want: "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local",
		},
		{
			name: "uppercase with underscores",
			in:   "BUILD_AND_DEPLOY_0",
			want: "build-and-deploy-0",
		},
		{
			name: "spaces to dashes",
			in:   "build and deploy",
			want: "build-and-deploy",
		},
		{
			name: "special char ampersand",
			in:   "build & deploy",
			want: "build-deploy",
		},
		{
			name: "backslash to dash",
			in:   "abc\\def",
			want: "abc-def",
		},
		{
			name: "leading dash trimmed",
			in:   "-build-and-deploy",
			want: "build-and-deploy",
		},
		{
			name: "trailing dash trimmed",
			in:   "test-",
			want: "test",
		},
		{
			name: "leading dot trimmed",
			in:   ".0-a",
			want: "0-a",
		},
		{
			name: "consecutive dots collapsed",
			in:   "ABC..DEF",
			want: "abc.def",
		},
		{
			name: "dot-dash collapsed",
			in:   "0.-a",
			want: "0.a",
		},
		{
			name: "trailing dot trimmed",
			in:   "0-a.",
			want: "0-a",
		},
		{
			name: "consecutive underscores single dash",
			in:   "a__b",
			want: "a-b",
		},
		{
			name: "mixed dots and dashes",
			in:   "a.-.-b",
			want: "a.b",
		},
		{
			name: "unicode emoji replaced",
			in:   "hello🚀world",
			want: "hello-world",
		},
		{
			name: "single valid char",
			in:   "a",
			want: "a",
		},
		{
			name: "numbers only",
			in:   "123",
			want: "123",
		},
		{
			name: "mixed leading trailing",
			in:   "-.-test-.-",
			want: "test",
		},
		{
			name:    "all special chars empty",
			in:      "!!!",
			wantErr: true,
		},
		{
			name:    "all dots empty",
			in:      "...",
			wantErr: true,
		},
		{
			name:    "all dashes empty",
			in:      "---",
			wantErr: true,
		},
		{
			name:    "empty string",
			in:      "",
			wantErr: true,
		},
		{
			name: "truncation needed",
			in:   strings.Repeat("a", 300),
		},
		{
			name: "truncation boundary exact 253",
			in:   strings.Repeat("a", 253),
			want: strings.Repeat("a", 253),
		},
		{
			name: "truncation with trailing dash at cut",
			in:   strings.Repeat("a", 240) + "----" + strings.Repeat("b", 60),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toDNSName(tt.in)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrDNSPatternInvalid)
				return
			}
			assert.NoError(t, err)
			if tt.name == "truncation needed" || tt.name == "truncation with trailing dash at cut" {
				assert.LessOrEqual(t, len(got), validation.DNS1123SubdomainMaxLength)
				assert.Len(t, validation.IsDNS1123Subdomain(got), 0)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToDnsNameTruncationUniqueness(t *testing.T) {
	long1 := strings.Repeat("a", 300)
	long2 := strings.Repeat("a", 299) + "b"

	got1, err := toDNSName(long1)
	assert.NoError(t, err)

	got2, err := toDNSName(long2)
	assert.NoError(t, err)

	assert.NotEqual(t, got1, got2, "different long inputs should produce different truncated names")
}

func TestToLabelValue(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "underscores preserved",
			in:   "wp_01he8bebctabr3kgk0qj36d2me_0_services_0",
			want: "wp_01he8bebctabr3kgk0qj36d2me_0_services_0",
		},
		{
			name: "dots preserved",
			in:   "a.0-AA",
			want: "a.0-aa",
		},
		{
			name: "uppercase with underscores",
			in:   "BUILD_AND_DEPLOY_0",
			want: "build_and_deploy_0",
		},
		{
			name: "spaces to dashes",
			in:   "build and deploy",
			want: "build-and-deploy",
		},
		{
			name: "special char ampersand",
			in:   "build & deploy",
			want: "build-deploy",
		},
		{
			name: "backslash to dash",
			in:   "abc\\def",
			want: "abc-def",
		},
		{
			name: "leading dash trimmed",
			in:   "-build-and-deploy",
			want: "build-and-deploy",
		},
		{
			name: "trailing dash trimmed",
			in:   "test-",
			want: "test",
		},
		{
			name: "leading dot trimmed",
			in:   ".0-a",
			want: "0-a",
		},
		{
			name: "trailing dot trimmed",
			in:   "0-a.",
			want: "0-a",
		},
		{
			name: "consecutive underscores collapsed",
			in:   "a__b",
			want: "a-b",
		},
		{
			name: "consecutive dots collapsed",
			in:   "ABC..DEF",
			want: "abc-def",
		},
		{
			name: "mixed separators collapsed",
			in:   "a._-b",
			want: "a-b",
		},
		{
			name: "unicode emoji replaced",
			in:   "hello🚀world",
			want: "hello-world",
		},
		{
			name: "single valid char",
			in:   "a",
			want: "a",
		},
		{
			name: "numbers only",
			in:   "123",
			want: "123",
		},
		{
			name: "mixed leading trailing",
			in:   "-.-test-.-",
			want: "test",
		},
		{
			name: "all special chars become empty",
			in:   "!!!",
			want: "",
		},
		{
			name: "empty string is valid",
			in:   "",
			want: "",
		},
		{
			name: "truncation needed",
			in:   strings.Repeat("a", 100),
		},
		{
			name: "truncation boundary exact 63",
			in:   strings.Repeat("a", 63),
			want: strings.Repeat("a", 63),
		},
		{
			name: "truncation with trailing dash at cut",
			in:   strings.Repeat("a", 50) + "----" + strings.Repeat("b", 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toLabelValue(tt.in)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrLabelInvalid)
				return
			}
			assert.NoError(t, err)
			if tt.name == "truncation needed" || tt.name == "truncation with trailing dash at cut" {
				assert.LessOrEqual(t, len(got), validation.LabelValueMaxLength)
				assert.Len(t, validation.IsValidLabelValue(got), 0)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToLabelValueTruncationUniqueness(t *testing.T) {
	long1 := strings.Repeat("a", 100)
	long2 := strings.Repeat("a", 99) + "b"

	got1, err := toLabelValue(long1)
	assert.NoError(t, err)

	got2, err := toLabelValue(long2)
	assert.NoError(t, err)

	assert.NotEqual(t, got1, got2, "different long inputs should produce different truncated labels")
}

func TestGetHostnameOrEmpty(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "spaces to dashes",
			in:   "Update repos",
			want: "update-repos",
		},
		{
			name: "underscores to dashes",
			in:   "MY_STEP",
			want: "my-step",
		},
		{
			name: "emoji removed",
			in:   "Build 🚀",
			want: "build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHostnameOrEmpty(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

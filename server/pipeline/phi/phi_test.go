// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package phi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeLogData(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     string
		wantDetected bool
	}{
		{
			name:         "plain build log without PHI",
			input:        "Compiling clinical module package version 1.0.4\n",
			expected:     "Compiling clinical module package version 1.0.4\n",
			wantDetected: false,
		},
		{
			name:         "BulletTrain Client Registry CR-ID",
			input:        "Querying patient payload for CR-000123 from client registry\n",
			expected:     "Querying patient payload for [REDACTED_CLIENT_REGISTRY_PATIENT_ID] from client registry\n",
			wantDetected: true,
		},
		{
			name:         "BulletTrain Client Registry MPI-ID",
			input:        "Dispatching appointment notification for MPI-102938\n",
			expected:     "Dispatching appointment notification for [REDACTED_CLIENT_REGISTRY_PATIENT_ID]\n",
			wantDetected: true,
		},
		{
			name:         "BulletTrain BT-PAT ID",
			input:        "Encounter created for BT-PAT-5544332 at ward handoff\n",
			expected:     "Encounter created for [REDACTED_CLIENT_REGISTRY_PATIENT_ID] at ward handoff\n",
			wantDetected: true,
		},
		{
			name:         "Valid Modulus 11 National Health Identifier with spaces",
			input:        "Patient identifier: 943 476 5919 assigned\n",
			expected:     "Patient identifier: [REDACTED_NATIONAL_HEALTH_ID] assigned\n",
			wantDetected: true,
		},
		{
			name:         "Valid Modulus 11 National Health Identifier without spaces",
			input:        "Processing encounter record 9434765919\n",
			expected:     "Processing encounter record [REDACTED_NATIONAL_HEALTH_ID]\n",
			wantDetected: true,
		},
		{
			name:         "Invalid 10-digit number (random phone or build hash)",
			input:        "Timestamp unix seconds: 1729785600 build artifact\n",
			expected:     "Timestamp unix seconds: 1729785600 build artifact\n",
			wantDetected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, detected := SanitizeLogData([]byte(tc.input))
			assert.Equal(t, tc.expected, string(output))
			assert.Equal(t, tc.wantDetected, detected)
		})
	}
}

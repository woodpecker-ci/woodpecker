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
	"regexp"
	"strconv"
	"strings"
)

var (
	// clientRegistryIDRegex matches BulletTrain Client Registry Patient IDs (e.g. CR-000001, MPI-102938, BT-PAT-9921)
	clientRegistryIDRegex = regexp.MustCompile(`\b(CR|MPI|BT-PAT|BT)-[0-9]{5,12}\b`)

	// nationalHealthIDCandidateRegex matches candidate 10-digit national patient health identifiers with optional spaces or dashes
	nationalHealthIDCandidateRegex = regexp.MustCompile(`\b\d{3}[ -]?\d{3}[ -]?\d{4}\b`)
)

const (
	RedactedPatientID        = "[REDACTED_CLIENT_REGISTRY_PATIENT_ID]"
	RedactedNationalHealthID = "[REDACTED_NATIONAL_HEALTH_ID]"
)

// SanitizeLogData scans a log line, redacting client registry patient IDs and valid Modulus 11 national health identifiers.
func SanitizeLogData(input []byte) ([]byte, bool) {
	str := string(input)
	detected := false

	// 1. Redact BulletTrain Client Registry Patient IDs
	if clientRegistryIDRegex.MatchString(str) {
		str = clientRegistryIDRegex.ReplaceAllString(str, RedactedPatientID)
		detected = true
	}

	// 2. Redact verified National Health IDs (Modulus 11 validated)
	if nationalHealthIDCandidateRegex.MatchString(str) {
		cleaned := nationalHealthIDCandidateRegex.ReplaceAllStringFunc(str, func(match string) string {
			digits := extractDigits(match)
			if len(digits) == 10 && isValidNationalHealthID(digits) {
				detected = true
				return RedactedNationalHealthID
			}
			return match
		})
		str = cleaned
	}

	return []byte(str), detected
}

func extractDigits(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// isValidNationalHealthID executes the standard Modulus 11 national health identifier algorithm.
func isValidNationalHealthID(digits string) bool {
	if len(digits) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		d, err := strconv.Atoi(string(digits[i]))
		if err != nil {
			return false
		}
		weight := 11 - (i + 1) // 10, 9, 8, 7, 6, 5, 4, 3, 2
		sum += d * weight
	}

	remainder := sum % 11
	checkDigit := 11 - remainder
	if checkDigit == 11 {
		checkDigit = 0
	} else if checkDigit == 10 {
		// 10 is reserved/invalid under Modulus 11 standard
		return false
	}

	expectedCheckDigit, err := strconv.Atoi(string(digits[9]))
	if err != nil {
		return false
	}

	return checkDigit == expectedCheckDigit
}

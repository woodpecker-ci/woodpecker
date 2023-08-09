// Copyright 2022 Woodpecker Authors
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

package utils

// DedupStrings deduplicate string list, empty items are dropped
func DedupStrings(src []string) []string {
	m := make(map[string]struct{}, len(src))
	dst := make([]string, 0, len(src))

	for _, v := range src {
		// Skip empty items
		if len(v) == 0 {
			continue
		}
		// Skip duplicates
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		dst = append(dst, v)
	}

	return dst
}

// EqualSliceValues compare two slices if they have equal values independent of how they are sorted
func EqualSliceValues[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}

	m1 := sliceToCountMap(s1)
	m2 := sliceToCountMap(s2)

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}

func sliceToCountMap[E comparable](list []E) map[E]int {
	m := make(map[E]int)
	for i := range list {
		m[list[i]]++
	}
	return m
}

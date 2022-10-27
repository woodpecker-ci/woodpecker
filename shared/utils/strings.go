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
func DedupStrings(list []string) []string {
	m := make(map[string]struct{}, len(list))

	for i := range list {
		if s := list[i]; len(s) > 0 {
			m[list[i]] = struct{}{}
		}
	}

	newList := make([]string, 0, len(m))
	for k := range m {
		newList = append(newList, k)
	}
	return newList
}

// EqualStringSlice compare two string slices if they have equal values independent of how they are sorted
func EqualStringSlice(l1, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}

	m1 := sliceToCountMap(l1)
	m2 := sliceToCountMap(l2)

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}

func sliceToCountMap(list []string) map[string]int {
	m := make(map[string]int)
	for i := range list {
		m[list[i]]++
	}
	return m
}

func SliceContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}

	return false
}

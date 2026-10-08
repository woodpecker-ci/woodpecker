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

package cron

import (
	"testing"
	"time"
)

// FuzzCalcNewNext feeds untrusted, user supplied cron schedule and timezone
// strings into the schedule parser. The property checked is that parsing
// never panics and that a successfully calculated execution time is never in
// the past.
func FuzzCalcNewNext(f *testing.F) {
	f.Add("@daily", "")
	f.Add("*/5 * * * *", "Europe/Berlin")
	f.Add("0 0 1 1 *", "UTC")
	f.Add("60 25 * * *", "Not/AZone")

	now := time.Unix(1257894000, 0)

	f.Fuzz(func(t *testing.T, schedule, tzLoc string) {
		next, err := CalcNewNext(schedule, tzLoc, now)
		if err != nil {
			return
		}
		if next.Before(now) {
			t.Fatalf("next execution %v is before now %v for schedule %q tz %q", next, now, schedule, tzLoc)
		}
	})
}

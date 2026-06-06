package main

import (
	"testing"
	"time"
)

// TestTimezoneDatabaseAvailable verifies that the embedded IANA timezone
// database is available. This ensures cron schedules work in minimal
// containers (Alpine, scratch) that lack system timezone data.
// Regression test for https://github.com/woodpecker-ci/woodpecker/issues/6688
func TestTimezoneDatabaseAvailable(t *testing.T) {
	timezones := []string{
		"Europe/Berlin",
		"America/New_York",
		"Asia/Tokyo",
		"Australia/Sydney",
		"UTC",
	}

	for _, tz := range timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Errorf("time.LoadLocation(%q) failed: %v", tz, err)
			continue
		}
		if loc == nil {
			t.Errorf("time.LoadLocation(%q) returned nil location", tz)
		}
	}
}

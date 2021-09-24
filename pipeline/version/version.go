package version

import "fmt"

var (
	// major is for an API incompatible changes
	major = 1
	// minor is for functionality in a backwards-compatible manner
	minor = 0
	// patch is for backwards-compatible bug fixes
	patch = 0
)

// String returns the supporeted specification versions in string format.
func String() string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

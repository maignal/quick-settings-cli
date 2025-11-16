// This file is in the 'bluetooth' package
package bluetooth

import (
	"fmt"
	"regexp"
)

// IMPORTANT: The function name must start with a
// capital letter to be "Exported" (public).
func ParsePairedStatus(output string) (bool, error) {
	re := regexp.MustCompile(`(?m)^\s*Connected:\s+(.+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 2 {
		return false, fmt.Errorf("could not find 'Paired' status")
	}

	return matches[1] == "yes", nil
}

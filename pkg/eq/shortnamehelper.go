// Package eq provides EverQuest archive extraction and processing utilities.
package eq

import (
	"regexp"
	"strings"
	"unicode"
)

// GetCorrectZoneShortname normalizes a zone shortname by removing trailing digits
// unless it's a known exception zone that uses digits as part of its name.
//
// Known exceptions that retain their digits:
//   - qeynos2: South Qeynos zone
//   - qey2hh1: Qeynos Hills connection zone
//   - global*: Global resource archives
func GetCorrectZoneShortname(shortName string) string {
	if shortName == "" {
		return shortName
	}

	// Check if the name ends with a digit
	lastChar := rune(shortName[len(shortName)-1])
	if !unicode.IsDigit(lastChar) {
		return shortName
	}

	// Check for known exceptions that should keep their digits
	if shortName == "qeynos2" || shortName == "qey2hh1" || strings.HasPrefix(shortName, "global") {
		return shortName
	}

	// Remove trailing digits and dashes
	re := regexp.MustCompile(`[\d-]+$`)
	return re.ReplaceAllString(shortName, "")
}

package infrastructure

import (
	"strings"
)

// ParseTextByNewline parses a string by newlines, pruning empty lines and comment lines.
// The commentChar parameter specifies which character denotes a comment line.
func ParseTextByNewline(text string, commentChar rune) []string {
	if text == "" {
		return nil
	}

	// Split by various newline formats
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	var result []string
	commentPrefix := string(commentChar)

	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, commentPrefix) {
			continue
		}
		result = append(result, line)
	}

	return result
}

// ParseTextByEmptyLines parses text using empty lines as delimiters.
// Returns blocks of text separated by one or more empty lines.
func ParseTextByEmptyLines(text string, commentChar rune) []string {
	if text == "" {
		return nil
	}

	// Normalize newlines
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Split by double newlines (empty lines)
	blocks := strings.Split(text, "\n\n")

	var result []string
	commentPrefix := string(commentChar)

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if strings.HasPrefix(block, commentPrefix) {
			continue
		}
		result = append(result, block)
	}

	return result
}

// ParseTextByDelimitedLines first parses the text into lines, then splits each line by a delimiter.
// Returns a slice of slices, where each inner slice contains the parts of a single line.
func ParseTextByDelimitedLines(text string, delimiter, commentChar rune) [][]string {
	if text == "" {
		return nil
	}

	lines := ParseTextByNewline(text, commentChar)
	if lines == nil {
		return nil
	}

	var result [][]string
	delimStr := string(delimiter)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, delimStr)
		result = append(result, parts)
	}

	return result
}

// ParseTextToDictionary parses lines with key-value pairs separated by a delimiter into a map.
// Useful for settings files (e.g., "PlayerHealth = 10").
// Lines that don't have exactly two parts after splitting are skipped.
func ParseTextToDictionary(text string, delimiter, commentChar rune) map[string]string {
	if text == "" {
		return nil
	}

	lines := ParseTextByNewline(text, commentChar)
	if lines == nil {
		return nil
	}

	result := make(map[string]string)
	delimStr := string(delimiter)

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, delimStr)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		result[key] = value
	}

	return result
}

// ParseStringToList splits a string by semicolons into a list.
func ParseStringToList(text string) []string {
	if text == "" {
		return []string{}
	}

	return strings.Split(text, ";")
}

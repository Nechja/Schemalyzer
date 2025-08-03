package models

import (
	"regexp"
	"strings"
)

// IgnorePattern represents a pattern to ignore during schema comparison
type IgnorePattern struct {
	Pattern    string
	ObjectType string // "table", "column", "constraint", "index", "view", "sequence", "procedure", "function", "trigger", or "*" for all
	Regex      *regexp.Regexp
}

// IgnoreConfig holds all ignore patterns
type IgnoreConfig struct {
	Patterns []IgnorePattern
}

// NewIgnoreConfig creates a new ignore configuration from pattern strings
func NewIgnoreConfig(patterns []string) (*IgnoreConfig, error) {
	config := &IgnoreConfig{
		Patterns: make([]IgnorePattern, 0, len(patterns)),
	}

	for _, pattern := range patterns {
		parts := strings.SplitN(pattern, ":", 2)
		var objectType, patternStr string
		
		if len(parts) == 2 {
			objectType = strings.ToLower(parts[0])
			patternStr = parts[1]
		} else {
			objectType = "*"
			patternStr = pattern
		}

		// Convert simple wildcards to regex
		regexPattern := convertWildcardToRegex(patternStr)
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, err
		}

		config.Patterns = append(config.Patterns, IgnorePattern{
			Pattern:    patternStr,
			ObjectType: objectType,
			Regex:      regex,
		})
	}

	return config, nil
}

// ShouldIgnore checks if an object should be ignored based on the patterns
func (ic *IgnoreConfig) ShouldIgnore(objectType, objectName string) bool {
	objectType = strings.ToLower(objectType)
	
	for _, pattern := range ic.Patterns {
		// Check if pattern applies to this object type
		if pattern.ObjectType != "*" && pattern.ObjectType != objectType {
			continue
		}

		// Check if name matches pattern
		if pattern.Regex.MatchString(objectName) {
			return true
		}
	}

	return false
}

// convertWildcardToRegex converts simple wildcard patterns to regex
func convertWildcardToRegex(pattern string) string {
	// Escape special regex characters except * and ?
	pattern = regexp.QuoteMeta(pattern)
	// Convert escaped wildcards back
	pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
	pattern = strings.ReplaceAll(pattern, `\?`, `.`)
	// Anchor the pattern
	return "^" + pattern + "$"
}
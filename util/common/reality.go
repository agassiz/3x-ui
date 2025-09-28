package common

import "strings"

// isHexString checks whether s contains only hexadecimal characters.
func isHexString(s string) bool {
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			continue
		}
		return false
	}
	return true
}

// NormalizeRealityShortIDsFromAny converts a heterogeneous slice into a slice of
// unique, lowercase short IDs that satisfy Reality's length and hex requirements.
func NormalizeRealityShortIDsFromAny(values []any) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	seen := make(map[string]struct{})

	for _, value := range values {
		str, ok := value.(string)
		if !ok {
			continue
		}

		id := strings.TrimSpace(str)
		length := len(id)
		if length < 8 || length > 16 || length%2 != 0 {
			continue
		}
		if !isHexString(id) {
			continue
		}

		normalized := strings.ToLower(id)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

// FirstRealityShortIDFromAny returns the first valid short ID from the provided slice.
func FirstRealityShortIDFromAny(values []any) (string, bool) {
	normalized := NormalizeRealityShortIDsFromAny(values)
	if len(normalized) == 0 {
		return "", false
	}
	return normalized[0], true
}

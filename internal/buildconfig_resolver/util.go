package buildconfig_resolver

import "strings"

func normalizeKey(k string) string {
	result := make([]rune, 0, len(k))
	for _, c := range k {
		if c == '-' {
			result = append(result, '_')
		} else {
			result = append(result, c)
		}
	}
	return strings.ToUpper(string(result))
}

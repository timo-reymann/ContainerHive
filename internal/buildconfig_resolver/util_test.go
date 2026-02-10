package buildconfig_resolver

import "testing"

func TestNormalizeKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello-world", "HELLO_WORLD"},
		{"test-key", "TEST_KEY"},
		{"already_upper", "ALREADY_UPPER"},
		{"mixed-Case-Key", "MIXED_CASE_KEY"},
		{"no-hyphens", "NO_HYPHENS"},
		{"", ""},
		{"single", "SINGLE"},
		{"a-b-c-d", "A_B_C_D"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeKey(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeKey(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

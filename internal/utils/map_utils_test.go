package utils

import "testing"

func TestMergeMapWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		existing map[string]string
		add      map[string]string
		expected map[string]string
	}{
		{
			name:     "empty add map",
			prefix:   "prefix_",
			existing: map[string]string{"key1": "value1"},
			add:      map[string]string{},
			expected: map[string]string{"key1": "value1"},
		},
		{
			name:     "nil add map",
			prefix:   "prefix_",
			existing: map[string]string{"key1": "value1"},
			add:      nil,
			expected: map[string]string{"key1": "value1"},
		},
		{
			name:     "empty existing map",
			prefix:   "prefix_",
			existing: map[string]string{},
			add:      map[string]string{"key1": "value1"},
			expected: map[string]string{"prefix_key1": "value1"},
		},
		{
			name:     "both maps populated",
			prefix:   "prefix_",
			existing: map[string]string{"key1": "value1"},
			add:      map[string]string{"key2": "value2", "key3": "value3"},
			expected: map[string]string{"key1": "value1", "prefix_key2": "value2", "prefix_key3": "value3"},
		},
		{
			name:     "empty prefix",
			prefix:   "",
			existing: map[string]string{"key1": "value1"},
			add:      map[string]string{"key2": "value2"},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "overwrite existing key with prefix",
			prefix:   "prefix_",
			existing: map[string]string{"prefix_key1": "old"},
			add:      map[string]string{"key1": "new"},
			expected: map[string]string{"prefix_key1": "new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeMapWithPrefix(tt.prefix, tt.existing, tt.add)

			if len(tt.existing) != len(tt.expected) {
				t.Errorf("expected map length %d, got %d", len(tt.expected), len(tt.existing))
			}

			for k, v := range tt.expected {
				if got, ok := tt.existing[k]; !ok {
					t.Errorf("expected key %q not found", k)
				} else if got != v {
					t.Errorf("for key %q: expected %q, got %q", k, v, got)
				}
			}
		})
	}
}

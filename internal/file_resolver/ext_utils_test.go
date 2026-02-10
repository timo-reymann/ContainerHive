package file_resolver

import "testing"

func TestRemoveTemplateExt(t *testing.T) {
	tests := map[string]struct {
		filename string
		expected string
	}{
		"removes gotpl extension": {
			filename: "Dockerfile.gotpl",
			expected: "Dockerfile",
		},
		"removes gotpl extension with multiple dots": {
			filename: "config.yaml.gotpl",
			expected: "config.yaml",
		},
		"keeps filename without extension": {
			filename: "Dockerfile",
			expected: "Dockerfile",
		},
		"keeps filename with unsupported extension": {
			filename: "config.txt",
			expected: "config.txt",
		},
		"keeps filename with unsupported extension after supported one": {
			filename: "config.gotpl.txt",
			expected: "config.gotpl.txt",
		},
		"handles empty string": {
			filename: "",
			expected: "",
		},
		"handles filename with only extension": {
			filename: ".gotpl",
			expected: "",
		},
		"handles hidden file without template extension": {
			filename: ".dockerignore",
			expected: ".dockerignore",
		},
		"handles hidden file with template extension": {
			filename: ".dockerignore.gotpl",
			expected: ".dockerignore",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := RemoveTemplateExt(tc.filename)

			if got != tc.expected {
				t.Errorf("RemoveTemplateExt(%q) = %q, expected %q", tc.filename, got, tc.expected)
			}
		})
	}
}

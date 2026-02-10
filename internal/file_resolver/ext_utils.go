package file_resolver

import "path/filepath"

func RemoveTemplateExt(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) < 1 {
		return filename
	}

	if _, ok := processorMapping[ext[1:]]; ok {
		return filename[:len(filename)-len(ext)]
	}

	return filename
}

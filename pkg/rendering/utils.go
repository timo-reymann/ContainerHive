package rendering

import "os"

func mkdir(targetPath string) error {
	return os.MkdirAll(targetPath, 0755)
}

package rendering

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Get source file info to preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Set permissions on the destination file
	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create a destination directory with source permissions
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
		} else {
			err = copyFile(srcPath, dstPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func copyRootFs(source, targetRoot string) error {
	if err := mkdir(targetRoot); err != nil {
		return errors.Join(errors.New("failed to create target root directory"), err)
	}

	err := copyDir(source, filepath.Join(targetRoot, "rootfs"))
	if err != nil {
		return err
	}
	return nil
}

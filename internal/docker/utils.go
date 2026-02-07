package docker

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func extractTar(tarPath, destDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return errors.Join(errors.New("failed to open tar"), err)
	}
	defer f.Close()

	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Sanitize: clean the name and ensure it stays within destDir
		clean := filepath.Clean(hdr.Name)
		target := filepath.Join(absDestDir, clean)
		if !filepath.IsLocal(clean) {
			return errors.New("tar entry escapes destination: " + hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

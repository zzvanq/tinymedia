package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UpdateFile(r io.Reader, fileName string) error {
	dir := filepath.Dir(fileName)
	base := filepath.Base(fileName)

	tmpFile, err := os.CreateTemp(dir, base)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), fileName); err != nil {
		if errRemove := os.Remove(tmpFile.Name()); errRemove != nil {
			err = errRemove
		}
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

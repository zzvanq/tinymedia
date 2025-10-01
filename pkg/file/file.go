package file

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrUnsupportedFileType = errors.New("unsupported file type")

type FileTypeMagic []byte

var (
	JPEGMagic = FileTypeMagic{0xFF, 0xD8}
	PNGMagic  = FileTypeMagic{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

const MagicPrefixMaxLength = 2

func ReadFileType(r io.Reader) (FileType, error) {
	prefix := make([]byte, MagicPrefixMaxLength)
	if _, err := io.ReadFull(r, prefix); err != nil {
		return "", err
	}

	if bytes.HasPrefix(prefix, JPEGMagic) {
		return FileTypeJPEG, nil
	}
	return "", ErrUnsupportedFileType
}

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

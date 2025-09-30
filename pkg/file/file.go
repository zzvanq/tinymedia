package file

import (
	"bytes"
	"errors"
	"io"
)

var ErrUnsupportedFileType = errors.New("unsupported file type")

type FileTypeMagic []byte

var JPEGMagic = FileTypeMagic{0xff, 0xd8}

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

package file

import (
	"bytes"
	"errors"
	"io"

	"github.com/zzvanq/tinymedia/internal/file/magic"
)

var ErrUnsupportedFileType = errors.New("unsupported file type")

func ReadFileType(r io.Reader) (io.Reader, FileType, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	prefix := make([]byte, magic.MagicPrefixMaxLength)
	if _, err := io.ReadFull(tee, prefix); err != nil {
		return nil, "", err
	}

	rewindReader := io.MultiReader(&buf, r)
	if bytes.HasPrefix(prefix, magic.JPEGMagic) {
		return rewindReader, FileTypeJPEG, nil
	}

	return nil, "", ErrUnsupportedFileType
}

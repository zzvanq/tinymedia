package manager

import (
	"io"

	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
)

type MetaManager interface {
	Insert(vendor codec.MetaCodecVendor, fields map[string]string) error
	Upsert(vendor codec.MetaCodecVendor, fields map[string]string) error
	Extract(vendor codec.MetaCodecVendor, fields ...string) (map[string]string, error)
	FileReader() io.Reader
}

func NewMetaManager(r io.Reader) (MetaManager, error) {
	fileType, err := file.ReadFileType(r)
	if err != nil {
		return nil, err
	}
	switch fileType {
	case file.FileTypeJPEG:
		return NewJpegMetaManager(r), nil
	default:
		return nil, file.ErrUnsupportedFileType
	}
}

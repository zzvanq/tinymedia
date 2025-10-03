package manager

import (
	"io"

	"github.com/zzvanq/tinymedia/internal/meta/manager/jpeg"
	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
)

type MetaManager interface {
	Insert(vendor codec.MetaCodecVendor, fields map[string]string) error
	Upsert(vendor codec.MetaCodecVendor, fields map[string]string) error
	Extract(vendor codec.MetaCodecVendor, fields ...string) (map[string]string, error)
	FileReader() io.Reader
}

func NewMetaManager(r io.Reader, ftype file.FileType) (MetaManager, error) {
	switch ftype {
	case file.FileTypeJPEG:
		return jpeg.NewJpegMetaManager(r)
	default:
		return nil, file.ErrUnsupportedFileType
	}
}

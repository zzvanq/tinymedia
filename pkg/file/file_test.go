package file

import (
	"bytes"
	"errors"
	"testing"

	"github.com/zzvanq/tinymedia/internal/file/magic"
)

func Test_ReadFileType(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    FileType
		wantErr error
	}{
		{
			name:    "jpeg",
			data:    magic.JPEGMagic,
			want:    FileTypeJPEG,
			wantErr: nil,
		},
		{
			name:    "png",
			data:    magic.PNGMagic,
			want:    "",
			wantErr: ErrUnsupportedFileType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := ReadFileType(bytes.NewReader(tt.data))
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("error: %v, want: %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}

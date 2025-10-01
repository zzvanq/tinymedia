package manager

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/manager/jpeg"
)

func Test_NewMetaManager(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    MetaManager
		wantErr error
	}{
		{
			name:    "jpeg",
			r:       bytes.NewReader([]byte{0xFF, 0xD8}),
			want:    &jpeg.JpegMetaManager{},
			wantErr: nil,
		},
		{
			name:    "gif",
			r:       bytes.NewReader([]byte{0x47, 0x49, 0x46, 0x38}),
			want:    nil,
			wantErr: file.ErrUnsupportedFileType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetaManager(tt.r)
			if err != tt.wantErr {
				t.Errorf("want error: %v, got: %v", tt.wantErr, err)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("want: %T, got: %T", tt.want, got)
			}
		})
	}
}

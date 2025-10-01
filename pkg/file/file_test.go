package file

import (
	"bytes"
	"errors"
	"os"
	"testing"
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
			data:    JPEGMagic,
			want:    FileTypeJPEG,
			wantErr: nil,
		},
		{
			name:    "png",
			data:    PNGMagic,
			want:    "",
			wantErr: ErrUnsupportedFileType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFileType(bytes.NewReader(tt.data))
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

func Test_UpdateFile(t *testing.T) {
	file, err := os.CreateTemp("./", "test")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	file.Close()

	want := []byte{0xFF, 0xD8, 0x00, 0x02}
	r := bytes.NewReader(want)
	if err := UpdateFile(r, file.Name()); err != nil {
		t.Errorf("want error: %v, got: %v", nil, err)
	}

	got, err := os.ReadFile(file.Name())
	if err != nil {
		t.Errorf("want error: %v, got: %v", nil, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("want: %v, got: %v", want, got)
	}
	defer file.Close()
}

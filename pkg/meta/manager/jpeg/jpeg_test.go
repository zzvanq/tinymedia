package jpeg

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zzvanq/tinymedia/pkg/meta/codec"
)

func Test_JpegMetaManager_Insert_Errors(t *testing.T) {
	tests := []struct {
		name    string
		vendor  codec.MetaCodecVendor
		fields  map[string]string
		wantErr error
	}{
		{
			name:    "vendor not supported",
			vendor:  "unsupported",
			fields:  map[string]string{"test": "test"},
			wantErr: ErrVendorNotSupported,
		},
		{
			name:    "data size too large",
			vendor:  codec.TinyMetaVendor,
			fields:  map[string]string{"key": strings.Repeat("a", dataMaxSize)},
			wantErr: ErrDataSizeTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &JpegMetaManager{}
			if err := m.Insert(tt.vendor, tt.fields); err != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_JpegMetaManager_Insert(t *testing.T) {
	c, _ := JpegVendorsCodec[codec.TinyMetaVendor]

	fields := map[string]string{"k": "v"}
	encoded, _ := c.Codec.Encode(fields)
	s, _ := createSegment(c.Marker, c.VendorMagic, encoded)

	m := &JpegMetaManager{}
	m.segments = [][]byte{[]byte("test")}

	if err := m.Insert(codec.TinyMetaVendor, fields); err != nil {
		t.Errorf("Insert error = %v", err)
	}

	if len(m.segments) != 2 {
		t.Errorf("wrong m.segments length = %d", len(m.segments))
	}

	if !bytes.Equal(m.segments[0], s) {
		t.Errorf("wrong segment at 0:\nsegment: %v\nm.segments[0]: %v\n", s, m.segments[0])
	}
}

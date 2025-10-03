package jpeg

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/zzvanq/tinymedia/pkg/meta/codec"
)

func Test_JpegMetaManager_findSegment_foundParsed(t *testing.T) {
	knownSegment := []byte{0xFF, 0xE0, 0x00, 0x04, 0xFF, 0xE0}
	m := &JpegMetaManager{}
	m.segments = append(m.segments, knownSegment)
	want := 0
	i, err := m.findSegment(0xFFE0, []byte{0xFF, 0xE0})
	if err != nil {
		t.Errorf("want error: %v, got: %v", nil, err)
	}
	if i != want {
		t.Errorf("want: %v, got: %v", want, i)
	}
}

func Test_JpegMetaManager_findSegment(t *testing.T) {
	sosSegment := []byte{0xFF, 0xDA, 0x00, 0x02}
	data := append([]byte{0xFF, 0xE0, 0x00, 0x04, 0xFF, 0xE0}, sosSegment...)
	vendor := []byte{0xFF, 0xE0}

	tests := []struct {
		name    string
		r       io.Reader
		marker  uint16
		vendor  []byte
		want    int
		wantErr error
	}{
		{
			name:    "nextSegment returned corrupted segment",
			r:       bytes.NewReader([]byte{}),
			marker:  0xFFE0,
			vendor:  vendor,
			want:    0,
			wantErr: ErrCorruptedSegment,
		},
		{
			name:    "small segment",
			r:       bytes.NewReader([]byte{0xFF, 0xE0, 0x00, 0x03, 0xFF}),
			marker:  0xFFE0,
			vendor:  vendor,
			want:    0,
			wantErr: ErrCorruptedSegment,
		},
		{
			name:    "wrong marker",
			r:       bytes.NewReader(data),
			marker:  0xFFE1,
			vendor:  vendor,
			want:    0,
			wantErr: ErrMarkerNotFound,
		},
		{
			name:    "wrong vendor",
			r:       bytes.NewReader(data),
			marker:  0xFFE0,
			vendor:  []byte{0xFF, 0xE1},
			want:    0,
			wantErr: ErrMarkerNotFound,
		},
		{
			name:    "found",
			r:       bytes.NewReader(data),
			marker:  0xFFE0,
			vendor:  vendor,
			want:    0,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &JpegMetaManager{r: tt.r}
			got, err := m.findSegment(tt.marker, tt.vendor)
			if err != tt.wantErr {
				t.Errorf("want error: %v, got: %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func Test_JpegMetaManager_findParsed_SOSMarker(t *testing.T) {
	validSegment := []byte{0xFF, 0xE0, 0x00, 0x04, 0xFF, 0xE0}
	sosSegment := []byte{0xFF, 0xDA, 0x00, 0x02}
	r := bytes.NewReader(append(sosSegment, validSegment...))
	m := &JpegMetaManager{r: r}
	wantErr := ErrMarkerNotFound
	_, got := m.findSegment(0xFFE0, nil)
	if got != wantErr {
		t.Errorf("want error: %v, got: %v", wantErr, got)
	}
	if len(m.segments) != 1 {
		t.Errorf("want len(m.segments): 1, got: %v", len(m.segments))
	}
}

func Test_JpegMetaManager_findParsed(t *testing.T) {
	m := &JpegMetaManager{}
	c, _ := JpegVendorsCodec[codec.TinyMetaVendor]
	offsetSegment, _ := createSegment(c.Marker, []byte("offset"), []byte("test"))
	segment, _ := createSegment(c.Marker, c.VendorMagic, []byte("test"))

	m.segments = [][]byte{offsetSegment, segment}
	tests := []struct {
		name        string
		marker      uint16
		vendorMagic []byte
		want        int
		wantErr     error
	}{
		{
			name:        "wrong marker",
			marker:      0xFA,
			vendorMagic: nil,
			wantErr:     ErrMarkerNotFound,
		},
		{
			name:        "segment too short",
			marker:      0xFFE0,
			vendorMagic: append(segment, []byte{0xEE}...),
			wantErr:     ErrMarkerNotFound,
		},
		{
			name:        "found",
			marker:      0xFFE0,
			vendorMagic: c.VendorMagic,
			want:        1,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.findParsed(tt.marker, tt.vendorMagic)
			if err != tt.wantErr {
				t.Errorf("want error: %v, got: %v", tt.wantErr, err)
			}

			if got != tt.want {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func Test_JpegMetaManager_nextSegment(t *testing.T) {
	dataSize := 10
	smallData := make([]byte, headerSize+dataSize-1)
	binary.BigEndian.PutUint16(smallData[headerSize:2*headerSize], uint16(dataSize))

	correctData := []byte{0xFF, 0xE0, 0x00, 0x03, 0xFF}

	tests := []struct {
		name    string
		r       io.Reader
		want    []byte
		wantErr error
	}{
		{
			name:    "less than 2 * headerSize bytes",
			r:       bytes.NewReader([]byte{0xFF, 0xE0}),
			want:    nil,
			wantErr: ErrCorruptedSegment,
		},
		{
			name:    "less than data size",
			r:       bytes.NewReader(smallData),
			want:    nil,
			wantErr: ErrCorruptedSegment,
		},
		{
			name:    "success",
			r:       bytes.NewReader(correctData),
			want:    correctData,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &JpegMetaManager{r: tt.r}
			got, err := m.nextSegment()
			if err != tt.wantErr {
				t.Errorf("want error: %v, got: %v", tt.wantErr, err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

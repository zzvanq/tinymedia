package jpeg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"

	"github.com/zzvanq/tinymedia/internal/file/magic"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/codec/tinymeta"
)

const (
	sosMarker   = 0xFFDA
	headerSize  = 2
	dataMaxSize = 1<<16 - 1
)

var (
	ErrVendorNotSupported = errors.New("vendor not supported")
	ErrMarkerNotFound     = errors.New("marker not found")
	ErrDataSizeTooLarge   = errors.New("data size too large")
	ErrCorruptedSegment   = errors.New("corrupted segment")
)

type Codec interface {
	Encode(map[string]string) ([]byte, error)
	Decode([]byte) (map[string]string, error)
}

type CodecVendor struct {
	Codec       Codec
	Marker      uint16
	VendorMagic []byte
}

var JpegVendorsCodec = map[codec.MetaCodecVendor]CodecVendor{
	codec.TinyMetaVendor:     {tinymeta.TinyMeta, 0xFFE0, append([]byte(codec.TinyMetaVendor), 0)},
	codec.TinyMetaGzipVendor: {tinymeta.TinyMetaGzip, 0xFFE1, append([]byte(codec.TinyMetaGzipVendor), 0)},
}

type JpegMetaManager struct {
	prefix   []byte
	r        io.Reader
	segments [][]byte
}

// no filler bytes before the marker
func NewJpegMetaManager(r io.Reader) (*JpegMetaManager, error) {
	prefix := make([]byte, 2)
	if _, err := io.ReadFull(r, prefix); err != nil {
		return nil, fmt.Errorf("failed to read the magic bytes")
	}

	if !bytes.Equal(prefix, magic.JPEGMagic) {
		panic("not a jpeg")
	}

	return &JpegMetaManager{
		prefix:   prefix,
		r:        r,
		segments: [][]byte{},
	}, nil
}

func (m *JpegMetaManager) Insert(vendor codec.MetaCodecVendor, fields map[string]string) error {
	c, ok := JpegVendorsCodec[vendor]
	if !ok {
		return ErrVendorNotSupported
	}

	encoded, err := c.Codec.Encode(fields)
	if err != nil {
		return err
	}

	s, err := createSegment(c.Marker, c.VendorMagic, encoded)
	if err != nil {
		return err
	}
	m.segments = append([][]byte{s}, m.segments...)
	return nil
}

func (m *JpegMetaManager) Upsert(vendor codec.MetaCodecVendor, fields map[string]string) error {
	c, ok := JpegVendorsCodec[vendor]
	if !ok {
		return ErrVendorNotSupported
	}

	i, err := m.findSegment(c.Marker, c.VendorMagic)
	if err != nil {
		if err == ErrMarkerNotFound {
			return m.Insert(vendor, fields)
		}
		return err
	}
	s := m.segments[i]
	dataOffset := 2*headerSize + len(c.VendorMagic)
	decoded := make(map[string]string)
	if len(s[dataOffset:]) > 0 {
		decoded, err = c.Codec.Decode(s[dataOffset:])
		if err != nil {
			return err
		}
	}

	maps.Copy(decoded, fields)

	encoded, err := c.Codec.Encode(decoded)
	if err != nil {
		return err
	}

	newDataSize := headerSize + len(c.VendorMagic) + len(encoded)
	if newDataSize > dataMaxSize {
		return ErrDataSizeTooLarge
	}

	newSegment := append(s[:dataOffset], encoded...)
	binary.BigEndian.PutUint16(newSegment[headerSize:2*headerSize], uint16(newDataSize))
	m.segments[i] = newSegment
	return nil
}

func (m *JpegMetaManager) Extract(vendor codec.MetaCodecVendor, fields ...string) (map[string]string, error) {
	codecVendor, ok := JpegVendorsCodec[vendor]
	if !ok {
		return nil, ErrVendorNotSupported
	}

	i, err := m.findSegment(codecVendor.Marker, codecVendor.VendorMagic)
	if err != nil {
		if err == ErrMarkerNotFound {
			return nil, ErrMarkerNotFound
		}
		return nil, err
	}
	segment := m.segments[i]
	dataOffset := 2*headerSize + len(codecVendor.VendorMagic)
	decoded, err := codecVendor.Codec.Decode(segment[dataOffset:])
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(fields))
	for _, field := range fields {
		df, ok := decoded[field]
		if ok {
			result[field] = df
		}
	}
	return result, nil
}

func (m *JpegMetaManager) FileReader() io.Reader {
	readers := make([]io.Reader, 0, len(m.segments)+2)
	readers = append(readers, bytes.NewReader(m.prefix))
	for _, segment := range m.segments {
		readers = append(readers, bytes.NewReader(segment))
	}
	readers = append(readers, m.r)
	return io.MultiReader(readers...)
}

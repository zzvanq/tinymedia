package manager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"maps"

	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/codec/tinymeta"
)

const (
	sosMarker  = 0xFFDA
	headerSize = 2
)

var (
	ErrVendorNotSupported = errors.New("vendor not supported")
	ErrMarkerNotFound     = errors.New("marker not found")
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

var jpegVendorsCodec = map[codec.MetaCodecVendor]CodecVendor{
	codec.TinyMetaVendor:     {tinymeta.TinyMeta, 0xFFE0, append([]byte(codec.TinyMetaVendor), 0)},
	codec.TinyMetaGzipVendor: {tinymeta.TinyMetaGzip, 0xFFE1, append([]byte(codec.TinyMetaGzipVendor), 0)},
}

type JpegMetaManager struct {
	r        io.Reader
	segments [][]byte
}

func NewJpegMetaManager(r io.Reader) *JpegMetaManager {
	return &JpegMetaManager{
		r:        r,
		segments: [][]byte{},
	}
}

func (m *JpegMetaManager) Insert(vendor codec.MetaCodecVendor, fields map[string]string) error {
	codecVendor, ok := jpegVendorsCodec[vendor]
	if !ok {
		return ErrVendorNotSupported
	}

	encoded, err := codecVendor.Codec.Encode(fields)
	if err != nil {
		return err
	}

	dataSize := headerSize + len(codecVendor.VendorMagic) + len(encoded)
	segment := make([]byte, headerSize+dataSize)
	binary.BigEndian.PutUint16(segment[:headerSize], codecVendor.Marker)
	copy(segment[headerSize:], encoded)

	m.segments = append([][]byte{segment}, m.segments...)
	return nil
}

func (m *JpegMetaManager) Upsert(vendor codec.MetaCodecVendor, fields map[string]string) error {
	codecVendor, ok := jpegVendorsCodec[vendor]
	if !ok {
		return ErrVendorNotSupported
	}

	i, ok := m.findSegment(codecVendor.Marker, codecVendor.VendorMagic)
	if !ok {
		return m.Insert(vendor, fields)
	}
	segment := m.segments[i]
	dataOffset := 2*headerSize + len(codecVendor.VendorMagic)
	decoded, err := codecVendor.Codec.Decode(segment[dataOffset:])
	if err != nil {
		return err
	}

	maps.Copy(decoded, fields)

	encoded, err := codecVendor.Codec.Encode(decoded)
	if err != nil {
		return err
	}

	newSegment := append(segment[:dataOffset], encoded...)
	newDataSize := headerSize + len(codecVendor.VendorMagic) + len(encoded)
	binary.BigEndian.PutUint16(newSegment[headerSize:2*headerSize], uint16(newDataSize))
	m.segments[i] = newSegment
	return nil
}

func (m *JpegMetaManager) Extract(vendor codec.MetaCodecVendor, fields ...string) (map[string]string, error) {
	codecVendor, ok := jpegVendorsCodec[vendor]
	if !ok {
		return nil, ErrVendorNotSupported
	}

	i, ok := m.findSegment(codecVendor.Marker, codecVendor.VendorMagic)
	if !ok {
		return nil, ErrMarkerNotFound
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
	readers[0] = bytes.NewReader(file.JPEGMagic)
	for _, segment := range m.segments {
		readers = append(readers, bytes.NewReader(segment))
	}
	readers = append(readers, m.r)
	return io.MultiReader(readers...)
}

// 'm.r' must be at the start of a marker.
func (m *JpegMetaManager) findSegment(marker uint16, vendorMagic []byte) (int, bool) {
	i, ok := m.findParsed(marker, vendorMagic)
	if ok {
		return i, true
	}

	for {
		segment, err := m.nextSegment()
		if err != nil {
			return 0, false
		}
		m.segments = append(m.segments, segment)

		if len(segment) < 2*headerSize+len(vendorMagic) {
			continue
		}

		segMarker := binary.BigEndian.Uint16(segment[0:headerSize])
		vendorBytes := segment[2*headerSize : 2*headerSize+len(vendorMagic)]
		if segMarker == marker && bytes.Equal(vendorBytes, vendorMagic) {
			return len(m.segments) - 1, true
		}

		// there is no more metadata
		if segMarker == sosMarker {
			return 0, false
		}
	}
}

func (m *JpegMetaManager) nextSegment() ([]byte, error) {
	headers := make([]byte, 2*headerSize)
	if _, err := io.ReadFull(m.r, headers); err != nil {
		return nil, err
	}

	segDataSize := binary.BigEndian.Uint16(headers[headerSize : 2*headerSize])

	segment := make([]byte, int(segDataSize)+headerSize)
	copy(segment[:len(headers)], headers)
	if _, err := io.ReadFull(m.r, segment[len(headers):]); err != nil {
		return nil, err
	}

	return segment, nil
}

func (m *JpegMetaManager) findParsed(marker uint16, vendorMagic []byte) (int, bool) {
	for i, s := range m.segments {
		sMarker := binary.BigEndian.Uint16(s[:headerSize])
		if sMarker != marker {
			continue
		}

		if len(s) < 2*headerSize+len(vendorMagic) {
			continue
		}

		sVendorBytes := s[2*headerSize : 2*headerSize+len(vendorMagic)]
		if bytes.Equal(sVendorBytes, vendorMagic) {
			return i, true
		}
	}
	return 0, false
}

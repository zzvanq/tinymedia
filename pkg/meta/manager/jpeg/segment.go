package jpeg

import (
	"bytes"
	"encoding/binary"
	"io"
)

// 'm.r' must be at the start of a marker.
func (m *JpegMetaManager) findSegment(marker uint16, vendorMagic []byte) (int, error) {
	i, err := m.findParsed(marker, vendorMagic)
	if err == nil {
		return i, nil
	}

	var segMarker uint16
	for segMarker != sosMarker {
		segment, err := m.nextSegment()
		if err != nil {
			return 0, err
		}
		m.segments = append(m.segments, segment)

		segMarker = binary.BigEndian.Uint16(segment[0:headerSize])
		if len(segment) < 2*headerSize+len(vendorMagic) {
			continue
		}

		vendorBytes := segment[2*headerSize : 2*headerSize+len(vendorMagic)]
		if segMarker == marker && bytes.Equal(vendorBytes, vendorMagic) {
			return len(m.segments) - 1, nil
		}
	}
	return 0, ErrMarkerNotFound
}

func (m *JpegMetaManager) nextSegment() ([]byte, error) {
	headers := make([]byte, 2*headerSize)
	if _, err := io.ReadFull(m.r, headers); err != nil {
		return nil, ErrCorruptedSegment
	}

	segDataSize := binary.BigEndian.Uint16(headers[headerSize : 2*headerSize])
	segment := make([]byte, int(segDataSize)+headerSize)
	copy(segment[:len(headers)], headers)
	if _, err := io.ReadFull(m.r, segment[len(headers):]); err != nil {
		return nil, ErrCorruptedSegment
	}

	return segment, nil
}

func (m *JpegMetaManager) findParsed(marker uint16, vendorMagic []byte) (int, error) {
	for i, s := range m.segments {
		sMarker := binary.BigEndian.Uint16(s[:headerSize])
		if sMarker != marker {
			continue
		}

		if len(s) < 2*headerSize+len(vendorMagic) {
			continue
		}

		sVendorMagic := s[2*headerSize : 2*headerSize+len(vendorMagic)]
		if bytes.Equal(sVendorMagic, vendorMagic) {
			return i, nil
		}
	}
	return 0, ErrMarkerNotFound
}

func createSegment(marker uint16, vendor []byte, data []byte) ([]byte, error) {
	dataSize := headerSize + len(vendor) + len(data)
	if dataSize > dataMaxSize {
		return nil, ErrDataSizeTooLarge
	}

	s := make([]byte, headerSize+dataSize)
	binary.BigEndian.PutUint16(s[:headerSize], marker)
	binary.BigEndian.PutUint16(s[headerSize:2*headerSize], uint16(dataSize))
	copy(s[2*headerSize:2*headerSize+len(vendor)], vendor)
	copy(s[2*headerSize+len(vendor):], data)
	return s, nil
}

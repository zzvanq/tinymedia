package tinymeta

import (
	"bytes"
	"compress/gzip"
	"io"
)

type tinyMetaGzip struct{}

var TinyMetaGzip = tinyMetaGzip{}

func (t tinyMetaGzip) Encode(fields map[string]string) ([]byte, error) {
	data, err := TinyMeta.Encode(fields)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t tinyMetaGzip) Decode(data []byte) (map[string]string, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	zData, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return TinyMeta.Decode(zData)
}

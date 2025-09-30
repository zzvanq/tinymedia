package tinymeta

import (
	"encoding/json"
)

type tinyMeta struct{}

var TinyMeta = tinyMeta{}

func (t tinyMeta) Encode(fields map[string]string) ([]byte, error) {
	return json.Marshal(fields)
}

func (t tinyMeta) Decode(data []byte) (map[string]string, error) {
	var fields map[string]string
	err := json.Unmarshal(data, &fields)
	return fields, err
}

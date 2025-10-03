package file

import (
	"bytes"
	"os"
	"testing"
)

func Test_UpdateFile(t *testing.T) {
	f, err := os.CreateTemp("./", "test")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	want := []byte{0xFF, 0xD8, 0x00, 0x02}
	r := bytes.NewReader(want)
	if err := UpdateFile(r, f.Name()); err != nil {
		t.Errorf("want error: %v, got: %v", nil, err)
	}

	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("want error: %v, got: %v", nil, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("want: %v, got: %v", want, got)
	}
	defer f.Close()
}

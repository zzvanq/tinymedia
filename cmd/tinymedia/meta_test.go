package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	build := exec.Command("go", "build", "-o", "tinymedia.test", ".")
	if err := build.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	code := m.Run()

	os.Remove("tinymedia.test")
	os.Exit(code)
}

func Test_handleMeta_Insert(t *testing.T) {
	testFile := filepath.Join("./", "test.jpg")

	createTestJPEG(t, testFile, "tinymeta", nil)
	defer os.Remove(testFile)

	artist := "Test Artist"
	title := "Test Title"
	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", fmt.Sprintf("artist=%s,title=%s", artist, title),
		"-mv", "tinymeta",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist,title",
		"-mv", "tinymeta",
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, kvQuote("artist", artist)) {
		t.Errorf("artist field not set:\n%s", got)
	}
	if !strings.Contains(got, kvQuote("title", title)) {
		t.Errorf("title field not set:\n%s", got)
	}
}

func Test_handleMeta_Update(t *testing.T) {
	testFile := filepath.Join("./", "test.jpg")

	createTestJPEG(t, testFile, "tinymeta", map[string]string{
		"artist": "Old Artist",
	})
	defer os.Remove(testFile)

	artist := "New Artist"
	title := "New Title"
	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", fmt.Sprintf("artist=%s,title=%s", artist, title),
		"-mv", "tinymeta",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("update failed: %v\noutput: %s", err, output)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist,title",
		"-mv", "tinymeta",
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, kvQuote("artist", artist)) {
		t.Errorf("artist not updated, got:\n%s", got)
	}
	if !strings.Contains(got, kvQuote("title", title)) {
		t.Errorf("title not updated, got:\n%s", got)
	}
}

func Test_handleMeta_MultipleFiles(t *testing.T) {
	file1 := filepath.Join("./", "test1.jpg")
	file2 := filepath.Join("./", "test2.jpg")

	createTestJPEG(t, file1, "tinymeta", nil)
	defer os.Remove(file1)
	createTestJPEG(t, file2, "tinymeta", nil)
	defer os.Remove(file2)

	artist := "Batch Artist"
	cmd := exec.Command("./tinymedia.test",
		"-i", file1,
		"-i", file2,
		"-m", "artist="+artist,
		"-mv", "tinymeta",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("batch update failed: %v\noutput: %s", err, output)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", file1,
		"-i", file2,
		"-m", "artist",
		"-mv", "tinymeta",
	)
	output, _ = cmd.CombinedOutput()
	if !strings.Contains(string(output), kvQuote("artist", artist)) {
		t.Errorf("no metadata returned: %s", output)
	}
	if !strings.Contains(string(output), "File="+file1) {
		t.Errorf("no data for file1: %s", output)
	}
	if !strings.Contains(string(output), "File="+file2) {
		t.Errorf("no data for file2: %s", output)
	}
}

func Test_handleMeta_MissingVendor(t *testing.T) {
	testFile := filepath.Join("./", "test.jpg")
	createTestJPEG(t, testFile, "tinymeta", nil)
	defer os.Remove(testFile)

	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist",
		// Missing -mv flag
	)

	output, _ := cmd.CombinedOutput()
	got := string(output)
	if !strings.Contains(got, "-mv is required") {
		t.Errorf("expected error message about -mv, got:\n%s", got)
	}
}

func Test_handleMeta_NonexistentFile(t *testing.T) {
	file := filepath.Join("./", "file.jpg")
	cmd := exec.Command("./tinymedia.test",
		"-i", file,
		"-m", "artist",
		"-mv", "tinymeta",
	)

	output, _ := cmd.CombinedOutput()
	got := string(output)
	if !strings.Contains(got, "failed to open") && !strings.Contains(got, "no such file") {
		t.Errorf("expected file not found error, got:\n%s", got)
	}
}

func Test_handleMeta_MixedReadAndUpdate(t *testing.T) {
	testFile := filepath.Join("./", "test.jpg")

	artist := "Original Artist"
	album := "Original Album"
	createTestJPEG(t, testFile, "tinymeta", map[string]string{
		"artist": artist,
		"album":  album,
	})
	defer os.Remove(testFile)

	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", fmt.Sprintf("artist,album=%s", album),
		"-mv", "tinymeta",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\noutput: %s", err, output)
	}

	got := string(output)

	if !strings.Contains(got, kvQuote("artist", artist)) {
		t.Errorf("output missing output field:\n%s", got)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "album",
		"-mv", "tinymeta",
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("verification read failed: %v", err)
	}

	if !strings.Contains(string(output), kvQuote("album", album)) {
		t.Errorf("album not updated, got:\n%s", output)
	}
}

func Test_handleMeta_TinyMetaGzip(t *testing.T) {
	testFile := "./test.jpg"
	createTestJPEG(t, testFile, "tinymetagzip", nil)
	defer os.Remove(testFile)

	artist := "Compressed Artist"
	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist="+artist,
		"-mv", "tinymetagzip",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gzip insert failed: %v\noutput: %s", err, output)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist",
		"-mv", "tinymetagzip",
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gzip read failed: %v", err)
	}

	if !strings.Contains(string(output), kvQuote("artist", artist)) {
		t.Errorf("gzip metadata not set, got:\n%s", output)
	}
}

func Test_handleMeta_EmptyFields(t *testing.T) {
	testFile := filepath.Join("./", "test.jpg")
	createTestJPEG(t, testFile, "tinymeta", nil)
	defer os.Remove(testFile)

	artist := "Test"
	title := "Title"
	cmd := exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", fmt.Sprintf("artist=%s,,title=%s", artist, title),
		"-mv", "tinymeta",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	cmd = exec.Command("./tinymedia.test",
		"-i", testFile,
		"-m", "artist,title",
		"-mv", "tinymeta",
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, kvQuote("artist", artist)) || !strings.Contains(got, kvQuote("title", title)) {
		t.Errorf("fields not set, got:\n%s", got)
	}
}

func createTestJPEG(t *testing.T, path string, vendor string, metadata map[string]string) {
	t.Helper()

	data := []byte{0xFF, 0xD8}

	if len(metadata) > 0 {
		vendorMagic := append([]byte(vendor), 0)
		dataSize := make([]byte, 2)
		binary.BigEndian.PutUint16(dataSize, uint16(2+len(vendorMagic)))
		app0 := []byte{
			0xFF, 0xE0,
			dataSize[0], dataSize[1],
		}
		data = append(data, app0...)
		data = append(data, vendorMagic...)
	}

	sos := []byte{
		0xFF, 0xDA,
		0x00, 0x02,
	}
	data = append(data, sos...)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to create test JPEG: %v", err)
	}

	if len(metadata) > 0 {
		var fields []string
		for k, v := range metadata {
			fields = append(fields, k+"="+v)
		}

		cmd := exec.Command("./tinymedia.test",
			"-i", path,
			"-m", strings.Join(fields, ","),
			"-mv", vendor,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to insert initial metadata: %v\noutput: %s", err, output)
		}
	}
}

func kvQuote(key, value string) string {
	return fmt.Sprintf(`"%s"="%s"`, key, value)
}

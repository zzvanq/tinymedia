package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/manager"
)

func handleMeta(fileNames []string, fields []string, vendor string) {
	if vendor == "" {
		fmt.Println("-mv is required when -m is used")
		return
	}

	readFields, updateFields := parseFields(fields)

	errCh := make(chan error, len(fileNames))

	var wg sync.WaitGroup
	wg.Add(len(fileNames))
	for _, fn := range fileNames {
		go func() {
			defer wg.Done()
			if err := processFile(fn, vendor, readFields, updateFields); err != nil {
				errCh <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		fmt.Println(err)
	}
}

func processFile(fn string, vendor string, readFields []string, updateFields map[string]string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	metaManager, err := manager.NewMetaManager(f)
	if err != nil {
		return err
	}

	var newReader io.Reader
	if len(updateFields) > 0 {
		if err := metaManager.Upsert(codec.MetaCodecVendor(vendor), updateFields); err != nil {
			return err
		}
		newReader = metaManager.FileReader()
	}

	if len(readFields) > 0 {
		result, err := printMeta(metaManager, vendor, readFields)
		if err != nil {
			return err
		}
		fmt.Print("File=", fn, "\n", result, "\n")
	}

	if len(updateFields) > 0 {
		file.UpdateFile(newReader, f.Name())
	}
	return nil
}

func parseFields(fields []string) ([]string, map[string]string) {
	var readFields []string
	updateFields := make(map[string]string)

	for _, field := range fields {
		if field == "" {
			continue
		}

		parts := strings.SplitN(field, "=", 2)
		switch len(parts) {
		case 1:
			readFields = append(readFields, parts[0])
		case 2:
			updateFields[parts[0]] = parts[1]
		}
	}

	return readFields, updateFields
}

func printMeta(metaManager manager.MetaManager, vendor string, fields []string) (string, error) {
	extracted, err := metaManager.Extract(codec.MetaCodecVendor(vendor), fields...)
	if err != nil {
		return "", err
	}

	resultBuilder := strings.Builder{}
	for k, v := range extracted {
		resultBuilder.WriteString(fmt.Sprintf("%s=%s\n", strconv.Quote(k), strconv.Quote(v)))
	}
	return resultBuilder.String(), nil
}

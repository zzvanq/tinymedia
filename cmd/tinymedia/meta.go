package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/manager"
)

func handleMeta(inputs []string, fields []string, vendor string) {
	if vendor == "" {
		fmt.Println("-mv is required when -m is used")
		return
	}

	var (
		readFields   []string
		updateFields = make(map[string]string)
	)
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

	var wg sync.WaitGroup
	errCh := make(chan error, len(inputs))

	wg.Add(len(inputs))
	for _, input := range inputs {
		go func() {
			defer wg.Done()
			file, err := os.Open(input)
			if err != nil {
				errCh <- fmt.Errorf("failed to open %s: %v", input, err)
				return
			}
			defer file.Close()
			_, err = handleFileMeta(file, vendor, updateFields, readFields)
			if err != nil {
				errCh <- fmt.Errorf("file %s: %s", input, err.Error())
				return
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

func handleFileMeta(file io.Reader, vendor string, updateFields map[string]string, readFields []string) (io.Reader, error) {
	newReader := file
	metaManager, err := manager.NewMetaManager(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	if len(updateFields) != 0 {
		if err := metaManager.Upsert(codec.MetaCodecVendor(vendor), updateFields); err != nil {
			return nil, fmt.Errorf("failed to update metadata: %w", err)
		}

		newReader = metaManager.FileReader()
	}

	fields, err := metaManager.Extract(codec.MetaCodecVendor(vendor), readFields...)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	resultBuilder := strings.Builder{}
	for field, value := range fields {
		resultBuilder.WriteString(fmt.Sprintf("%s=%s\n", strconv.Quote(field), strconv.Quote(value)))
	}
	fmt.Println(resultBuilder.String())
	return newReader, nil
}

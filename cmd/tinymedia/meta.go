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

func handleMeta(files []string, fields []string, vendor string) {
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
	errCh := make(chan error, len(files))

	wg.Add(len(files))
	for _, f := range files {
		go func() {
			defer wg.Done()
			file, err := os.Open(f)
			if err != nil {
				errCh <- fmt.Errorf("failed to open %s: %w", f, err)
				return
			}
			modifiedFile, err := handleFileMeta(file, vendor, updateFields, readFields)
			if err != nil {
				errCh <- fmt.Errorf("file %s: %s", f, err.Error())
				return
			}

			file.Close()
			if updateFields != nil {
				tmpFile, err := os.CreateTemp("", "tinymedia-*")
				if err != nil {
					errCh <- fmt.Errorf("failed to create temp file: %w", err)
					return
				}

				defer tmpFile.Close()
				if _, err := io.Copy(tmpFile, modifiedFile); err != nil {
					errCh <- fmt.Errorf("failed to copy file: %w", err)
					return
				}

				if err := os.Rename(tmpFile.Name(), file.Name()); err != nil {
					os.Remove(tmpFile.Name())
					errCh <- fmt.Errorf("failed to rename temp file: %w", err)
					return
				}
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
	metaManager, err := manager.NewMetaManager(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	newReader := file
	if len(updateFields) > 0 {
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

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

	errCh := make(chan error, len(fileNames))

	var wg sync.WaitGroup
	wg.Add(len(fileNames))
	for _, fn := range fileNames {
		go func() {
			defer wg.Done()
			f, err := os.Open(fn)
			if err != nil {
				errCh <- fmt.Errorf("failed to open %s: %w", fn, err)
				return
			}

			metaManager, err := manager.NewMetaManager(f)
			if err != nil {
				errCh <- fmt.Errorf("failed to read metadata: %w", err)
				return
			}

			var newReader io.Reader
			if len(updateFields) > 0 {
				if err := metaManager.Upsert(codec.MetaCodecVendor(vendor), updateFields); err != nil {
					errCh <- fmt.Errorf("failed to update metadata: %w", err)
					return
				}
				newReader = metaManager.FileReader()
			}

			if len(readFields) > 0 {
				if err := printMeta(metaManager, vendor, readFields); err != nil {
					errCh <- err
				}
			}
			f.Close()

			if updateFields != nil {
				file.UpdateFile(newReader, f.Name())
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

func printMeta(metaManager manager.MetaManager, vendor string, fields []string) error {
	extracted, err := metaManager.Extract(codec.MetaCodecVendor(vendor), fields...)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	resultBuilder := strings.Builder{}
	for k, v := range extracted {
		resultBuilder.WriteString(fmt.Sprintf("%s=%s\n", strconv.Quote(k), strconv.Quote(v)))
	}
	fmt.Println(resultBuilder.String())
	return nil
}

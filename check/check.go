package check

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/xackery/rof2plus/checksum"
)

var (
	mux       sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	summaries []*Summary
)

// Summary is a result list from the last check
type Summary struct {
	Path    string
	Message string
}

// Check checks the path.
func Check(client checksum.ChecksumClient, rootPath string) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Check took %0.2fs seconds\n", time.Since(start).Seconds())
	}()
	mux.Lock()
	ctx, cancel = context.WithCancel(context.Background())
	mux.Unlock()

	defer Close()

	wg := &sync.WaitGroup{}
	summaryChan := make(chan *Summary, 9999999)

	totalCount := 0

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		wg.Add(1)
		totalCount++
		go checkPath(wg, summaryChan, client, rootPath, path, d, err)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walkdir: %w", err)
	}

	fmt.Println("Total files to check:", totalCount)
	wg.Wait()
	mux.Lock()
	defer mux.Unlock()

	for {
		if len(summaryChan) == 0 {
			break
		}

		summary := <-summaryChan
		fmt.Println(summary.Path, summary.Message)
	}

	return nil
}

// Close cancels the check.
func Close() {
	mux.Lock()
	defer mux.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Report returns the last check result.
func Report() []*Summary {
	mux.RLock()
	defer mux.RUnlock()
	return summaries
}

func checkPath(wg *sync.WaitGroup, summaryChan chan *Summary, client checksum.ChecksumClient, rootPath string, path string, d fs.DirEntry, err error) error {
	defer wg.Done()

	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fi, err := d.Info()
	if err != nil {
		return fmt.Errorf("info: %w", err)
	}

	relativePath := filepath.ToSlash(path[len(rootPath):])

	size := checksum.FileSize(client, relativePath)
	if size > 0 && size != fi.Size() {
		summaryChan <- &Summary{
			Path:    relativePath,
			Message: fmt.Sprintf("size mismatch %d vs %d", size, fi.Size()),
		}

		return nil
	}
	if size == -1 {
		//fmt.Println(relativePath, "Untracked file")
		return nil
	}

	//fmt.Printf("%s %d vs %d OK\n", relativePath, size, fi.Size())

	return nil
}

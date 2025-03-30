// patch fetches files from github or another url and downloads it to a server directory
package patch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xackery/rof2plus/check"
	"github.com/xackery/rof2plus/checksum"
)

var (
	isDownloading   atomic.Bool
	progressPercent atomic.Int32
)

type downloadRequest struct {
	Name string
	Path string
	URL  string
}

type downloadResult struct {
	Name string
	Size int64
	Err  error
}

// Download downloads a list of files to path
func Download(filelist *checksum.FileList, path string) error {
	var err error

	if isDownloading.Load() {
		return fmt.Errorf("already downloading")
	}
	isDownloading.Store(true)
	defer isDownloading.Store(false)

	err = checksum.SetPatcherFilelist(filelist)
	if err != nil {
		return fmt.Errorf("set patcher filelist: %w", err)
	}

	checksum.SetExcludedClients(checksum.ClientRoF2Core)

	err = check.Check(checksum.ClientPatcher, path)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}

	downloads := []checksum.FileEntry{}

	isPatchNeeded := false
	report := check.Report()
	if report != nil {
		for _, fail := range report.Failures {
			switch fail.Error {
			case check.ErrorNotFound:
				isFound := false
				for _, file := range filelist.Downloads {
					if fail.Path != file.Name {
						continue
					}
					isFound = true
					downloads = append(downloads, file)
					break
				}
				if !isFound {
					return fmt.Errorf("file %s not found in filelist", fail.Path)
				}

				isPatchNeeded = true
				continue
			default:
			}
		}
	}
	if !isPatchNeeded {
		fmt.Println("No patch needed")
		return nil
	}

	start := time.Now()
	totalSizeToDownloadInKB := int64(0)
	totalSizeDownloadedInKB := int64(0)
	totalCount := len(downloads)

	downloadRequestChan := make(chan *downloadRequest, 100000)
	downloadResultChan := make(chan *downloadResult, 1000)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer func() {
		if err != nil {
			cancel(err)
			return
		}
		cancel(nil)
	}()

	if strings.Contains(filelist.DownloadPrefix, "master/rof") {
		filelist.DownloadPrefix = strings.ReplaceAll(filelist.DownloadPrefix, "master/rof", "refs/heads/master/rof")
	}

	for _, file := range downloads {
		totalSizeToDownloadInKB += int64(file.Size) / 1024
		dirPath := filepath.Dir(filepath.Join(path, file.Name))
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}

		downloadRequestChan <- &downloadRequest{Name: file.Name, Path: path, URL: strings.TrimSuffix(filelist.DownloadPrefix, "/")}
	}
	fmt.Println("Downloading", totalCount, "files")
	if totalSizeToDownloadInKB < 1 {
		totalSizeToDownloadInKB = 1
	}

	isDone := make(chan bool)

	numJobsConcurrent := 100
	// job consumer
	for range numJobsConcurrent {
		go downloader(ctx, downloadRequestChan, downloadResultChan)
	}

	// result consumer
	go func() {
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-downloadResultChan:
				count++
				if result.Err != nil {
					err = result.Err
					cancel(result.Err)
					return
				}
				totalSizeDownloadedInKB += result.Size / 1024
				progressPercent.Store(int32(totalSizeDownloadedInKB * 100 / totalSizeToDownloadInKB))

				size := fmt.Sprintf("(%0.2f KB)", float64(result.Size)/1024)
				if result.Size > 1024*1024 {
					size = fmt.Sprintf("(%0.2f MB)", float64(result.Size)/1024/1024)
				}

				fmt.Printf("%s %s (%d/%d)  %d%%\n", result.Name, size, count, totalCount, progressPercent.Load())
			}
			if count >= totalCount {
				size := fmt.Sprintf("(%0.2f KB)", float64(totalSizeDownloadedInKB))
				if totalSizeDownloadedInKB > 1024 {
					size = fmt.Sprintf("(%0.2f MB)", float64(totalSizeDownloadedInKB)/1024)
				}

				fmt.Printf("Downloaded %d files %s in %0.2fs\n", count, size, time.Since(start).Seconds())
				break
			}
		}
		close(isDone)
	}()

	// wait for consumer to finish
	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}
		return fmt.Errorf("download failed")
	case <-isDone:
	}

	return nil
}

func downloader(ctx context.Context, downloadRequestChan chan *downloadRequest, downloadResultChan chan *downloadResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-downloadRequestChan:
			if request == nil {
				return
			}

			requestNameToURL := strings.ReplaceAll(request.Name, "\\", "/")
			//requestNameToURL = strings.ReplaceAll(requestNameToURL, " ", "%20")
			copiedBytes, err := downloadFile(ctx, request.URL+"/"+requestNameToURL, filepath.Join(request.Path, request.Name))
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case downloadResultChan <- &downloadResult{Name: request.Name, Size: copiedBytes, Err: err}:
				}
				return
			}
			select {
			case <-ctx.Done():
				return
			case downloadResultChan <- &downloadResult{Name: request.Name, Size: copiedBytes, Err: nil}:
			}
		}
	}
}

// downloadFile downloads a file from the given URL to the specified path
func downloadFile(ctx context.Context, url string, path string) (int64, error) {

	// Create the file
	w, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("create file %s: %w", path, err)
	}
	defer w.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download %s responded HTTP status code %d", url, resp.StatusCode)
	}

	// Write the body to file
	copiedBytes, err := io.Copy(w, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("write file %s: %w", path, err)
	}

	return copiedBytes, nil
}

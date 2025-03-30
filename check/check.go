// check validates a client based on context to verify files are accurate and up to date
package check

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/xackery/rof2plus/checksum"
)

var (
	mux    sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	report *ReportDetail
)

type ReportDetail struct {
	FileTotal int
	OKTotal   int
	FailTotal int
	Failures  []*Summary
	Successes []*Summary
}

func (e *ReportDetail) String() string {
	if e.FailTotal > 0 {
		return fmt.Sprintf("Total: %d OK: %d Fail: %d (First Failure: Client: %s, Path: %s, Reason: %s)", e.FileTotal, e.OKTotal, e.FailTotal, e.Failures[0].Client.String(), e.Failures[0].Path, e.Failures[0].Directions)
	}
	return fmt.Sprintf("Total: %d OK: %d Fail: %d", e.FileTotal, e.OKTotal, e.FailTotal)
}

type SummaryError int

const (
	ErrorNone SummaryError = iota
	ErrorNotFound
	ErrorSize
	ErrorHash
	ErrorIsDir
	ErrorCancelled
)

// Summary is a result list from the last check
type Summary struct {
	Path       string
	Error      SummaryError
	Directions string
	Client     checksum.ChecksumClient
}

func (e *Summary) String() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Directions)
}

// Check checks the path.
func Check(client checksum.ChecksumClient, rootPath string) error {

	if rootPath == "" {
		return fmt.Errorf("path is empty")
	}

	fi, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %w", err)
		}
		return fmt.Errorf("stat path: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

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

	checksums := map[string]*checksum.ChecksumEntry{}

	chk, err := checksum.ByClient(client)
	if err != nil {
		return fmt.Errorf("checksum byclient rof2: %w", err)
	}
	for k, v := range chk {
		checksums[k] = v
	}

	for filePath, entry := range checksums {
		wg.Add(1)
		totalCount++
		if entry.Path == "arena.eqg" {
			fmt.Println("arena.eqg")
		}

		go checkPath(wg, summaryChan, client, rootPath, filePath, entry.IsDeleted)
	}

	wg.Wait()
	mux.Lock()
	defer mux.Unlock()

	report = &ReportDetail{
		FileTotal: totalCount,
	}

	for {
		if len(summaryChan) == 0 {
			break
		}
		summary := <-summaryChan
		if summary.Error != ErrorNone {
			report.FailTotal++
			report.Failures = append(report.Failures, summary)
			continue
		}
		report.OKTotal++
		report.Successes = append(report.Successes, summary)
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
func Report() *ReportDetail {
	mux.RLock()
	defer mux.RUnlock()
	if report != nil {
		sort.Slice(report.Failures, func(i, j int) bool {
			return report.Failures[i].Path < report.Failures[j].Path
		})
		sort.Slice(report.Successes, func(i, j int) bool {
			return report.Successes[i].Path < report.Successes[j].Path
		})
	}
	return report
}

func checkPath(wg *sync.WaitGroup, summaryChan chan *Summary, client checksum.ChecksumClient, rootPath string, relativePath string, isDeleted bool) {
	defer wg.Done()

	fullPath := fmt.Sprintf("%s/%s", rootPath, relativePath)
	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			if isDeleted {
				summaryChan <- &Summary{
					Path:       relativePath,
					Error:      ErrorNone,
					Client:     client,
					Directions: "Deleted",
				}
				return
			}
			summaryChan <- &Summary{
				Path:       relativePath,
				Error:      ErrorNotFound,
				Directions: "File not found",
				Client:     client,
			}
			return
		}
		summaryChan <- &Summary{
			Path:       relativePath,
			Error:      ErrorNotFound,
			Client:     client,
			Directions: fmt.Sprintf("File not found: %v", err),
		}
		return
	}
	if fi.IsDir() {
		summaryChan <- &Summary{
			Path:       relativePath,
			Error:      ErrorIsDir,
			Client:     client,
			Directions: "Is a directory",
		}
		return
	}

	select {
	case <-ctx.Done():
		summaryChan <- &Summary{
			Path:       relativePath,
			Error:      ErrorCancelled,
			Client:     client,
			Directions: "Job Cancelled",
		}
		return
	default:
	}

	size := checksum.FileSize(client, relativePath)
	if size == -1 {
		summaryChan <- &Summary{
			Path:       relativePath,
			Error:      ErrorSize,
			Client:     client,
			Directions: "Size returned -1",
		}
		return

		//fmt.Println(relativePath, "Untracked file")
	}
	//fmt.Println("Checking", relativePath, "size", size)
	if size > 0 && size != fi.Size() {
		// fall back to checking md5

		md5, err := checksum.MD5Generate(fullPath)
		if err != nil {
			summaryChan <- &Summary{
				Path:       relativePath,
				Error:      ErrorSize,
				Client:     client,
				Directions: fmt.Sprintf("MD5 Failure: %v", err),
			}
		}

		if md5 != checksum.MD5Hash(client, relativePath) {
			summaryChan <- &Summary{
				Path:       relativePath,
				Error:      ErrorSize,
				Client:     client,
				Directions: fmt.Sprintf("MD5 Failure: sizes %d vs %d", size, fi.Size()),
			}
		}
	}

	//fmt.Printf("%s %d vs %d OK\n", relativePath, size, fi.Size())

	summaryChan <- &Summary{
		Path:       relativePath,
		Error:      ErrorNone,
		Client:     client,
		Directions: "OK",
	}
}

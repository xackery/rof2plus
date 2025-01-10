package checksum

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestChecksumGen(t *testing.T) {
	rofPath := os.Getenv("ROF_BASE_PATH")
	if rofPath == "" {
		t.Skip("ROF_BASE_PATH not set")
	}

	lsPath := os.Getenv("LS_BASE_PATH")
	if lsPath == "" {
		t.Skip("LS_BASE_PATH not set")
	}

	rootPath := rofPath
	start := time.Now()
	context := "ls"
	w, err := os.Create(fmt.Sprintf("%s.txt", context))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer w.Close()

	var wopt *os.File
	if context == "ls" {
		wopt, err = os.Create(fmt.Sprintf("%s_opt.txt", context))
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		defer wopt.Close()

		wopt.WriteString("package checksum\n\n")
		wopt.WriteString(fmt.Sprintf("var %sOptChecksums = map[string]*ChecksumEntry{\n", context))
	}

	w.WriteString("package checksum\n\n")
	w.WriteString(fmt.Sprintf("var %sChecksums = map[string]*ChecksumEntry{\n", context))

	tmpFiles := map[string]ChecksumEntry{}

	err = filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relativePath := filepath.ToSlash(path[len(rootPath)+1:])

		size := info.Size()

		if isFileExcluded(context, path, relativePath) {
			return nil
		}

		tmpFiles[relativePath] = ChecksumEntry{FileSize: size}

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk dir: %v", err)
	}

	wg := &sync.WaitGroup{}
	chkChan := make(chan ChecksumEntry, 100)

	for relativePath, entry := range tmpFiles {
		path := filepath.Join(rootPath, relativePath)
		size := entry.FileSize
		wg.Add(1)
		go concurrentChecksum(wg, chkChan, path, relativePath, size)
		continue
	}

	wg.Wait()

	files := []string{}
	optFiles := []string{}

	for {
		if len(chkChan) == 0 {
			break
		}
		chkEntry := <-chkChan
		if context == "ls" {
			rofHash := MD5Hash(ClientRoF2, chkEntry.Path)

			if rofHash != "" && rofHash != chkEntry.MD5Hash {
				optFiles = append(optFiles, fmt.Sprintf("\t\"%s\": {MD5Hash: \"%s\", FileSize: %d},\n", chkEntry.Path, chkEntry.MD5Hash, tmpFiles[chkEntry.Path].FileSize))
			}
		}
		files = append(files, fmt.Sprintf("\t\"%s\": {MD5Hash: \"%s\", FileSize: %d},\n", chkEntry.Path, chkEntry.MD5Hash, tmpFiles[chkEntry.Path].FileSize))
	}

	sort.Strings(files)
	sort.Strings(optFiles)

	for _, file := range files {
		w.WriteString(file)
	}

	if len(optFiles) > 0 {
		for _, file := range optFiles {
			wopt.WriteString(file)
		}
		wopt.WriteString("}\n")
	}

	w.WriteString("}\n")

	fmt.Printf("Finished in %0.2f seconds\n", time.Since(start).Seconds())

}

func concurrentChecksum(wg *sync.WaitGroup, chkChan chan ChecksumEntry, path string, relPath string, size int64) {
	md5, _ := MD5Generate(path)

	wg.Done()
	chkChan <- ChecksumEntry{Path: relPath, MD5Hash: md5}
}

func isFileExcluded(context string, path string, relPath string) bool {
	stringExcludes := []string{
		"AutoChannels.txt",
		"Darkhollow_Manual.pdf",
		"EverQuest Seeds Of Destruction Quick Start Guide.pdf",
		"gates_manual.pdf",
		"Help/",
		"LaunchPad",
		"LegendsOfNorrath",
		"maps/",
		"Omens_MANUAL.pdf",
		"OmensCredits.txt",
		"prophecy_of_ro.pdf",
		"Secrets_of_Faydwer_Manual.pdf",
		"The_Buried_Sea_Manual.pdf",
		"gates_manual.pdf",
		"prophecy_of_ro.pdf",
		"tss_manual.pdf",
		"eqhost.txt",
	}

	for _, exclude := range stringExcludes {
		if strings.Contains(relPath, exclude) {
			return true
		}
	}
	return false
}

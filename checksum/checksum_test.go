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
	start := time.Now()
	context := "ls"
	rootPath := fmt.Sprintf("C:/Program Files (x86)/Steam/steamapps/content/app_205710/depot_205711_%s", context)
	w, err := os.Create(fmt.Sprintf("%s_checksums.txt", context))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer w.Close()

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

	for {
		if len(chkChan) == 0 {
			break
		}
		chkEntry := <-chkChan
		if context == "ls" {
			rofHash := MD5Hash(ClientRoF2, chkEntry.Path)

			if rofHash != "" {
				continue
			}

			if rofHash != "" && rofHash != chkEntry.MD5Hash {
				fmt.Printf("MD5 hash mismatch for %s: %s != %s\n", chkEntry.Path, rofHash, chkEntry.MD5Hash)
				//t.Fatalf("MD5 hash mismatch for %s: %s != %s", chkEntry.Path, rofHash, chkEntry.MD5Hash)
			}
		}
		files = append(files, fmt.Sprintf("\t\"%s\": {MD5Hash: \"%s\", FileSize: %d},\n", chkEntry.Path, chkEntry.MD5Hash, tmpFiles[chkEntry.Path].FileSize))
	}

	sort.Strings(files)

	for _, file := range files {
		w.WriteString(file)
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
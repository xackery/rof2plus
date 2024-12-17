package checksum

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cespare/xxhash"
)

type ChecksumClient int

const (
	ClientRoF2 ChecksumClient = iota
	ClientLS
	ClientPatcher
)

var (
	mux              sync.RWMutex
	isClientLimited  bool
	patcherChecksums = make(map[string]*ChecksumEntry)
)

type ChecksumEntry struct {
	Path     string
	MD5Hash  string
	XXH3Hash string
	FileSize int64
}

func SetClientLimit(isLimited bool) {
	mux.Lock()
	defer mux.Unlock()
	isClientLimited = isLimited
}

// FileSize returns the size of the from checksum cache
func FileSize(context ChecksumClient, filename string) int64 {
	mux.RLock()
	defer mux.RUnlock()

	if isClientLimited {
		switch context {
		case ClientRoF2:
			entry, ok := rofChecksums[filename]
			if ok {
				return entry.FileSize
			}
		case ClientLS:
			entry, ok := lsChecksums[filename]
			if ok {
				return entry.FileSize
			}
		case ClientPatcher:
			entry, ok := patcherChecksums[filename]
			if ok {
				return entry.FileSize
			}
		}
	}

	entry, ok := patcherChecksums[filename]
	if ok {
		return entry.FileSize
	}

	entry, ok = lsChecksums[filename]
	if ok {
		return entry.FileSize
	}

	entry, ok = rofChecksums[filename]
	if ok {
		return entry.FileSize
	}

	return -1
}

// MD5Hash returns the MD5 hash of the from checksum cache
func MD5Hash(context ChecksumClient, filename string) string {
	mux.RLock()
	defer mux.RUnlock()

	switch context {
	case ClientRoF2:
		entry, ok := rofChecksums[filename]
		if ok {
			return entry.MD5Hash
		}
	case ClientLS:
		entry, ok := lsChecksums[filename]
		if ok {
			return entry.MD5Hash
		}
	case ClientPatcher:
		entry, ok := patcherChecksums[filename]
		if ok {
			return entry.MD5Hash
		}
	}
	return ""
}

// XXH3Hash returns the XXH3 hash of the from checksum cache
func XXH3Hash(context ChecksumClient, filename string) string {
	mux.RLock()
	defer mux.RUnlock()

	switch context {
	case ClientRoF2:
		entry, ok := rofChecksums[filename]
		if ok {
			return entry.XXH3Hash
		}
	case ClientLS:
		entry, ok := lsChecksums[filename]
		if ok {
			return entry.XXH3Hash
		}
	case ClientPatcher:
		entry, ok := patcherChecksums[filename]
		if ok {
			return entry.XXH3Hash
		}
	}
	return ""
}

func XXH3Generate(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", filePath, err)
	}
	defer file.Close()

	hasher := xxhash.New()

	buffer := make([]byte, 4096)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", err
		}
		if bytesRead == 0 {
			break
		}

		_, hashErr := hasher.Write(buffer[:bytesRead])
		if hashErr != nil {
			return "", hashErr
		}
	}

	return fmt.Sprintf("%016X", hasher.Sum64()), nil
}

func MD5Generate(path string) (value string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}
	value = fmt.Sprintf("%x", h.Sum(nil))
	return
}

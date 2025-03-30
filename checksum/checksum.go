// package checksum manages the checksum of various file states
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
	ClientRoF2Core
	ClientLS
	ClientPatcher
)

func (e *ChecksumClient) String() string {
	switch *e {
	case ClientRoF2:
		return "rof2"
	case ClientLS:
		return "ls"
	case ClientPatcher:
		return "patcher"
	}
	return "unknown"
}

var (
	mux              sync.RWMutex
	isClientLimited  bool
	excludedClients  []ChecksumClient
	patcherChecksums = make(map[string]*ChecksumEntry)
)

type ChecksumEntry struct {
	IsDeleted bool
	Path      string
	MD5Hash   string
	XXH3Hash  string
	FileSize  int64
}

func SetClientLimit(isLimited bool) {
	mux.Lock()
	defer mux.Unlock()
	isClientLimited = isLimited
}

// FileSize returns the size of the from checksum cache
func FileSize(client ChecksumClient, filename string) int64 {
	mux.RLock()
	defer mux.RUnlock()

	if isClientLimited {
		switch client {
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
		case ClientRoF2Core:
			entry, ok := rofCoreChecksums[filename]
			if ok {
				return entry.FileSize
			}
		case ClientPatcher:
			entry, ok := patcherChecksums[filename]
			if ok {
				return entry.FileSize
			}
		default:
			return -1
		}
	}

	var entry *ChecksumEntry
	var ok bool

	if !isClientExcluded(ClientPatcher) {
		entry, ok = patcherChecksums[filename]
		if ok {
			return entry.FileSize
		}
	}

	if !isClientExcluded(ClientLS) {
		entry, ok = lsChecksums[filename]
		if ok {
			return entry.FileSize
		}
	}

	if !isClientExcluded(ClientRoF2Core) {
		entry, ok = rofCoreChecksums[filename]
		if ok {
			return entry.FileSize
		}
	}

	if !isClientExcluded(ClientRoF2) {
		entry, ok = rofChecksums[filename]
		if ok {
			return entry.FileSize
		}
	}

	return -1
}

// MD5Hash returns the MD5 hash of the from checksum cache
func MD5Hash(client ChecksumClient, filename string) string {
	mux.RLock()
	defer mux.RUnlock()

	switch client {
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
func XXH3Hash(client ChecksumClient, filename string) string {
	mux.RLock()
	defer mux.RUnlock()

	switch client {
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

func ByClient(client ChecksumClient) (map[string]*ChecksumEntry, error) {
	mux.RLock()
	defer mux.RUnlock()

	if isClientLimited {
		switch client {
		case ClientRoF2:
			return rofChecksums, nil
		case ClientRoF2Core:
			return rofCoreChecksums, nil
		case ClientLS:
			return lsChecksums, nil
		case ClientPatcher:
			return patcherChecksums, nil
		default:
			return nil, fmt.Errorf("unknown client: %d", client)
		}
	}

	checksums := make(map[string]*ChecksumEntry)
	if !isClientExcluded(ClientPatcher) {
		for k, v := range patcherChecksums {
			if k == "Resources/BaseData.txt" {
				fmt.Println("test")
			}
			checksums[k] = v
		}
	}

	if !isClientExcluded(ClientLS) {
		for k, v := range lsChecksums {
			checksums[k] = v
		}
	}

	if !isClientExcluded(ClientRoF2) {
		for k, v := range rofChecksums {
			checksums[k] = v
		}
	}

	if !isClientExcluded(ClientRoF2Core) {
		for k, v := range rofCoreChecksums {
			checksums[k] = v
		}
	}

	return checksums, nil
}

func SetExcludedClients(clients ...ChecksumClient) {
	mux.Lock()
	defer mux.Unlock()
	excludedClients = append(excludedClients, clients...)
}

func isClientExcluded(client ChecksumClient) bool {
	for _, c := range excludedClients {
		if c == client {
			return true
		}
	}
	return false
}

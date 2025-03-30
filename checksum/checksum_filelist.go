package checksum

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FileList represents a file_list.yml file downloaded from server
type FileList struct {
	Version        string      `yaml:"version"`
	DownloadPrefix string      `yaml:"downloadprefix"`
	Deletes        []FileEntry `yaml:"deletes"`
	Downloads      []FileEntry `yaml:"downloads"`
	Unpacks        []FileEntry `yaml:"unpacks"`
}

// FileEntry is an entry inside FileList
type FileEntry struct {
	Name string `yaml:"name"`
	Md5  string `yaml:"md5"`
	Date string `yaml:"date"`
	Zip  string `yaml:"zip"`
	Size int    `yaml:"size"`
}

// FetchPatcherFilelist fetches the filelist from the patcher server
func FetchPatcherFilelist(baseURL string) (*FileList, error) {
	fileList := &FileList{}
	data, err := os.ReadFile("filelist_rof.yml")
	if err == nil {
		err = yaml.Unmarshal(data, fileList)
		if err != nil {
			return nil, fmt.Errorf("decode filelist: %w", err)
		}

		err = SetPatcherFilelist(fileList)
		if err != nil {
			return nil, fmt.Errorf("set patcher filelist: %w", err)
		}
		return fileList, nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s/filelist_rof.yml", baseURL)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download %s responded HTTP status code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	err = yaml.Unmarshal(data, fileList)
	if err != nil {
		return nil, fmt.Errorf("decode filelist: %w", err)
	}

	err = SetPatcherFilelist(fileList)
	if err != nil {
		return nil, fmt.Errorf("set patcher filelist: %w", err)
	}

	//slog.Print("patch version is", fileList.Version, "and we are version", c.cfg.ClientVersion)

	patcherChecksums = map[string]*ChecksumEntry{}

	for _, entry := range fileList.Downloads {
		patcherChecksums[entry.Name] = &ChecksumEntry{
			Path:     entry.Name,
			MD5Hash:  entry.Md5,
			FileSize: int64(entry.Size),
		}
	}
	return fileList, nil
}

// SetPatcherFilelist sets the patcher filelist from the given data
func SetPatcherFilelist(fileList *FileList) error {

	patcherChecksums = map[string]*ChecksumEntry{}

	for _, entry := range fileList.Downloads {
		patcherChecksums[entry.Name] = &ChecksumEntry{
			Path:     entry.Name,
			MD5Hash:  entry.Md5,
			FileSize: int64(entry.Size),
		}
	}

	for _, entry := range fileList.Deletes {
		patcherChecksums[entry.Name] = &ChecksumEntry{
			IsDeleted: true,
			Path:      entry.Name,
			MD5Hash:   "DELETE",
			FileSize:  0,
		}
	}

	return nil
}

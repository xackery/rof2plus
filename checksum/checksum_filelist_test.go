package checksum

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestChecksumFilelist(t *testing.T) {
	data, err := os.ReadFile("../bin/filelist_rof.yml")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	fileList := &FileList{}

	err = yaml.Unmarshal(data, fileList)
	if err != nil {
		t.Fatalf("failed to unmarshal filelist: %v", err)
	}

	err = SetPatcherFilelist(fileList)
	if err != nil {
		t.Fatalf("failed to set patcher filelist: %v", err)
	}

}

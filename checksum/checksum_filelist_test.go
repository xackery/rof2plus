package checksum

import (
	"os"
	"testing"
)

func TestChecksumFilelist(t *testing.T) {
	data, err := os.ReadFile("../bin/filelist_rof.yml")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	err = SetPatcherFilelist(data)
	if err != nil {
		t.Fatalf("failed to set patcher filelist: %v", err)
	}

}

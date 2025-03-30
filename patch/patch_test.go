package patch

import (
	"testing"

	"github.com/xackery/rof2plus/checksum"
)

func TestPatch(t *testing.T) {
	testDir := "../bin/thj"

	//url := "https://github.com/jamfesteq/eqemupatcher/releases/download/1.0.6.34"
	//url := "https://github.com/carolus21rex/eqemupatcher/releases/download/1.0.6.34/"
	url := "https://github.com/The-Heroes-Journey-EQEMU/eqemupatcher/releases/download/1.0.6.453/"
	fileList, err := checksum.FetchPatcherFilelist(url)
	if err != nil {
		t.Fatalf("Failed to fetch filelist: %v", err)
	}

	err = Download(fileList, testDir)
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}

}
